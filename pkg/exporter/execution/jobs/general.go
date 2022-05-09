package jobs

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GeneralMetrics struct {
	MostRecentBlockNumber prometheus.Gauge
	GasPrice              prometheus.Gauge
	NetworkID             prometheus.Gauge
	ChainID               prometheus.Gauge
}

func NewGeneralMetrics(namespace string, constLabels map[string]string) GeneralMetrics {
	return GeneralMetrics{
		MostRecentBlockNumber: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "most_recent_number",
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

func (g *GeneralMetrics) ObserveMostRecentBlock(blockNumber int64) {
	g.MostRecentBlockNumber.Set(float64(blockNumber))
}

func (g *GeneralMetrics) ObserveGasPrice(gasPrice float64) {
	g.GasPrice.Set(gasPrice)
}

func (g *GeneralMetrics) ObserveNetworkID(networkID int64) {
	g.NetworkID.Set(float64(networkID))
}

func (g *GeneralMetrics) ObserveChainID(id int64) {
	g.ChainID.Set(float64(id))
}
