package beacon

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api/types"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon/state"
)

const (
	// Custom events derived from our pseudo beacon node
	topicEpochChanged       = "epoch_changed"
	topicSlotChanged        = "slot_changed"
	topicEpochSlotChanged   = "epoch_slot_changed"
	topicBlockInserted      = "block_inserted"
	topicReady              = "ready"
	topicSyncStatus         = "sync_status"
	topicNodeVersionUpdated = "node_version_updated"
	topicPeersUpdated       = "peers_updated"
	topicSpecUpdated        = "spec_updated"

	// Official beacon events that are proxied
	topicAttestation          = "attestation"
	topicBlock                = "block"
	topicChainReorg           = "chain_reorg"
	topicFinalizedCheckpoint  = "finalized_checkpoint"
	topicHead                 = "head"
	topicVoluntaryExit        = "voluntary_exit"
	topicContributionAndProof = "contribution_and_proof"
	topicEvent                = "raw_event"
)

type EpochChangedEvent struct {
	Epoch phase0.Epoch
}

type SlotChangedEvent struct {
	Slot phase0.Slot
}

type EpochSlotChangedEvent struct {
	Epoch phase0.Epoch
	Slot  phase0.Slot
}

type BlockInsertedEvent struct {
	Slot phase0.Slot
}

type ReadyEvent struct {
}

type SyncStatusEvent struct {
	State *v1.SyncState
}

type NodeVersionUpdatedEvent struct {
	Version string
}

type PeersUpdatedEvent struct {
	Peers types.Peers
}

type SpecUpdatedEvent struct {
	Spec *state.Spec
}
