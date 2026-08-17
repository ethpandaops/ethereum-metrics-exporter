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
	client    *ethclient.Client
	api       api.ExecutionClient
	log       logrus.FieldLogger
	PeerCount prometheus.Gauge
}

const (
	NameNet = "net"
)

func (n *Net) Name() string {
	return NameNet
}

func (n *Net) RequiredModules() []string {
	return []string{NameNet}
}

// NewNet returns a new Net instance.
func NewNet(client *ethclient.Client, internalAPI api.ExecutionClient, ethRPCClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) Net {
	namespace += "_net"

	constLabels["module"] = NameNet

	return Net{
		client: client,
		api:    internalAPI,
		log:    log.WithField("module", NameNet),
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
	// Use the go-ethereum client, which omits the params member for
	// zero-argument calls; some clients (nimbus-eth1) reject "params": null.
	count, err := n.client.PeerCount(ctx)
	if err != nil {
		n.log.WithError(err).Error("Failed to get peer count")
	} else {
		n.PeerCount.Set(float64(count))
	}
}
