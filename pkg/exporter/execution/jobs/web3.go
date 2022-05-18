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

// Web3 exposes metrics defined by the Web3 module.
type Web3 struct {
	MetricExporter
	client        *ethclient.Client
	api           api.ExecutionClient
	ethrpcClient  *ethrpc.EthRPC
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
func NewWeb3(client *ethclient.Client, internalApi api.ExecutionClient, ethRpcClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) Web3 {
	namespace = namespace + "_web3"
	constLabels["module"] = NameWeb3

	return Web3{
		client:       client,
		api:          internalApi,
		ethrpcClient: ethRpcClient,
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

func (a *Web3) Start(ctx context.Context) {
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

func (a *Web3) tick(ctx context.Context) {
	clientVersion, err := a.ethrpcClient.Web3ClientVersion()
	if err != nil {
		a.log.WithError(err).Error("Failed to get node info")
	} else {
		a.ClientVersion.WithLabelValues(clientVersion).Set(1)
	}

}
