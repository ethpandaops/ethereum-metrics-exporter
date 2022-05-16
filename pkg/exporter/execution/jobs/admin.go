package jobs

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api/types"
	"github.com/sirupsen/logrus"
)

// Admin exposes metrics defined by the admin module.
type Admin struct {
	MetricExporter
	client                   *ethclient.Client
	api                      api.ExecutionClient
	log                      logrus.FieldLogger
	NodeInfo                 prometheus.GaugeVec
	Port                     prometheus.GaugeVec
	Peers                    prometheus.Gauge
	TotalDifficultyTrillions prometheus.Gauge
	TotalDifficulty          prometheus.Gauge
}

const (
	NameAdmin = "admin"
)

func (t *Admin) Name() string {
	return NameAdmin
}

func (t *Admin) RequiredModules() []string {
	return []string{"admin"}
}

// NewAdmin returns a new Admin instance.
func NewAdmin(client *ethclient.Client, internalApi api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string) Admin {
	namespace = namespace + "_admin"
	constLabels["module"] = NameAdmin

	return Admin{
		client: client,
		api:    internalApi,
		log:    log.WithField("module", NameAdmin),
		NodeInfo: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "node_info",
				Help:        "Node info.",
				ConstLabels: constLabels,
			},
			[]string{
				"ip",
				"listenAddr",
				"name",
				"discovery_port",
				"listener_port",
				"network",
			},
		),
		Port: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "node_port",
				Help:        "The ports for the node.",
				ConstLabels: constLabels,
			},
			[]string{
				"name",
				"port_name",
			},
		),
		Peers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "peers",
				Help:        "The number of peers connected with the node.",
				ConstLabels: constLabels,
			},
		),
		TotalDifficulty: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "total_difficulty",
				Help:        "The total difficulty of the chain in trillions.",
				ConstLabels: constLabels,
			},
		),
		TotalDifficultyTrillions: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "total_difficulty_trillions",
				Help:        "The total difficulty of the chain.",
				ConstLabels: constLabels,
			},
		),
	}
}

func (a *Admin) Start(ctx context.Context) {
	a.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			a.tick(ctx)
		}
	}
}

func (a *Admin) tick(ctx context.Context) {
	nodeInfo, err := a.api.AdminNodeInfo(ctx)
	if err != nil {
		a.log.WithError(err).Error("Failed to get node info")
	} else {
		a.ObserveNodeInfo(nodeInfo)
	}

	peers, err := a.api.AdminPeers(ctx)
	if err != nil {
		a.log.WithError(err).Error("Failed to get peers")
	} else {
		a.ObservePeers(len(peers))
	}
}

func (a *Admin) ObserveNodeInfo(nodeInfo *types.NodeInfo) {
	// Info
	a.NodeInfo.WithLabelValues(nodeInfo.IP,
		nodeInfo.ListenAddr,
		nodeInfo.Name,
		fmt.Sprint(nodeInfo.Ports.Discovery),
		fmt.Sprint(nodeInfo.Ports.Listener),
		fmt.Sprint(nodeInfo.Protocols.Eth.NetworkID),
	).Set(1)

	// Ports
	a.Port.WithLabelValues("discovery", "discovery").Set(float64(nodeInfo.Ports.Discovery))
	a.Port.WithLabelValues("listener", "listener").Set(float64(nodeInfo.Ports.Listener))

	// Total Difficulty
	a.TotalDifficulty.Set(float64(nodeInfo.Difficulty().Uint64()))
	// Since we can't represent a big.Int as a float64, and the TD on mainnet is beyond float64, we'll divide the number by a trillion
	trillion := big.NewInt(1e12)
	divided := new(big.Int).Quo(nodeInfo.Difficulty(), trillion)
	a.TotalDifficultyTrillions.Set(float64(divided.Uint64()))
}

func (a *Admin) ObservePeers(peers int) {
	a.Peers.Set(float64(peers))
}
