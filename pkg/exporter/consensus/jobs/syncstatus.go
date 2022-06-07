package jobs

import (
	"context"
	"errors"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/sirupsen/logrus"
)

// Sync reports metrics on the sync status of the node.
type Sync struct {
	client               eth2client.Service
	log                  logrus.FieldLogger
	Percentage           prometheus.Gauge
	EstimatedHighestSlot prometheus.Gauge
	HeadSlot             prometheus.Gauge
	Distance             prometheus.Gauge
	IsSyncing            prometheus.Gauge
}

const (
	NameSync = "sync"
)

// NewSyncJob returns a new Sync instance.
func NewSyncJob(client eth2client.Service, ap api.ConsensusClient, log logrus.FieldLogger, namespace string, constLabels map[string]string) Sync {
	constLabels["module"] = NameSync

	namespace += "_sync"

	return Sync{
		client: client,
		log:    log,
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

func (s *Sync) Name() string {
	return NameSync
}

func (s *Sync) Start(ctx context.Context) {
	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			s.tick(ctx)
		}
	}
}

func (s *Sync) HandleEvent(ctx context.Context, event *v1.Event) {
}

func (s *Sync) tick(ctx context.Context) {
	if err := s.GetSyncState(ctx); err != nil {
		s.log.WithError(err).Error("failed to get sync state")
	}
}

func (s *Sync) GetSyncState(ctx context.Context) error {
	provider, isProvider := s.client.(eth2client.NodeSyncingProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.NodeSyncingProvider")
	}

	status, err := provider.NodeSyncing(ctx)
	if err != nil {
		return err
	}

	s.ObserveSyncDistance(uint64(status.SyncDistance))
	s.ObserveSyncHeadSlot(uint64(status.HeadSlot))
	s.ObserveSyncIsSyncing(status.IsSyncing)

	estimatedHighestHeadSlot := status.SyncDistance + status.HeadSlot
	s.ObserveSyncEstimatedHighestSlot(uint64(estimatedHighestHeadSlot))

	percent := (float64(status.HeadSlot) / float64(estimatedHighestHeadSlot) * 100)
	if !status.IsSyncing {
		percent = 100
	}

	s.ObserveSyncPercentage(percent)

	return nil
}

func (s *Sync) ObserveSyncPercentage(percent float64) {
	s.Percentage.Set(percent)
}

func (s *Sync) ObserveSyncEstimatedHighestSlot(slot uint64) {
	s.EstimatedHighestSlot.Set(float64(slot))
}

func (s *Sync) ObserveSyncHeadSlot(slot uint64) {
	s.HeadSlot.Set(float64(slot))
}

func (s *Sync) ObserveSyncDistance(slots uint64) {
	s.Distance.Set(float64(slots))
}

func (s *Sync) ObserveSyncIsSyncing(syncing bool) {
	if syncing {
		s.IsSyncing.Set(1)
		return
	}

	s.IsSyncing.Set(0)
}
