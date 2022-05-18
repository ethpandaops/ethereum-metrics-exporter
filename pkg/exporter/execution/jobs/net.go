package jobs

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/sirupsen/logrus"
)

// Net exposes metrics defined by the net module.
type Net struct {
	MetricExporter
	client       *ethclient.Client
	api          api.ExecutionClient
	ethrpcClient *ethrpc.EthRPC
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
func NewNet(client *ethclient.Client, internalApi api.ExecutionClient, ethRpcClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) Net {
	namespace = namespace + "_net"
	constLabels["module"] = NameWeb3

	return Net{
		client:       client,
		api:          internalApi,
		ethrpcClient: ethRpcClient,
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

func (n *Net) tick(ctx context.Context) {
	count, err := n.ethrpcClient.NetPeerCount()
	if err != nil {
		n.log.WithError(err).Error("Failed to get peer count")
	} else {
		n.PeerCount.Set(float64(count))
	}

}
