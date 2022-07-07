package beacon

import (
	"context"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api/types"
)

// Official beacon events that are proxied
func (n *node) publishBlock(ctx context.Context, event *v1.BlockEvent) error {
	return n.broker.Publish(topicBlock, event)
}

func (n *node) publishAttestation(ctx context.Context, event *phase0.Attestation) error {
	return n.broker.Publish(topicAttestation, event)
}

func (n *node) publishChainReOrg(ctx context.Context, event *v1.ChainReorgEvent) error {
	return n.broker.Publish(topicChainReorg, event)
}

func (n *node) publishFinalizedCheckpoint(ctx context.Context, event *v1.FinalizedCheckpointEvent) error {
	return n.broker.Publish(topicFinalizedCheckpoint, event)
}

func (n *node) publishHead(ctx context.Context, event *v1.HeadEvent) error {
	return n.broker.Publish(topicHead, event)
}

func (n *node) publishVoluntaryExit(ctx context.Context, event *phase0.VoluntaryExit) error {
	return n.broker.Publish(topicVoluntaryExit, event)
}

func (n *node) publishEvent(ctx context.Context, event *v1.Event) error {
	return n.broker.Publish(topicEvent, event)
}

// Custom Events derived from our pseudo beacon node
func (n *node) publishReady(ctx context.Context) error {
	return n.broker.Publish(topicReady, nil)
}

func (n *node) publishEpochChanged(ctx context.Context, epoch phase0.Epoch) error {
	return n.broker.Publish(topicEpochChanged, &EpochChangedEvent{
		Epoch: epoch,
	})
}

func (n *node) publishSlotChanged(ctx context.Context, slot phase0.Slot) error {
	return n.broker.Publish(topicSlotChanged, &SlotChangedEvent{
		Slot: slot,
	})
}

func (n *node) publishEpochSlotChanged(ctx context.Context, epoch phase0.Epoch, slot phase0.Slot) error {
	return n.broker.Publish(topicEpochSlotChanged, &EpochSlotChangedEvent{
		Epoch: epoch,
		Slot:  slot,
	})
}

func (n *node) publishBlockInserted(ctx context.Context, slot phase0.Slot) error {
	return n.broker.Publish(topicBlockInserted, &BlockInsertedEvent{
		Slot: slot,
	})
}

func (n *node) publishSyncStatus(ctx context.Context, state *v1.SyncState) error {
	return n.broker.Publish(topicSyncStatus, &SyncStatusEvent{
		State: state,
	})
}

func (n *node) publishNodeVersionUpdated(ctx context.Context, version string) error {
	return n.broker.Publish(topicNodeVersionUpdated, &NodeVersionUpdatedEvent{
		Version: version,
	})
}

func (n *node) publishPeersUpdated(ctx context.Context, peers types.Peers) error {
	return n.broker.Publish(topicPeersUpdated, &PeersUpdatedEvent{
		Peers: peers,
	})
}
