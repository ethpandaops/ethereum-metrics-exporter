package jobs

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/sirupsen/logrus"
)

type TXPool struct {
	MetricExporter
	client       *ethclient.Client
	api          api.ExecutionClient
	log          logrus.FieldLogger
	Transactions prometheus.GaugeVec
}

const (
	NameTxPool = "txpool"
)

func (t *TXPool) Name() string {
	return NameTxPool
}

func (t *TXPool) RequiredModules() []string {
	return []string{"txpool"}
}

func NewTXPool(client *ethclient.Client, internalApi api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string) TXPool {
	constLabels["module"] = NameTxPool
	namespace = namespace + "_txpool"
	return TXPool{
		client: client,
		api:    internalApi,
		log:    log.WithField("module", NameGeneral),
		Transactions: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "transactions",
				Help:        "How many transactions are in the txpool.",
				ConstLabels: constLabels,
			},
			[]string{
				"status",
			},
		),
	}
}

func (t *TXPool) Start(ctx context.Context) {
	t.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			t.tick(ctx)
		}
	}
}

func (t *TXPool) tick(ctx context.Context) {
	if err := t.GetStatus(ctx); err != nil {
		t.log.Errorf("Failed to get txpool status: %s", err)
	}
}

func (t *TXPool) GetStatus(ctx context.Context) error {
	status, err := t.api.TXPoolStatus(ctx)
	if err != nil {
		return err
	}

	t.Transactions.WithLabelValues("pending").Set(float64(status.Pending))
	t.Transactions.WithLabelValues("queued").Set(float64(status.Queued))

	return nil
}
