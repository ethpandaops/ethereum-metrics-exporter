package beacon

import (
	"context"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/nats-io/nats.go"
)

// Official Beacon events
func (n *node) OnBlock(ctx context.Context, handler func(ctx context.Context, event *v1.BlockEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicBlock, func(event *v1.BlockEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnAttestation(ctx context.Context, handler func(ctx context.Context, event *phase0.Attestation) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicAttestation, func(event *phase0.Attestation) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnChainReOrg(ctx context.Context, handler func(ctx context.Context, event *v1.ChainReorgEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicChainReorg, func(event *v1.ChainReorgEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnFinalizedCheckpoint(ctx context.Context, handler func(ctx context.Context, event *v1.FinalizedCheckpointEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicFinalizedCheckpoint, func(event *v1.FinalizedCheckpointEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnHead(ctx context.Context, handler func(ctx context.Context, event *v1.HeadEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicHead, func(event *v1.HeadEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnVoluntaryExit(ctx context.Context, handler func(ctx context.Context, event *phase0.VoluntaryExit) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicHead, func(event *phase0.VoluntaryExit) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnEvent(ctx context.Context, handler func(ctx context.Context, event *v1.Event) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicEvent, func(event *v1.Event) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

// Custom Events
func (n *node) OnEpochChanged(ctx context.Context, handler func(ctx context.Context, event *EpochChangedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicEpochChanged, func(event *EpochChangedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnSlotChanged(ctx context.Context, handler func(ctx context.Context, event *SlotChangedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicSlotChanged, func(event *SlotChangedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnEpochSlotChanged(ctx context.Context, handler func(ctx context.Context, event *EpochSlotChangedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicEpochSlotChanged, func(event *EpochSlotChangedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnBlockInserted(ctx context.Context, handler func(ctx context.Context, event *BlockInsertedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicBlockInserted, func(event *BlockInsertedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnReady(ctx context.Context, handler func(ctx context.Context, event *ReadyEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicReady, func(event *ReadyEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnSyncStatus(ctx context.Context, handler func(ctx context.Context, event *SyncStatusEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicSyncStatus, func(event *SyncStatusEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnNodeVersionUpdated(ctx context.Context, handler func(ctx context.Context, event *NodeVersionUpdatedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicNodeVersionUpdated, func(event *NodeVersionUpdatedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnPeersUpdated(ctx context.Context, handler func(ctx context.Context, event *PeersUpdatedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicPeersUpdated, func(event *PeersUpdatedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func (n *node) OnSpecUpdated(ctx context.Context, handler func(ctx context.Context, event *SpecUpdatedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicSpecUpdated, func(event *SpecUpdatedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}
