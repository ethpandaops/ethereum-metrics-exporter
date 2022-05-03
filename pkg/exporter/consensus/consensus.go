package consensus

import (
	"context"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
)

type Node interface {
	Name() string
	SyncStatus(ctx context.Context) (*SyncStatus, error)
}

type node struct {
	name    string
	URL     string
	client  eth2client.Service
	log     logrus.FieldLogger
	metrics Metrics
}

func NewConsensusNode(ctx context.Context, log logrus.FieldLogger, name string, url string, metrics Metrics) (*node, error) {
	client, err := http.New(ctx,
		http.WithAddress(url),
		http.WithLogLevel(zerolog.WarnLevel),
	)
	if err != nil {
		log.WithError(err).Error("Failed to create consensus client")
	}

	return &node{
		name:    name,
		URL:     url,
		client:  client,
		log:     log,
		metrics: metrics,
	}, nil
}

func (c *node) Name() string {
	return c.name
}

func (c *node) SyncStatus(ctx context.Context) (*SyncStatus, error) {
	provider := c.client.(eth2client.NodeSyncingProvider)
	status, err := provider.NodeSyncing(ctx)
	if err != nil {
		return nil, err
	}

	syncStatus := &SyncStatus{
		IsSyncing:         status.IsSyncing,
		HeadSlot:          uint64(status.HeadSlot),
		SyncDistance:      uint64(status.SyncDistance),
		EstimatedHeadSlot: uint64(status.HeadSlot + status.SyncDistance),
	}

	c.metrics.ObserveSyncPercentage(syncStatus.Percent())
	c.metrics.ObserveSyncEstimatedHighestSlot(syncStatus.EstimatedHeadSlot)
	c.metrics.ObserveSyncHeadSlot(syncStatus.HeadSlot)
	c.metrics.ObserveSyncDistance(syncStatus.SyncDistance)
	c.metrics.ObserveSyncIsSyncing(syncStatus.IsSyncing)

	return syncStatus, nil
}
