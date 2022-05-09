package execution

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/jobs"
)

type Metrics interface {
	ObserveSyncStatus(status SyncStatus)
	ObserveNetworkID(networkID int64)
}

type metrics struct {
	networkID prometheus.Gauge

	syncMetrics jobs.SyncStatus
}

func NewMetrics(nodeName, namespace string) Metrics {
	constLabels := make(prometheus.Labels)
	constLabels["ethereum_role"] = "execution"
	constLabels["node_name"] = nodeName

	m := &metrics{
		networkID: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "network_id",
				Help:        "The network id of the node.",
				ConstLabels: constLabels,
			},
		),
		syncMetrics: jobs.NewSyncStatus(namespace, constLabels),
	}

	prometheus.MustRegister(m.syncMetrics.Percentage)
	prometheus.MustRegister(m.syncMetrics.StartingBlock)
	prometheus.MustRegister(m.syncMetrics.CurrentBlock)
	prometheus.MustRegister(m.syncMetrics.IsSyncing)
	prometheus.MustRegister(m.syncMetrics.Highestblock)
	prometheus.MustRegister(m.networkID)

	return m
}

func (m *metrics) ObserveSyncStatus(status SyncStatus) {
	m.syncMetrics.ObserveSyncPercentage(status.Percent())
	m.syncMetrics.ObserveSyncCurrentBlock(status.CurrentBlock)
	m.syncMetrics.ObserveSyncHighestBlock(status.HighestBlock)
	m.syncMetrics.ObserveSyncStartingBlock(status.StartingBlock)
	m.syncMetrics.ObserveSyncIsSyncing(status.IsSyncing)
}

func (m *metrics) ObserveNetworkID(networkID int64) {
	m.networkID.Set(float64(networkID))
}
