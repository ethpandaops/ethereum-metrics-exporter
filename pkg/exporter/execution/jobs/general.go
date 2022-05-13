package jobs

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/sirupsen/logrus"
)

type GeneralMetrics struct {
	MetricExporter
	client                *ethclient.Client
	api                   api.ExecutionClient
	log                   logrus.FieldLogger
	MostRecentBlockNumber prometheus.Gauge
	GasPrice              prometheus.Gauge
	NetworkID             prometheus.Gauge
	ChainID               prometheus.Gauge
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
