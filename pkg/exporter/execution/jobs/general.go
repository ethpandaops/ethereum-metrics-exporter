package jobs

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/sirupsen/logrus"
)

// GeneralMetrics exposes metrics that otherwise don't fit in to a specific module.
type GeneralMetrics struct {
	MetricExporter
	client                *ethclient.Client
	api                   api.ExecutionClient
	log                   logrus.FieldLogger
	MostRecentBlockNumber prometheus.Gauge
	GasPrice              prometheus.Gauge
	NetworkID             prometheus.Gauge
	ChainID               prometheus.Gauge
	GasUsed               prometheus.Gauge
	GasLimit              prometheus.Gauge
	BaseFeePerGas         prometheus.Gauge
	BlockSize             prometheus.Gauge
	TransactionCount      prometheus.Counter
}

const (
	NameGeneral = "general"
)

func (g *GeneralMetrics) Name() string {
	return NameGeneral
}

func (g *GeneralMetrics) RequiredModules() []string {
	return []string{"eth", "net"}
}

// NewGeneralMetrics returns a new General metrics instance.
func NewGeneralMetrics(client *ethclient.Client, internalApi api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string) GeneralMetrics {
	constLabels["module"] = NameGeneral
	return GeneralMetrics{
		client: client,
		api:    internalApi,
		log:    log.WithField("module", NameGeneral),
		MostRecentBlockNumber: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "most_recent_block_number",
				Help:        "The most recent block number.",
				ConstLabels: constLabels,
			},
		),
		GasPrice: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "gas_price_gwei",
				Help:        "The current gas price in gwei.",
				ConstLabels: constLabels,
			},
		),
		NetworkID: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "network_id",
				Help:        "The network id of the node.",
				ConstLabels: constLabels,
			},
		),
		ChainID: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "chain_id",
				Help:        "The chain id the node.",
				ConstLabels: constLabels,
			},
		),
		GasUsed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "gas_used",
				Help:        "The gas used in the most recent block.",
				ConstLabels: constLabels,
			},
		),
		GasLimit: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "gas_limit",
				Help:        "The gas limit of the most recent block.",
				ConstLabels: constLabels,
			},
		),
		BaseFeePerGas: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "base_fee_per_gas",
				Help:        "The base fee per gas in the most recent block.",
				ConstLabels: constLabels,
			},
		),
		BlockSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "block_size_bytes",
				Help:        "The size of the most recent block (in bytes).",
				ConstLabels: constLabels,
			},
		),
		TransactionCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "transaction_count",
				Help:        "The number of transactions in the most recent block.",
				ConstLabels: constLabels,
			},
		),
	}
}

func (g *GeneralMetrics) Start(ctx context.Context) {
	g.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			g.tick(ctx)
		}
	}
}

func (g *GeneralMetrics) tick(ctx context.Context) {
	if _, err := g.GetMostRecentBlockNumber(ctx); err != nil {
		g.log.WithError(err).Error("failed to get most recent block number")
	}

	if _, err := g.GetGasPrice(ctx); err != nil {
		g.log.WithError(err).Error("failed to get gas price")
	}

	if _, err := g.GetNetworkID(ctx); err != nil {
		g.log.WithError(err).Error("failed to get network id")
	}

	if _, err := g.GetChainID(ctx); err != nil {
		g.log.WithError(err).Error("failed to get chain id")
	}

	if err := g.GetMostRecentBlockStats(ctx); err != nil {
		g.log.WithError(err).Error("failed to get most recent block stats")
	}
}

func (g *GeneralMetrics) GetMostRecentBlockNumber(ctx context.Context) (uint64, error) {
	mostRecentBlockNumber, err := g.client.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	g.MostRecentBlockNumber.Set(float64(mostRecentBlockNumber))

	return mostRecentBlockNumber, nil
}

func (g *GeneralMetrics) GetGasPrice(ctx context.Context) (uint64, error) {
	gasPrice, err := g.client.SuggestGasPrice(ctx)
	if err != nil {
		return 0, err
	}

	g.GasPrice.Set(float64(gasPrice.Uint64()))

	return gasPrice.Uint64(), nil
}

func (g *GeneralMetrics) GetNetworkID(ctx context.Context) (uint64, error) {
	networkID, err := g.client.NetworkID(ctx)
	if err != nil {
		return 0, err
	}

	g.NetworkID.Set(float64(networkID.Uint64()))

	return networkID.Uint64(), nil
}

func (g *GeneralMetrics) GetChainID(ctx context.Context) (uint64, error) {
	chainID, err := g.client.ChainID(ctx)
	if err != nil {
		return 0, err
	}

	g.ChainID.Set(float64(chainID.Uint64()))

	return chainID.Uint64(), nil
}

func (g *GeneralMetrics) GetMostRecentBlockStats(ctx context.Context) error {
	mostRecentBlockNumber, err := g.client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	block, err := g.client.BlockByNumber(ctx, big.NewInt(int64(mostRecentBlockNumber)))
	if err != nil {
		return err
	}

	g.GasUsed.Set(float64(block.GasUsed()))
	g.GasLimit.Set(float64(block.GasLimit()))
	g.BaseFeePerGas.Set(float64(block.BaseFee().Int64()))
	g.BlockSize.Set(float64(block.Size()))
	g.TransactionCount.Add(float64(len(block.Transactions())))

	return nil
}
