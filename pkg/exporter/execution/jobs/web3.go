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

// Web3 exposes metrics defined by the Web3 module.
type Web3 struct {
	client        *ethclient.Client
	api           api.ExecutionClient
	ethRPCClient  *ethrpc.EthRPC
	log           logrus.FieldLogger
	ClientVersion prometheus.GaugeVec
}

const (
	NameWeb3 = "web3"
)

func (w *Web3) Name() string {
	return NameWeb3
}

func (w *Web3) RequiredModules() []string {
	return []string{"web3"}
}

// NewWeb3 returns a new Web3 instance.
func NewWeb3(client *ethclient.Client, internalAPI api.ExecutionClient, ethRPCClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) Web3 {
	namespace += "_web3"

	constLabels["module"] = NameWeb3

	return Web3{
		client:       client,
		api:          internalAPI,
		ethRPCClient: ethRPCClient,
		log:          log.WithField("module", NameWeb3),
		ClientVersion: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "client_version",
				Help:        "Client version.",
				ConstLabels: constLabels,
			},
			[]string{
				"version",
			},
		),
	}
}

func (w *Web3) Start(ctx context.Context) {
	w.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			w.tick(ctx)
		}
	}
}

//nolint:unparam // context will be used in the future
func (w *Web3) tick(ctx context.Context) {
	clientVersion, err := w.ethRPCClient.Web3ClientVersion()
	if err != nil {
		w.log.WithError(err).Error("Failed to get node info")
	} else {
		w.ClientVersion.WithLabelValues(clientVersion).Set(1)
	}
}
