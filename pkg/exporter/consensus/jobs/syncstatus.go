package jobs

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SyncStatus struct {
	Percentage           prometheus.Gauge
	EstimatedHighestSlot prometheus.Gauge
	HeadSlot             prometheus.Gauge
	Distance             prometheus.Gauge
	IsSyncing            prometheus.Gauge
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
		EstimatedHighestSlot: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "estimated_highest_slot",
				Help:        "The estimated highest slot of the network.",
				ConstLabels: constLabels,
			},
		),
		HeadSlot: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_slot",
				Help:        "The current slot of the node.",
				ConstLabels: constLabels,
			},
		),
		Distance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "distance",
				Help:        "The sync distance of the node.",
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
	}
}

func (s *SyncStatus) ObserveSyncPercentage(percent float64) {
	s.Percentage.Set(percent)
}

func (s *SyncStatus) ObserveSyncEstimatedHighestSlot(slot uint64) {
	s.EstimatedHighestSlot.Set(float64(slot))
}

func (s *SyncStatus) ObserveSyncHeadSlot(slot uint64) {
	s.HeadSlot.Set(float64(slot))
}

func (s *SyncStatus) ObserveSyncDistance(slots uint64) {
	s.Distance.Set(float64(slots))
}

func (s *SyncStatus) ObserveSyncIsSyncing(syncing bool) {
	if syncing {
		s.IsSyncing.Set(1)
		return
	}

	s.IsSyncing.Set(0)
}
