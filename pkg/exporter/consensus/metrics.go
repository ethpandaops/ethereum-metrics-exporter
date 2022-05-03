package consensus

import "github.com/prometheus/client_golang/prometheus"

type Metrics interface {
	ObserveSyncPercentage(percent float64)
	ObserveSyncEstimatedHighestSlot(slot uint64)
	ObserveSyncHeadSlot(slot uint64)
	ObserveSyncDistance(slots uint64)
	ObserveSyncIsSyncing(syncing bool)
}

type metrics struct {
	syncPercentage           prometheus.Gauge
	syncEstimatedHighestSlot prometheus.Gauge
	syncHeadSlot             prometheus.Gauge
	syncDistance             prometheus.Gauge
	syncIsSyncing            prometheus.Gauge
}

func NewMetrics(nodeName, namespace string) Metrics {
	constLabels := make(prometheus.Labels)
	constLabels["ethereum_role"] = "consensus"
	constLabels["node_name"] = nodeName

	m := &metrics{
		syncPercentage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "eth_sync_percentage",
				Help:        "How synced the node is with the network (0-100%).",
				ConstLabels: constLabels,
			},
		),
		syncEstimatedHighestSlot: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "eth_sync_estimated_highest_slot",
				Help:        "The estimated highest slot of the network.",
				ConstLabels: constLabels,
			},
		),
		syncHeadSlot: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "eth_sync_head_slot",
				Help:        "The current slot of the node.",
				ConstLabels: constLabels,
			},
		),
		syncDistance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "eth_sync_distance",
				Help:        "The sync distance of the node.",
				ConstLabels: constLabels,
			},
		),
		syncIsSyncing: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "eth_sync_is_syncing",
				Help:        "1 if the node is in syncing state.",
				ConstLabels: constLabels,
			},
		),
	}

	prometheus.MustRegister(m.syncPercentage)
	prometheus.MustRegister(m.syncEstimatedHighestSlot)
	prometheus.MustRegister(m.syncHeadSlot)
	prometheus.MustRegister(m.syncDistance)
	prometheus.MustRegister(m.syncIsSyncing)
	return m
}

func (m *metrics) ObserveSyncPercentage(percent float64) {
	m.syncPercentage.Set(percent)
}

func (m *metrics) ObserveSyncEstimatedHighestSlot(slot uint64) {
	m.syncEstimatedHighestSlot.Set(float64(slot))
}

func (m *metrics) ObserveSyncHeadSlot(slot uint64) {
	m.syncHeadSlot.Set(float64(slot))
}

func (m *metrics) ObserveSyncDistance(slots uint64) {
	m.syncDistance.Set(float64(slots))
}

func (m *metrics) ObserveSyncIsSyncing(syncing bool) {
	if syncing {
		m.syncIsSyncing.Set(1)
		return
	}

	m.syncIsSyncing.Set(0)
}
