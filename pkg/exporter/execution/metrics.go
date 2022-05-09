package execution

import "github.com/prometheus/client_golang/prometheus"

type Metrics interface {
	ObserveSyncPercentage(percent float64)
	ObserveSyncIsSyncing(syncing bool)
	ObserveSyncHighestBlock(block uint64)
	ObserveSyncCurrentBlock(block uint64)
	ObserveSyncStartingBlock(block uint64)
	ObserveNetworkID(networkID int64)
}

type metrics struct {
	syncPercentage    prometheus.Gauge
	syncCurrentBlock  prometheus.Gauge
	syncStartingBlock prometheus.Gauge
	syncIsSyncing     prometheus.Gauge
	syncHighestblock  prometheus.Gauge
	networkID         prometheus.Gauge
}

func NewMetrics(nodeName, namespace string) Metrics {
	constLabels := make(prometheus.Labels)
	constLabels["ethereum_role"] = "execution"
	constLabels["node_name"] = nodeName

	m := &metrics{
		syncPercentage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_percentage",
				Help:        "How synced the node is with the network (0-100%).",
				ConstLabels: constLabels,
			},
		),
		syncStartingBlock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_starting_block",
				Help:        "The starting block of the sync procedure.",
				ConstLabels: constLabels,
			},
		),
		syncCurrentBlock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_current_block",
				Help:        "The current block of the sync procedure.",
				ConstLabels: constLabels,
			},
		),
		syncIsSyncing: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_is_syncing",
				Help:        "1 if the node is in syncing state.",
				ConstLabels: constLabels,
			},
		),
		syncHighestblock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_highest_block",
				Help:        "The highest block of the sync procedure.",
				ConstLabels: constLabels,
			},
		),
		networkID: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "network_id",
				Help:        "The network id of the node.",
				ConstLabels: constLabels,
			},
		),
	}

	prometheus.MustRegister(m.syncPercentage)
	prometheus.MustRegister(m.syncStartingBlock)
	prometheus.MustRegister(m.syncCurrentBlock)
	prometheus.MustRegister(m.syncIsSyncing)
	prometheus.MustRegister(m.syncHighestblock)
	prometheus.MustRegister(m.networkID)

	return m
}

func (m *metrics) ObserveSyncPercentage(percent float64) {
	m.syncPercentage.Set(percent)
}

func (m *metrics) ObserveSyncStartingBlock(block uint64) {
	m.syncStartingBlock.Set(float64(block))
}

func (m *metrics) ObserveSyncCurrentBlock(block uint64) {
	m.syncCurrentBlock.Set(float64(block))
}

func (m *metrics) ObserveSyncIsSyncing(syncing bool) {
	if syncing {
		m.syncIsSyncing.Set(1)
		return
	}

	m.syncIsSyncing.Set(0)
}

func (m *metrics) ObserveSyncHighestBlock(block uint64) {
	m.syncHighestblock.Set(float64(block))
}

func (m *metrics) ObserveNetworkID(networkID int64) {
	m.networkID.Set(float64(networkID))
}
