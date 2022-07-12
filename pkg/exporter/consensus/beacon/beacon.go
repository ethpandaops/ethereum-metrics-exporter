package beacon

import (
	"context"
	"errors"
	"fmt"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats.go"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api/types"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon/state"
	"github.com/sirupsen/logrus"
)

type Node interface {
	// Lifecycle
	// Start starts the node.
	Start(ctx context.Context) error
	// StartAsync starts the node asynchronously.
	StartAsync(ctx context.Context)

	// Getters
	// GetEpoch returns the epoch for the given epoch.
	GetEpoch(ctx context.Context, epoch phase0.Epoch) (*state.Epoch, error)
	// GetSlot returns the slot for the given slot.
	GetSlot(ctx context.Context, slot phase0.Slot) (*state.Slot, error)
	// GetSpec returns the spec for the node.
	GetSpec(ctx context.Context) (*state.Spec, error)
	// GetSyncState returns the sync state for the node.
	GetSyncState(ctx context.Context) (*v1.SyncState, error)
	// GetGenesis returns the genesis for the node.
	GetGenesis(ctx context.Context) (*v1.Genesis, error)
	// GetNodeVersion returns the node version.
	GetNodeVersion(ctx context.Context) (string, error)

	// Subscriptions
	// - Proxied Beacon events
	// OnEvent is called when a beacon event is received.
	OnEvent(ctx context.Context, handler func(ctx context.Context, ev *v1.Event) error) (*nats.Subscription, error)
	// OnBlock is called when a block is received.
	OnBlock(ctx context.Context, handler func(ctx context.Context, ev *v1.BlockEvent) error) (*nats.Subscription, error)
	// OnAttestation is called when an attestation is received.
	OnAttestation(ctx context.Context, handler func(ctx context.Context, ev *phase0.Attestation) error) (*nats.Subscription, error)
	// OnFinalizedCheckpoint is called when a finalized checkpoint is received.
	OnFinalizedCheckpoint(ctx context.Context, handler func(ctx context.Context, ev *v1.FinalizedCheckpointEvent) error) (*nats.Subscription, error)
	// OnHead is called when the head is received.
	OnHead(ctx context.Context, handler func(ctx context.Context, ev *v1.HeadEvent) error) (*nats.Subscription, error)
	// OnChainReOrg is called when a chain reorg is received.
	OnChainReOrg(ctx context.Context, handler func(ctx context.Context, ev *v1.ChainReorgEvent) error) (*nats.Subscription, error)
	// OnVoluntaryExit is called when a voluntary exit is received.
	OnVoluntaryExit(ctx context.Context, handler func(ctx context.Context, ev *phase0.VoluntaryExit) error) (*nats.Subscription, error)

	// - Custom events
	// OnReady is called when the node is ready.
	OnReady(ctx context.Context, handler func(ctx context.Context, event *ReadyEvent) error) (*nats.Subscription, error)
	// OnEpochChanged is called when the current epoch changes.
	OnEpochChanged(ctx context.Context, handler func(ctx context.Context, event *EpochChangedEvent) error) (*nats.Subscription, error)
	// OnSlotChanged is called when the current slot changes.
	OnSlotChanged(ctx context.Context, handler func(ctx context.Context, event *SlotChangedEvent) error) (*nats.Subscription, error)
	// OnEpochSlotChanged is called when the current epoch or slot changes.
	OnEpochSlotChanged(ctx context.Context, handler func(ctx context.Context, event *EpochSlotChangedEvent) error) (*nats.Subscription, error)
	// OnBlockInserted is called when a block is inserted.
	OnBlockInserted(ctx context.Context, handler func(ctx context.Context, event *BlockInsertedEvent) error) (*nats.Subscription, error)
	// OnSyncStatus is called when the sync status changes.
	OnSyncStatus(ctx context.Context, handler func(ctx context.Context, event *SyncStatusEvent) error) (*nats.Subscription, error)
	// OnNodeVersionUpdated is called when the node version is updated.
	OnNodeVersionUpdated(ctx context.Context, handler func(ctx context.Context, event *NodeVersionUpdatedEvent) error) (*nats.Subscription, error)
	// OnPeersUpdated is called when the peers are updated.
	OnPeersUpdated(ctx context.Context, handler func(ctx context.Context, event *PeersUpdatedEvent) error) (*nats.Subscription, error)
	// OnSpecUpdated is called when the spec is updated.
	OnSpecUpdated(ctx context.Context, handler func(ctx context.Context, event *SpecUpdatedEvent) error) (*nats.Subscription, error)
	// OnEmptySlot is called when an empty slot is detected.
	OnEmptySlot(ctx context.Context, handler func(ctx context.Context, event *EmptySlotEvent) error) (*nats.Subscription, error)
}

// Node represents an Ethereum beacon node. It computes values based on the spec.
type node struct {
	// Helpers
	log logrus.FieldLogger

	// Clients
	api    api.ConsensusClient
	client eth2client.Service
	broker *nats.EncodedConn

	// Internal data stores
	genesis       *v1.Genesis
	state         *state.Container
	lastEventTime time.Time
	syncing       *v1.SyncState
	nodeVersion   string
	peers         types.Peers
}

func NewNode(ctx context.Context, log logrus.FieldLogger, ap api.ConsensusClient, client eth2client.Service, broker *nats.EncodedConn) Node {
	return &node{
		log:    log.WithField("module", "consensus/beacon"),
		api:    ap,
		client: client,
		broker: broker,

		syncing: &v1.SyncState{
			IsSyncing: false,
		},
	}
}

func (n *node) Start(ctx context.Context) error {
	if err := n.bootstrap(ctx); err != nil {
		return err
	}

	if err := n.fetchSyncStatus(ctx); err != nil {
		return err
	}

	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every("15s").Do(func() {
		if err := n.fetchSyncStatus(ctx); err != nil {
			n.log.WithError(err).Error("Failed to fetch sync status")
		}
	}); err != nil {
		return err
	}

	if _, err := s.Every("15m").Do(func() {
		if err := n.fetchNodeVersion(ctx); err != nil {
			n.log.WithError(err).Error("Failed to fetch node version")
		}
	}); err != nil {
		return err
	}

	if _, err := s.Every("15s").Do(func() {
		if err := n.fetchPeers(ctx); err != nil {
			n.log.WithError(err).Error("Failed to fetch peers")
		}
	}); err != nil {
		return err
	}

	s.StartAsync()

	return nil
}

func (n *node) StartAsync(ctx context.Context) {
	go func() {
		if err := n.Start(ctx); err != nil {
			n.log.WithError(err).Error("Failed to start beacon node")
		}
	}()
}

func (n *node) GetEpoch(ctx context.Context, epoch phase0.Epoch) (*state.Epoch, error) {
	if n.state == nil {
		return nil, errors.New("state is not initialized")
	}

	return n.state.GetEpoch(ctx, epoch)
}

func (n *node) GetSlot(ctx context.Context, slot phase0.Slot) (*state.Slot, error) {
	if n.state == nil {
		return nil, errors.New("state is not initialized")
	}

	return n.state.GetSlot(ctx, slot)
}

func (n *node) GetSpec(ctx context.Context) (*state.Spec, error) {
	if n.state == nil {
		return nil, errors.New("state is not initialized")
	}

	sp := n.state.Spec()

	if sp == nil {
		return nil, errors.New("spec not yet available")
	}

	return sp, nil
}

func (n *node) GetSyncState(ctx context.Context) (*v1.SyncState, error) {
	return n.syncing, nil
}

func (n *node) GetGenesis(ctx context.Context) (*v1.Genesis, error) {
	return n.genesis, nil
}

func (n *node) GetNodeVersion(ctx context.Context) (string, error) {
	return n.nodeVersion, nil
}

func (n *node) bootstrap(ctx context.Context) error {
	if err := n.initializeState(ctx); err != nil {
		return err
	}

	if err := n.subscribeDownstream(ctx); err != nil {
		return err
	}

	if err := n.subscribeToSelf(ctx); err != nil {
		return err
	}

	if err := n.publishReady(ctx); err != nil {
		return err
	}

	//nolint:errcheck // we dont care if this errors out since it runs indefinitely in a goroutine
	go n.ensureBeaconSubscription(ctx)

	return nil
}

func (n *node) fetchSyncStatus(ctx context.Context) error {
	provider, isProvider := n.client.(eth2client.NodeSyncingProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.NodeSyncingProvider")
	}

	status, err := provider.NodeSyncing(ctx)
	if err != nil {
		return err
	}

	n.syncing = status

	if err := n.publishSyncStatus(ctx, status); err != nil {
		return err
	}

	return nil
}

func (n *node) fetchPeers(ctx context.Context) error {
	peers, err := n.api.NodePeers(ctx)
	if err != nil {
		return err
	}

	n.peers = peers

	return n.publishPeersUpdated(ctx, peers)
}

func (n *node) subscribeToSelf(ctx context.Context) error {
	// Listen for beacon block events and insert them in to our state
	if _, err := n.OnBlock(ctx, func(ctx context.Context, ev *v1.BlockEvent) error {
		if n.syncing.IsSyncing {
			return nil
		}

		start := time.Now()

		// Sleep a little for the beacon node to actually save the block
		time.Sleep(200 * time.Millisecond)

		// Grab the entire block from the beacon node
		block, err := n.getBlock(ctx, fmt.Sprintf("%v", ev.Slot))
		if err != nil {
			return err
		}

		if block == nil {
			return errors.New("fetched block is nil")
		}

		// Insert the beacon block into the state
		if err := n.state.AddBeaconBlock(ctx, block, start); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) subscribeDownstream(ctx context.Context) error {
	if err := n.state.OnEpochSlotChanged(ctx, n.handleStateEpochSlotChanged); err != nil {
		return err
	}

	if err := n.state.OnBlockInserted(ctx, n.handleDownstreamBlockInserted); err != nil {
		return err
	}

	if err := n.state.OnEmptySlot(ctx, n.handleDownstreamEmptySlot); err != nil {
		return err
	}

	return nil
}

func (n *node) fetchNodeVersion(ctx context.Context) error {
	provider, isProvider := n.client.(eth2client.NodeVersionProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.NodeVersionProvider")
	}

	version, err := provider.NodeVersion(ctx)
	if err != nil {
		return err
	}

	n.nodeVersion = version

	return n.publishNodeVersionUpdated(ctx, version)
}

func (n *node) handleDownstreamBlockInserted(ctx context.Context, epoch phase0.Epoch, slot state.Slot) error {
	if err := n.publishBlockInserted(ctx, slot.Number()); err != nil {
		return err
	}

	return nil
}

func (n *node) handleDownstreamEmptySlot(ctx context.Context, epoch phase0.Epoch, slot state.Slot) error {
	if n.syncing.IsSyncing {
		return nil
	}

	if err := n.publishEmptySlot(ctx, slot.Number()); err != nil {
		return err
	}

	return nil
}

func (n *node) handleStateEpochSlotChanged(ctx context.Context, epochNumber phase0.Epoch, slot phase0.Slot) error {
	n.log.WithFields(logrus.Fields{
		"epoch": epochNumber,
		"slot":  slot,
	}).Debug("Current epoch/slot changed")

	for i := epochNumber; i < epochNumber+1; i++ {
		epoch, err := n.state.GetEpoch(ctx, i)
		if err != nil {
			return err
		}

		if epoch.HaveProposerDuties() {
			continue
		}

		if n.syncing.IsSyncing {
			continue
		}

		if err := n.fetchEpochProposerDuties(ctx, i); err != nil {
			return err
		}
	}

	return nil
}

func (n *node) fetchEpochProposerDuties(ctx context.Context, epoch phase0.Epoch) error {
	duties, err := n.getProserDuties(ctx, epoch)
	if err != nil {
		return err
	}

	if err := n.state.SetProposerDuties(ctx, epoch, duties); err != nil {
		return err
	}

	return nil
}

func (n *node) initializeState(ctx context.Context) error {
	n.log.Info("Initializing beacon state")

	sp, err := n.getSpec(ctx)
	if err != nil {
		return err
	}

	genesis, err := n.fetchGenesis(ctx)
	if err != nil {
		return err
	}

	st := state.NewContainer(ctx, n.log, sp, genesis)

	if err := st.Init(ctx); err != nil {
		return err
	}

	n.state = &st

	n.log.Info("Beacon state initialized! Ready to serve requests...")

	return nil
}

func (n *node) getSpec(ctx context.Context) (*state.Spec, error) {
	provider, isProvider := n.client.(eth2client.SpecProvider)
	if !isProvider {
		return nil, errors.New("client does not implement eth2client.SpecProvider")
	}

	data, err := provider.Spec(ctx)
	if err != nil {
		return nil, err
	}

	sp := state.NewSpec(data)

	if err := n.publishSpecUpdated(ctx, &sp); err != nil {
		return nil, err
	}

	return &sp, nil
}

func (n *node) getProserDuties(ctx context.Context, epoch phase0.Epoch) ([]*v1.ProposerDuty, error) {
	n.log.WithField("epoch", epoch).Debug("Fetching proposer duties")

	provider, isProvider := n.client.(eth2client.ProposerDutiesProvider)
	if !isProvider {
		return nil, errors.New("client does not implement eth2client.ProposerDutiesProvider")
	}

	duties, err := provider.ProposerDuties(ctx, epoch, nil)
	if err != nil {
		return nil, err
	}

	return duties, nil
}

func (n *node) getBlock(ctx context.Context, blockID string) (*spec.VersionedSignedBeaconBlock, error) {
	provider, isProvider := n.client.(eth2client.SignedBeaconBlockProvider)
	if !isProvider {
		return nil, errors.New("client does not implement eth2client.SignedBeaconBlockProvider")
	}

	signedBeaconBlock, err := provider.SignedBeaconBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	return signedBeaconBlock, nil
}
