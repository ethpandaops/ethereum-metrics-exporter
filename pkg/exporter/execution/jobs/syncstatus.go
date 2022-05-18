package jobs

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/sirupsen/logrus"
)

// SyncStatus exposes metrics about the sync status of the node.
type SyncStatus struct {
	client        *ethclient.Client
	api           api.ExecutionClient
	ethRPCClient  *ethrpc.EthRPC
	log           logrus.FieldLogger
	Percentage    prometheus.Gauge
	CurrentBlock  prometheus.Gauge
	StartingBlock prometheus.Gauge
	IsSyncing     prometheus.Gauge
	HighestBlock  prometheus.Gauge
}

const (
	NameSyncStatus = "sync"
)

func (s *SyncStatus) Name() string {
	return NameSyncStatus
}

func (s *SyncStatus) RequiredModules() []string {
	return []string{"eth"}
}

type syncingStatus struct {
	IsSyncing     bool
	StartingBlock uint64
	CurrentBlock  uint64
	HighestBlock  uint64
}

func (s *syncingStatus) Percent() float64 {
	if !s.IsSyncing {
		return 100 //notlint:gomnd
	}

	return float64(s.CurrentBlock) / float64(s.HighestBlock) * 100 //notlint:gomnd // 100 will never change.
}

// NewSyncStatus returns a new SyncStatus instance.
func NewSyncStatus(client *ethclient.Client, internalAPI api.ExecutionClient, ethRPCClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) SyncStatus {
	constLabels["module"] = NameSyncStatus

	namespace += "_sync"

	return SyncStatus{
		client:       client,
		api:          internalAPI,
		ethRPCClient: ethRPCClient,
		log:          log.WithField("module", NameSyncStatus),
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
		HighestBlock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "highest_block",
				Help:        "The highest block of the sync procedure.",
				ConstLabels: constLabels,
			},
		),
	}
}

func (s *SyncStatus) Start(ctx context.Context) {
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

func (s *SyncStatus) tick(ctx context.Context) {
	if err := s.GetSyncStatus(ctx); err != nil {
		s.log.Errorf("Failed to get sync status: %s", err)
	}
}

func (s *SyncStatus) GetSyncStatus(ctx context.Context) error {
	status, err := s.client.SyncProgress(ctx)
	if err != nil {
		return err
	}

	if status == nil && err == nil {
		// Not syncing
		ss := &syncingStatus{}
		ss.IsSyncing = false
		s.observeStatus(ss)

		return nil
	}

	syncStatus := &syncingStatus{
		IsSyncing:     true,
		CurrentBlock:  status.CurrentBlock,
		HighestBlock:  status.HighestBlock,
		StartingBlock: status.StartingBlock,
	}

	s.observeStatus(syncStatus)

	return nil
}

func (s *SyncStatus) observeStatus(status *syncingStatus) {
	if status.IsSyncing {
		s.IsSyncing.Set(1)
	} else {
		s.IsSyncing.Set(0)
	}

	s.StartingBlock.Set(float64(status.StartingBlock))
	s.CurrentBlock.Set(float64(status.CurrentBlock))
	s.HighestBlock.Set(float64(status.HighestBlock))
	s.Percentage.Set(status.Percent())
}
