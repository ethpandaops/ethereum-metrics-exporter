package docker

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/knftables"
)

const (
	dockerPortMonitorTable = "docker_port_monitor"
	ingressChainName       = "ingress"
	egressChainName        = "egress"
)

type NFTablesManager interface {
	Start(ctx context.Context) error
	Stop() error
	CreatePortRules(containerIP string, port int, protocol string) (*RuleSet, error)
	RemovePortRules(containerIP string, port int, protocol string) error
	ReadCounters() (map[string]CounterStats, error)
}

type RuleSet struct {
	IngressRuleID string
	EgressRuleID  string
	Port          int
	Protocol      string
	ContainerIP   string
	CreatedAt     time.Time
}

type CounterStats struct {
	Bytes     uint64
	Packets   uint64
	Timestamp time.Time
}

type nftablesManager struct {
	nft    knftables.Interface
	table  string
	rules  map[string]*RuleSet
	mu     sync.RWMutex
	pid    int
	logger logrus.FieldLogger
}

var _ NFTablesManager = (*nftablesManager)(nil)

func NewNFTablesManager(logger logrus.FieldLogger) (NFTablesManager, error) {
	nft, err := knftables.New(knftables.InetFamily, dockerPortMonitorTable)
	if err != nil {
		return nil, fmt.Errorf("failed to create knftables interface: %w", err)
	}

	manager := &nftablesManager{
		nft:    nft,
		table:  dockerPortMonitorTable,
		rules:  make(map[string]*RuleSet, 100),
		pid:    os.Getpid(),
		logger: logger.WithField("component", "nftables_manager"),
	}

	return manager, nil
}

func (n *nftablesManager) checkNFTablesAvailable() error {
	// Try a simple nftables command to check if it's available
	ctx := context.Background()
	tx := n.nft.NewTransaction()

	// Try to list tables - this is a safe read-only operation
	if err := n.nft.Run(ctx, tx); err != nil {
		return fmt.Errorf("nftables is not available or accessible: %w", err)
	}

	return nil
}

func (n *nftablesManager) Start(ctx context.Context) error {
	n.logger.Info("Starting nftables manager")

	// Check if nftables is available
	if err := n.checkNFTablesAvailable(); err != nil {
		n.logger.WithError(err).Error("NFTables is not available in this environment")
		return fmt.Errorf("nftables not available: %w", err)
	}

	if err := n.StartupCleanup(ctx); err != nil {
		n.logger.WithError(err).Warn("Failed to clean up stale rules during startup")
	}

	if err := n.initializeTable(ctx); err != nil {
		n.logger.WithError(err).Error("Failed to initialize nftables table - ensure container has --privileged flag and kernel modules are loaded")
		return fmt.Errorf("failed to initialize nftables table: %w", err)
	}

	if err := n.createPIDMarker(ctx); err != nil {
		n.logger.WithError(err).Warn("Failed to create PID marker")
	}

	n.logger.Info("NFTables manager started successfully")

	return nil
}

func (n *nftablesManager) Stop() error {
	n.logger.Info("Stopping nftables manager and cleaning up rules")

	n.mu.Lock()
	defer n.mu.Unlock()

	ctx := context.Background()
	tx := n.nft.NewTransaction()
	tx.Delete(&knftables.Table{})

	if err := n.nft.Run(ctx, tx); err != nil {
		n.logger.WithError(err).Error("Failed to delete nftables table during shutdown")
		return fmt.Errorf("failed to cleanup nftables table: %w", err)
	}

	n.rules = make(map[string]*RuleSet)
	n.logger.Info("NFTables manager stopped and rules cleaned up")

	return nil
}

func (n *nftablesManager) initializeTable(ctx context.Context) error {
	tx := n.nft.NewTransaction()

	// Helper variables for constants
	filterType := knftables.FilterType
	inputHook := knftables.InputHook
	outputHook := knftables.OutputHook
	filterPriority := knftables.FilterPriority

	// Create ingress chain
	tx.Add(&knftables.Chain{
		Name:     ingressChainName,
		Type:     &filterType,
		Hook:     &inputHook,
		Priority: &filterPriority,
	})

	// Create egress chain
	tx.Add(&knftables.Chain{
		Name:     egressChainName,
		Type:     &filterType,
		Hook:     &outputHook,
		Priority: &filterPriority,
	})

	return n.nft.Run(ctx, tx)
}

func (n *nftablesManager) CreatePortRules(containerIP string, port int, protocol string) (*RuleSet, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	ruleKey := fmt.Sprintf("%s:%d:%s", containerIP, port, protocol)
	if _, exists := n.rules[ruleKey]; exists {
		return nil, fmt.Errorf("rules already exist for %s", ruleKey)
	}

	containerIPNet := net.ParseIP(containerIP)
	if containerIPNet == nil {
		return nil, fmt.Errorf("invalid container IP: %s", containerIP)
	}

	if protocol != "tcp" && protocol != "udp" {
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	ctx := context.Background()
	tx := n.nft.NewTransaction()

	ingressRuleID := fmt.Sprintf("ingress-%s-%d-%s", strings.ReplaceAll(containerIP, ".", "-"), port, protocol)
	egressRuleID := fmt.Sprintf("egress-%s-%d-%s", strings.ReplaceAll(containerIP, ".", "-"), port, protocol)

	// Create ingress rule (incoming traffic to container)
	tx.Add(&knftables.Rule{
		Chain: ingressChainName,
		Rule: knftables.Concat(
			"ip", "daddr", containerIP,
			protocol, "dport", strconv.Itoa(port),
			"counter",
		),
		Comment: knftables.PtrTo(fmt.Sprintf("Docker container %s port %d/%s ingress", containerIP, port, protocol)),
	})

	// Create egress rule (outgoing traffic from container)
	tx.Add(&knftables.Rule{
		Chain: egressChainName,
		Rule: knftables.Concat(
			"ip", "saddr", containerIP,
			protocol, "sport", strconv.Itoa(port),
			"counter",
		),
		Comment: knftables.PtrTo(fmt.Sprintf("Docker container %s port %d/%s egress", containerIP, port, protocol)),
	})

	if err := n.nft.Run(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to add nftables rules for %s: %w", ruleKey, err)
	}

	ruleSet := &RuleSet{
		IngressRuleID: ingressRuleID,
		EgressRuleID:  egressRuleID,
		Port:          port,
		Protocol:      protocol,
		ContainerIP:   containerIP,
		CreatedAt:     time.Now(),
	}

	n.rules[ruleKey] = ruleSet

	n.logger.WithFields(logrus.Fields{
		"container_ip": containerIP,
		"port":         port,
		"protocol":     protocol,
	}).Debug("Created nftables rules for container port")

	return ruleSet, nil
}

func (n *nftablesManager) RemovePortRules(containerIP string, port int, protocol string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	ruleKey := fmt.Sprintf("%s:%d:%s", containerIP, port, protocol)
	_, exists := n.rules[ruleKey]

	if !exists {
		return fmt.Errorf("rules not found for %s", ruleKey)
	}

	ctx := context.Background()
	tx := n.nft.NewTransaction()

	// Delete ingress rule
	tx.Delete(&knftables.Rule{
		Chain: ingressChainName,
		Rule: knftables.Concat(
			"ip", "daddr", containerIP,
			protocol, "dport", strconv.Itoa(port),
			"counter",
		),
	})

	// Delete egress rule
	tx.Delete(&knftables.Rule{
		Chain: egressChainName,
		Rule: knftables.Concat(
			"ip", "saddr", containerIP,
			protocol, "sport", strconv.Itoa(port),
			"counter",
		),
	})

	if err := n.nft.Run(ctx, tx); err != nil {
		return fmt.Errorf("failed to remove nftables rules for %s: %w", ruleKey, err)
	}

	delete(n.rules, ruleKey)

	n.logger.WithFields(logrus.Fields{
		"container_ip": containerIP,
		"port":         port,
		"protocol":     protocol,
	}).Debug("Removed nftables rules for container port")

	return nil
}

func (n *nftablesManager) ReadCounters() (map[string]CounterStats, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Note: knftables doesn't provide direct counter reading API like google/nftables
	// This would require parsing nft list command output or using a different approach
	// For now, return empty stats with a warning
	n.logger.Warn("Counter reading not yet implemented with knftables library")

	stats := make(map[string]CounterStats, len(n.rules)*2)
	timestamp := time.Now()

	// Return zero counters for now - this would need to be implemented
	// by parsing nft list ruleset output or using alternative methods
	for ruleKey := range n.rules {
		stats[fmt.Sprintf("%s:ingress", ruleKey)] = CounterStats{
			Bytes:     0,
			Packets:   0,
			Timestamp: timestamp,
		}
		stats[fmt.Sprintf("%s:egress", ruleKey)] = CounterStats{
			Bytes:     0,
			Packets:   0,
			Timestamp: timestamp,
		}
	}

	return stats, nil
}

func (n *nftablesManager) StartupCleanup(ctx context.Context) error {
	n.logger.Debug("Performing startup cleanup of stale nftables rules")

	// Delete the entire table if it exists - knftables will recreate it
	tx := n.nft.NewTransaction()
	tx.Delete(&knftables.Table{})

	// Ignore errors here as the table might not exist
	_ = n.nft.Run(ctx, tx)

	return nil
}

func (n *nftablesManager) createPIDMarker(ctx context.Context) error {
	comment := fmt.Sprintf("docker_exporter_pid_%d", n.pid)

	tx := n.nft.NewTransaction()
	tx.Add(&knftables.Rule{
		Chain: ingressChainName,
		Rule: knftables.Concat(
			"meta", "mark", "==", strconv.Itoa(n.pid),
		),
		Comment: knftables.PtrTo(comment),
	})

	return n.nft.Run(ctx, tx)
}
