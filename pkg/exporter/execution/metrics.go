package execution

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/jobs"
)

type Metrics interface {
	ObserveSyncStatus(status SyncStatus)
	ObserveNetworkID(networkID int64)
	ObserveChainID(chainID int64)
	ObserveMostRecentBlock(block int64)
	ObserveGasPrice(gasPrice float64)
}

type metrics struct {
	syncMetrics    jobs.SyncStatus
	generalMetrics jobs.GeneralMetrics
}

func NewMetrics(nodeName, namespace string) Metrics {
	constLabels := make(prometheus.Labels)
	constLabels["ethereum_role"] = "execution"
	constLabels["node_name"] = nodeName

	m := &metrics{

		syncMetrics:    jobs.NewSyncStatus(namespace, constLabels),
		generalMetrics: jobs.NewGeneralMetrics(namespace, constLabels),
	}

	prometheus.MustRegister(m.syncMetrics.Percentage)
	prometheus.MustRegister(m.syncMetrics.StartingBlock)
	prometheus.MustRegister(m.syncMetrics.CurrentBlock)
	prometheus.MustRegister(m.syncMetrics.IsSyncing)
	prometheus.MustRegister(m.syncMetrics.Highestblock)
	prometheus.MustRegister(m.generalMetrics.NetworkID)
	prometheus.MustRegister(m.generalMetrics.GasPrice)
	prometheus.MustRegister(m.generalMetrics.MostRecentBlockNumber)
	prometheus.MustRegister(m.generalMetrics.ChainID)

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
	m.generalMetrics.ObserveNetworkID(networkID)
}

func (m *metrics) ObserveMostRecentBlock(block int64) {
	m.generalMetrics.ObserveMostRecentBlock(block)
}

func (m *metrics) ObserveGasPrice(gasPrice float64) {
	m.generalMetrics.ObserveGasPrice(gasPrice)
}

func (m *metrics) ObserveChainID(chainID int64) {
	m.generalMetrics.ObserveChainID(chainID)
}
