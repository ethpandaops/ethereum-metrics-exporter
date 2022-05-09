package consensus

import (
	"context"
	"errors"
	"strings"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type Node interface {
	Name() string
	URL() string
	Bootstrapped() bool
	Bootstrap(ctx context.Context) error
	SyncStatus(ctx context.Context) (*SyncStatus, error)
	NodeVersion(ctx context.Context) (string, error)
	Spec(ctx context.Context) (map[string]interface{}, error)
	BlockNumbers(ctx context.Context) (*BlockchainSlots, error)
	Forks(ctx context.Context) ([]Fork, error)
}

type node struct {
	name    string
	url     string
	client  eth2client.Service
	log     logrus.FieldLogger
	metrics Metrics
}

func NewConsensusNode(ctx context.Context, log logrus.FieldLogger, name string, url string, metrics Metrics) (*node, error) {
	return &node{
		name:    name,
		url:     url,
		log:     log,
		metrics: metrics,
	}, nil
}

func (c *node) Name() string {
	return c.name
}

func (c *node) URL() string {
	return c.url
}

func (c *node) Bootstrap(ctx context.Context) error {
	client, err := http.New(ctx,
		http.WithAddress(c.url),
		http.WithLogLevel(zerolog.WarnLevel),
	)
	if err != nil {
		return err
	}

	c.client = client

	return nil
}

func (c *node) Bootstrapped() bool {
	_, isProvider := c.client.(eth2client.NodeSyncingProvider)
	if !isProvider {
		return false
	}

	return true
}

func (c *node) refreshClient(ctx context.Context) error {
	client, err := http.New(ctx,
		http.WithAddress(c.url),
		http.WithLogLevel(zerolog.WarnLevel),
	)
	if err != nil {
		return err
	}

	c.client = client
	return nil
}

func (c *node) SyncStatus(ctx context.Context) (*SyncStatus, error) {
	provider, isProvider := c.client.(eth2client.NodeSyncingProvider)
	if !isProvider {
		c.refreshClient(ctx)
		return nil, errors.New("client does not implement eth2client.NodeSyncingProvider")
	}

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

	c.metrics.ObserveSyncStatus(*syncStatus)

	return syncStatus, nil
}

func (c *node) NodeVersion(ctx context.Context) (string, error) {
	provider, isProvider := c.client.(eth2client.NodeVersionProvider)
	if !isProvider {
		c.refreshClient(ctx)
		return "", errors.New("client does not implement eth2client.NodeVersionProvider")
	}

	version, err := provider.NodeVersion(ctx)
	if err != nil {
		return "", err
	}

	c.metrics.ObserveNodeVersion(version)

	return version, nil
}

func (c *node) Spec(ctx context.Context) (map[string]interface{}, error) {
	provider, isProvider := c.client.(eth2client.SpecProvider)
	if !isProvider {
		c.refreshClient(ctx)
		return nil, errors.New("client does not implement eth2client.SpecProvider")
	}

	spec, err := provider.Spec(ctx)
	if err != nil {
		return nil, err
	}

	c.metrics.ObserveSpec(spec)

	return spec, nil
}

func (c *node) BlockNumbers(ctx context.Context) (*BlockchainSlots, error) {
	provider, isProvider := c.client.(eth2client.BeaconBlockHeadersProvider)
	if !isProvider {
		c.refreshClient(ctx)
		return nil, errors.New("client does not implement eth2client.BeaconBlockHeadersProvider")
	}

	errs := []error{}

	slots := &BlockchainSlots{}

	head, err := provider.BeaconBlockHeader(ctx, "head")
	if err != nil {
		errs = append(errs, err)
	} else {
		slots.Head = uint64(head.Header.Message.Slot)
	}

	genesis, err := provider.BeaconBlockHeader(ctx, "genesis")
	if err != nil {
		errs = append(errs, err)
	} else {
		slots.Genesis = uint64(genesis.Header.Message.Slot)
	}

	finalized, err := provider.BeaconBlockHeader(ctx, "finalized")
	if err != nil {
		errs = append(errs, err)
	} else {
		slots.Finalized = uint64(finalized.Header.Message.Slot)
	}

	if len(errs) > 0 {
		errMsg := ""
		for _, e := range errs {
			errMsg += e.Error() + ", "
		}

		return slots, errors.New(errMsg)
	}

	c.metrics.ObserveBlockchainSlots(*slots)

	return slots, nil
}

func (c *node) Forks(ctx context.Context) ([]Fork, error) {
	// Extract the forks out of the spec.
	spec, err := c.Spec(ctx)
	if err != nil {
		return nil, err
	}

	var forks []Fork
	for k, v := range spec {
		if strings.Contains(k, "_FORK_EPOCH") {
			fork := Fork{
				Name:  strings.Replace(k, "_FORK_EPOCH", "", -1),
				Epoch: cast.ToUint64(v),
			}

			forks = append(forks, fork)
		}
	}

	c.metrics.ObserveForks(forks)

	return forks, nil
}
