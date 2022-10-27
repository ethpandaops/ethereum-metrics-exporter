package jobs

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Net exposes metrics defined by the net module.
type Net struct {
	client       *ethclient.Client
	api          api.ExecutionClient
	ethRPCClient *ethrpc.EthRPC
	log          logrus.FieldLogger
	PeerCount    prometheus.Gauge
}

const (
	NameNet = "net"
)

func (n *Net) Name() string {
	return NameNet
}

func (n *Net) RequiredModules() []string {
	return []string{"net"}
}

// NewNet returns a new Net instance.
func NewNet(client *ethclient.Client, internalAPI api.ExecutionClient, ethRPCClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) Net {
	namespace += "_net"

	constLabels["module"] = NameWeb3

	return Net{
		client:       client,
		api:          internalAPI,
		ethRPCClient: ethRPCClient,
		log:          log.WithField("module", NameNet),
		PeerCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "peer_count",
				Help:        "The amount of peers connected to the node.",
				ConstLabels: constLabels,
			},
		),
	}
}

func (n *Net) Start(ctx context.Context) {
	n.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			n.tick(ctx)
		}
	}
}

//nolint:unparam // context will be used in the future
func (n *Net) tick(ctx context.Context) {
	count, err := n.ethRPCClient.NetPeerCount()
	if err != nil {
		n.log.WithError(err).Error("Failed to get peer count")
	} else {
		n.PeerCount.Set(float64(count))
	}
}
