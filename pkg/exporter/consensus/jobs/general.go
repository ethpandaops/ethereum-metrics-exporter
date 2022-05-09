package jobs

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GeneralMetrics struct {
	Slots       prometheus.GaugeVec
	NodeVersion prometheus.GaugeVec
	NetworkdID  prometheus.Gauge
}

func NewGeneralMetrics(namespace string, constLabels map[string]string) GeneralMetrics {
	return GeneralMetrics{
		Slots: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "slot_number",
				Help:        "The slot number of the beacon chain.",
				ConstLabels: constLabels,
			},
			[]string{
				"identifier",
			},
		),
		NodeVersion: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "node_version",
				Help:        "The version of the running beacon node.",
				ConstLabels: constLabels,
			},
			[]string{
				"version",
			},
		),
		NetworkdID: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "network_id",
				Help:        "The network id of the node.",
				ConstLabels: constLabels,
			},
		),
	}
}

func (g *GeneralMetrics) ObserveSlot(identifier string, slot uint64) {
	g.Slots.WithLabelValues(identifier).Set(float64(slot))
}

func (g *GeneralMetrics) ObserveNodeVersion(version string) {
	g.NodeVersion.WithLabelValues(version).Set(1)
}

func (g *GeneralMetrics) ObserveNetworkID(networkID uint64) {
	g.NetworkdID.Set(float64(networkID))
}
