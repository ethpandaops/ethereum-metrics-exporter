package docker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type ContainerPortMonitor struct {
	ContainerID  string
	ContainerIP  string
	Ports        []PortInfo
	Rules        map[string]*RuleSet
	LastCounters map[string]CounterStats
}

type PortMonitor interface {
	Start(ctx context.Context) error
	Stop() error
	AddContainer(ctx context.Context, containerID string) error
	RemoveContainer(containerID string) error
	CollectMetrics() (map[string]map[string]CounterStats, error)
}

type portMonitor struct {
	dockerClient     client.APIClient
	nftablesManager  NFTablesManager
	namespaceManager NamespaceManager
	portDiscovery    PortDiscovery
	config           *PortBandwidthConfig
	activeMonitors   map[string]*ContainerPortMonitor
	mu               sync.RWMutex
	logger           logrus.FieldLogger
	done             chan struct{}
	wg               sync.WaitGroup
}

var _ PortMonitor = (*portMonitor)(nil)

func NewPortMonitor(dockerClient client.APIClient, config *PortBandwidthConfig, logger logrus.FieldLogger) (PortMonitor, error) {
	nftablesManager, err := NewNFTablesManager(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create nftables manager: %w", err)
	}

	namespaceManager := NewNamespaceManager(dockerClient, logger)
	portDiscovery := NewPortDiscovery(dockerClient, logger)

	return &portMonitor{
		dockerClient:     dockerClient,
		nftablesManager:  nftablesManager,
		namespaceManager: namespaceManager,
		portDiscovery:    portDiscovery,
		config:           config,
		activeMonitors:   make(map[string]*ContainerPortMonitor, 20),
		logger:           logger.WithField("component", "port_monitor"),
		done:             make(chan struct{}),
	}, nil
}

func (pm *portMonitor) Start(ctx context.Context) error {
	pm.logger.Info("Starting port monitor")

	if err := pm.nftablesManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start nftables manager: %w", err)
	}

	pm.wg.Add(1)

	go pm.monitoringLoop(ctx)

	pm.logger.Info("Port monitor started successfully")

	return nil
}

func (pm *portMonitor) Stop() error {
	pm.logger.Info("Stopping port monitor")

	close(pm.done)
	pm.wg.Wait()

	if err := pm.nftablesManager.Stop(); err != nil {
		pm.logger.WithError(err).Error("Failed to stop nftables manager")
		return fmt.Errorf("failed to stop nftables manager: %w", err)
	}

	pm.logger.Info("Port monitor stopped successfully")

	return nil
}

func (pm *portMonitor) monitoringLoop(ctx context.Context) {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			pm.logger.Debug("Port monitoring loop stopping due to context cancellation")
			return
		case <-pm.done:
			pm.logger.Debug("Port monitoring loop stopping due to shutdown signal")
			return
		case <-ticker.C:
			if _, err := pm.CollectMetrics(); err != nil {
				pm.logger.WithError(err).Warn("Failed to collect port bandwidth metrics")
			}
		}
	}
}

func (pm *portMonitor) AddContainer(ctx context.Context, containerID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.activeMonitors[containerID]; exists {
		return nil // Already monitoring
	}

	containerIP, err := pm.getContainerIP(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container IP: %w", err)
	}

	if containerIP == "" {
		pm.logger.WithField("container_id", containerID).Debug("Container has no IP address, skipping port monitoring")
		return nil
	}

	ports, err := pm.portDiscovery.DiscoverContainerPorts(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to discover container ports: %w", err)
	}

	if len(ports) == 0 {
		pm.logger.WithField("container_id", containerID).Debug("No ports discovered for container")
		return nil
	}

	filteredPorts := pm.filterPorts(ports)
	if len(filteredPorts) == 0 {
		pm.logger.WithFields(logrus.Fields{
			"container_id":     containerID,
			"discovered_ports": len(ports),
			"config": map[string]interface{}{
				"monitor_all_ports": pm.config.MonitorAllPorts,
				"specific_ports":    pm.config.SpecificPorts,
				"protocols":         pm.config.Protocols,
			},
		}).Info("No ports match monitoring criteria")

		return nil
	}

	pm.logger.WithFields(logrus.Fields{
		"container_id":     containerID,
		"discovered_ports": len(ports),
		"filtered_ports":   len(filteredPorts),
	}).Debug("Port filtering completed")

	monitor := &ContainerPortMonitor{
		ContainerID:  containerID,
		ContainerIP:  containerIP,
		Ports:        filteredPorts,
		Rules:        make(map[string]*RuleSet, len(filteredPorts)),
		LastCounters: make(map[string]CounterStats, len(filteredPorts)*2),
	}

	for _, port := range filteredPorts {
		err := pm.namespaceManager.ExecuteInContainerNamespace(ctx, containerID, func() error {
			ruleSet, err := pm.nftablesManager.CreatePortRules(containerIP, port.Port, port.Protocol)
			if err != nil {
				return fmt.Errorf("failed to create rules for port %d/%s: %w", port.Port, port.Protocol, err)
			}

			ruleKey := fmt.Sprintf("%d:%s", port.Port, port.Protocol)
			monitor.Rules[ruleKey] = ruleSet

			return nil
		})
		if err != nil {
			pm.logger.WithError(err).WithFields(logrus.Fields{
				"container_id": containerID,
				"port":         port.Port,
				"protocol":     port.Protocol,
			}).Warn("Failed to create nftables rules for port")

			continue
		}
	}

	if len(monitor.Rules) > 0 {
		pm.activeMonitors[containerID] = monitor

		// Log detailed information about monitored ports
		monitoredPorts := make([]map[string]interface{}, 0, len(monitor.Rules))
		for ruleKey := range monitor.Rules {
			parts := strings.Split(ruleKey, ":")
			if len(parts) == 2 {
				monitoredPorts = append(monitoredPorts, map[string]interface{}{
					"port":     parts[0],
					"protocol": parts[1],
				})
			}
		}

		pm.logger.WithFields(logrus.Fields{
			"container_id":    containerID,
			"container_ip":    containerIP,
			"port_count":      len(monitor.Rules),
			"monitored_ports": monitoredPorts,
			"total_ports":     len(filteredPorts),
			"skipped_ports":   len(filteredPorts) - len(monitor.Rules),
		}).Info("Added container to port monitoring")
	} else {
		pm.logger.WithFields(logrus.Fields{
			"container_id":     containerID,
			"container_ip":     containerIP,
			"discovered_ports": len(filteredPorts),
		}).Warn("No nftables rules were successfully created for container")
	}

	return nil
}

func (pm *portMonitor) RemoveContainer(containerID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	monitor, exists := pm.activeMonitors[containerID]
	if !exists {
		return nil // Not monitoring
	}

	for ruleKey, ruleSet := range monitor.Rules {
		err := pm.nftablesManager.RemovePortRules(ruleSet.ContainerIP, ruleSet.Port, ruleSet.Protocol)
		if err != nil {
			pm.logger.WithError(err).WithFields(logrus.Fields{
				"container_id": containerID,
				"rule_key":     ruleKey,
			}).Warn("Failed to remove nftables rules")
		}
	}

	delete(pm.activeMonitors, containerID)
	pm.logger.WithField("container_id", containerID).Info("Removed container from port monitoring")

	return nil
}

func (pm *portMonitor) CollectMetrics() (map[string]map[string]CounterStats, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	allStats, err := pm.nftablesManager.ReadCounters()
	if err != nil {
		return nil, fmt.Errorf("failed to read nftables counters: %w", err)
	}

	result := make(map[string]map[string]CounterStats, len(pm.activeMonitors))

	for containerID, monitor := range pm.activeMonitors {
		containerStats := make(map[string]CounterStats, len(monitor.Rules)*2)

		for ruleKey := range monitor.Rules {
			ingressKey := fmt.Sprintf("%s:%s:%s:ingress", monitor.ContainerIP, ruleKey, "ingress")
			if stats, exists := allStats[ingressKey]; exists {
				containerStats[ruleKey+":ingress"] = stats
				monitor.LastCounters[ruleKey+":ingress"] = stats
			}

			egressKey := fmt.Sprintf("%s:%s:%s:egress", monitor.ContainerIP, ruleKey, "egress")
			if stats, exists := allStats[egressKey]; exists {
				containerStats[ruleKey+":egress"] = stats
				monitor.LastCounters[ruleKey+":egress"] = stats
			}
		}

		if len(containerStats) > 0 {
			result[containerID] = containerStats
		}
	}

	return result, nil
}

func (pm *portMonitor) getContainerIP(ctx context.Context, containerID string) (string, error) {
	container, err := pm.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	if container.NetworkSettings == nil {
		return "", nil
	}

	if container.NetworkSettings.IPAddress != "" {
		return container.NetworkSettings.IPAddress, nil
	}

	for _, network := range container.NetworkSettings.Networks {
		if network.IPAddress != "" {
			return network.IPAddress, nil
		}
	}

	return "", nil
}

func (pm *portMonitor) filterPorts(ports []PortInfo) []PortInfo {
	if !pm.config.MonitorAllPorts && len(pm.config.SpecificPorts) == 0 {
		return []PortInfo{}
	}

	var filtered []PortInfo

	protocolMap := make(map[string]bool, len(pm.config.Protocols))
	for _, protocol := range pm.config.Protocols {
		protocolMap[protocol] = true
	}

	specificPortMap := make(map[int]bool, len(pm.config.SpecificPorts))
	for _, port := range pm.config.SpecificPorts {
		specificPortMap[port] = true
	}

	for _, port := range ports {
		if len(protocolMap) > 0 && !protocolMap[port.Protocol] {
			continue
		}

		if pm.config.MonitorAllPorts {
			filtered = append(filtered, port)
		} else if specificPortMap[port.Port] {
			filtered = append(filtered, port)
		}
	}

	return filtered
}
