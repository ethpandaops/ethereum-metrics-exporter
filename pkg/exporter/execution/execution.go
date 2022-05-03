package execution

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type Node interface {
	Name() string
	SyncStatus(ctx context.Context) (*SyncStatus, error)
}

type node struct {
	name    string
	URL     string
	client  ethclient.Client
	log     logrus.FieldLogger
	metrics Metrics
}

func NewExecutionNode(ctx context.Context, log logrus.FieldLogger, name string, url string, metrics Metrics) (*node, error) {
	// Make a dial call since its easier to create the client. We don't actually care if the
	// client is alive or not at this point.
	client, _ := ethclient.Dial(url)

	return &node{
		name:    name,
		URL:     url,
		client:  *client,
		log:     log,
		metrics: metrics,
	}, nil
}

func (e *node) Name() string {
	return e.name
}

func (e *node) SyncStatus(ctx context.Context) (*SyncStatus, error) {
	status, err := e.client.SyncProgress(ctx)
	if err != nil {
		return nil, err
	}

	if status == nil && err == nil {
		// Not syncing
		syncStatus := &SyncStatus{}
		syncStatus.IsSyncing = false
		e.metrics.ObserveSyncIsSyncing(syncStatus.IsSyncing)
		return syncStatus, nil
	}

	syncStatus := &SyncStatus{
		IsSyncing:     true,
		CurrentBlock:  status.CurrentBlock,
		HighestBlock:  status.HighestBlock,
		StartingBlock: status.StartingBlock,
	}

	e.metrics.ObserveSyncPercentage(syncStatus.Percent())
	e.metrics.ObserveSyncCurrentBlock(syncStatus.CurrentBlock)
	e.metrics.ObserveSyncHighestBlock(syncStatus.HighestBlock)
	e.metrics.ObserveSyncStartingBlock(syncStatus.StartingBlock)
	e.metrics.ObserveSyncIsSyncing(syncStatus.IsSyncing)

	return syncStatus, nil
}
