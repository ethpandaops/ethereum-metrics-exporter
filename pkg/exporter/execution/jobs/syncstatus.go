package jobs

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SyncStatus struct {
	Percentage    prometheus.Gauge
	CurrentBlock  prometheus.Gauge
	StartingBlock prometheus.Gauge
	IsSyncing     prometheus.Gauge
	Highestblock  prometheus.Gauge
}

func NewSyncStatus(namespace string, constLabels map[string]string) SyncStatus {
	namespace = namespace + "_sync"
	return SyncStatus{
		Percentage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "percentage",
				Help:        "How synced the node is with the network (0-100%).",
				ConstLabels: constLabels,
			},
		),
		StartingBlock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "starting_block",
				Help:        "The starting block of the sync procedure.",
				ConstLabels: constLabels,
			},
		),
		CurrentBlock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "current_block",
				Help:        "The current block of the sync procedure.",
				ConstLabels: constLabels,
			},
		),
		IsSyncing: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "is_syncing",
				Help:        "1 if the node is in syncing state.",
				ConstLabels: constLabels,
			},
		),
		Highestblock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "highest_block",
				Help:        "The highest block of the sync procedure.",
				ConstLabels: constLabels,
			},
		),
	}
}

func (s *SyncStatus) ObserveSyncPercentage(percent float64) {
	s.Percentage.Set(percent)
}

func (s *SyncStatus) ObserveSyncStartingBlock(block uint64) {
	s.StartingBlock.Set(float64(block))
}

func (s *SyncStatus) ObserveSyncCurrentBlock(block uint64) {
	s.CurrentBlock.Set(float64(block))
}

func (s *SyncStatus) ObserveSyncIsSyncing(syncing bool) {
	if syncing {
		s.IsSyncing.Set(1)
		return
	}

	s.IsSyncing.Set(0)
}

func (s *SyncStatus) ObserveSyncHighestBlock(block uint64) {
	s.Highestblock.Set(float64(block))
}
