package beacon

import (
	"context"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/nats-io/nats.go"
)

func (n *node) handleSubscriberError(err error, topic string) {
	if err != nil {
		n.log.WithError(err).WithField("topic", topic).Error("Subscriber error")
	}
}

// Official Beacon events
func (n *node) OnBlock(ctx context.Context, handler func(ctx context.Context, event *v1.BlockEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicBlock, func(event *v1.BlockEvent) {
		n.handleSubscriberError(handler(ctx, event), topicBlock)
	})
}

func (n *node) OnAttestation(ctx context.Context, handler func(ctx context.Context, event *phase0.Attestation) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicAttestation, func(event *phase0.Attestation) {
		n.handleSubscriberError(handler(ctx, event), topicAttestation)
	})
}

func (n *node) OnChainReOrg(ctx context.Context, handler func(ctx context.Context, event *v1.ChainReorgEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicChainReorg, func(event *v1.ChainReorgEvent) {
		n.handleSubscriberError(handler(ctx, event), topicChainReorg)
	})
}

func (n *node) OnFinalizedCheckpoint(ctx context.Context, handler func(ctx context.Context, event *v1.FinalizedCheckpointEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicFinalizedCheckpoint, func(event *v1.FinalizedCheckpointEvent) {
		n.handleSubscriberError(handler(ctx, event), topicFinalizedCheckpoint)
	})
}

func (n *node) OnHead(ctx context.Context, handler func(ctx context.Context, event *v1.HeadEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicHead, func(event *v1.HeadEvent) {
		n.handleSubscriberError(handler(ctx, event), topicHead)
	})
}

func (n *node) OnVoluntaryExit(ctx context.Context, handler func(ctx context.Context, event *phase0.VoluntaryExit) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicVoluntaryExit, func(event *phase0.VoluntaryExit) {
		n.handleSubscriberError(handler(ctx, event), topicVoluntaryExit)
	})
}

func (n *node) OnEvent(ctx context.Context, handler func(ctx context.Context, event *v1.Event) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicEvent, func(event *v1.Event) {
		n.handleSubscriberError(handler(ctx, event), topicEvent)
	})
}

// Custom Events
func (n *node) OnEpochChanged(ctx context.Context, handler func(ctx context.Context, event *EpochChangedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicEpochChanged, func(event *EpochChangedEvent) {
		n.handleSubscriberError(handler(ctx, event), topicEpochChanged)
	})
}

func (n *node) OnSlotChanged(ctx context.Context, handler func(ctx context.Context, event *SlotChangedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicSlotChanged, func(event *SlotChangedEvent) {
		n.handleSubscriberError(handler(ctx, event), topicSlotChanged)
	})
}

func (n *node) OnEpochSlotChanged(ctx context.Context, handler func(ctx context.Context, event *EpochSlotChangedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicEpochSlotChanged, func(event *EpochSlotChangedEvent) {
		n.handleSubscriberError(handler(ctx, event), topicEpochSlotChanged)
	})
}

func (n *node) OnBlockInserted(ctx context.Context, handler func(ctx context.Context, event *BlockInsertedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicBlockInserted, func(event *BlockInsertedEvent) {
		n.handleSubscriberError(handler(ctx, event), topicBlockInserted)
	})
}

func (n *node) OnReady(ctx context.Context, handler func(ctx context.Context, event *ReadyEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicReady, func(event *ReadyEvent) {
		n.handleSubscriberError(handler(ctx, event), topicReady)
	})
}

func (n *node) OnSyncStatus(ctx context.Context, handler func(ctx context.Context, event *SyncStatusEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicSyncStatus, func(event *SyncStatusEvent) {
		n.handleSubscriberError(handler(ctx, event), topicSyncStatus)
	})
}

func (n *node) OnNodeVersionUpdated(ctx context.Context, handler func(ctx context.Context, event *NodeVersionUpdatedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicNodeVersionUpdated, func(event *NodeVersionUpdatedEvent) {
		n.handleSubscriberError(handler(ctx, event), topicNodeVersionUpdated)
	})
}

func (n *node) OnPeersUpdated(ctx context.Context, handler func(ctx context.Context, event *PeersUpdatedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicPeersUpdated, func(event *PeersUpdatedEvent) {
		n.handleSubscriberError(handler(ctx, event), topicPeersUpdated)
	})
}

func (n *node) OnSpecUpdated(ctx context.Context, handler func(ctx context.Context, event *SpecUpdatedEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicSpecUpdated, func(event *SpecUpdatedEvent) {
		n.handleSubscriberError(handler(ctx, event), topicSpecUpdated)
	})
}

func (n *node) OnEmptySlot(ctx context.Context, handler func(ctx context.Context, event *EmptySlotEvent) error) (*nats.Subscription, error) {
	return n.broker.Subscribe(topicEmptySlot, func(event *EmptySlotEvent) {
		n.handleSubscriberError(handler(ctx, event), topicEmptySlot)
	})
}
