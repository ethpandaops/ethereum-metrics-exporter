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
	Start(ctx context.Context) error
	StartAsync(ctx context.Context)

	// Getters
	GetEpoch(ctx context.Context, epoch phase0.Epoch) (*state.Epoch, error)
	GetSlot(ctx context.Context, slot phase0.Slot) (*state.Slot, error)
	GetSpec(ctx context.Context) (*state.Spec, error)
	GetSyncState(ctx context.Context) (*v1.SyncState, error)

	// Subscriptions
	// - Proxied Beacon events
	OnEvent(ctx context.Context, handler func(ctx context.Context, ev *v1.Event) error) (*nats.Subscription, error)
	OnBlock(ctx context.Context, handler func(ctx context.Context, ev *v1.BlockEvent) error) (*nats.Subscription, error)
	OnAttestation(ctx context.Context, handler func(ctx context.Context, ev *phase0.Attestation) error) (*nats.Subscription, error)
	OnFinalizedCheckpoint(ctx context.Context, handler func(ctx context.Context, ev *v1.FinalizedCheckpointEvent) error) (*nats.Subscription, error)
	OnHead(ctx context.Context, handler func(ctx context.Context, ev *v1.HeadEvent) error) (*nats.Subscription, error)
	OnChainReOrg(ctx context.Context, handler func(ctx context.Context, ev *v1.ChainReorgEvent) error) (*nats.Subscription, error)
	OnVoluntaryExit(ctx context.Context, handler func(ctx context.Context, ev *phase0.VoluntaryExit) error) (*nats.Subscription, error)

	// - Custom events
	OnReady(ctx context.Context, handler func(ctx context.Context, event *ReadyEvent) error) (*nats.Subscription, error)
	OnEpochChanged(ctx context.Context, handler func(ctx context.Context, event *EpochChangedEvent) error) (*nats.Subscription, error)
	OnSlotChanged(ctx context.Context, handler func(ctx context.Context, event *SlotChangedEvent) error) (*nats.Subscription, error)
	OnEpochSlotChanged(ctx context.Context, handler func(ctx context.Context, event *EpochSlotChangedEvent) error) (*nats.Subscription, error)
	OnBlockInserted(ctx context.Context, handler func(ctx context.Context, event *BlockInsertedEvent) error) (*nats.Subscription, error)
	OnSyncStatus(ctx context.Context, handler func(ctx context.Context, event *SyncStatusEvent) error) (*nats.Subscription, error)
	OnNodeVersionUpdated(ctx context.Context, handler func(ctx context.Context, event *NodeVersionUpdatedEvent) error) (*nats.Subscription, error)
	OnPeersUpdated(ctx context.Context, handler func(ctx context.Context, event *PeersUpdatedEvent) error) (*nats.Subscription, error)
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
		log:    log,
		api:    ap,
		client: client,
		broker: broker,

		syncing: &v1.SyncState{
			IsSyncing: false,
		},
	}
}

func (n *node) Start(ctx context.Context) error {
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

	if _, err := s.Every("1s").Do(func() {
		n.tick(ctx)
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
	return n.state.GetEpoch(ctx, epoch)
}

func (n *node) GetSlot(ctx context.Context, slot phase0.Slot) (*state.Slot, error) {
	return n.state.GetSlot(ctx, slot)
}

func (n *node) GetSpec(ctx context.Context) (*state.Spec, error) {
	sp := n.state.Spec()

	if sp == nil {
		return nil, errors.New("spec not yet available")
	}

	return sp, nil
}

func (n *node) GetSyncState(ctx context.Context) (*v1.SyncState, error) {
	return n.syncing, nil
}

func (n *node) tick(ctx context.Context) {
	if n.state == nil {
		if err := n.initializeState(ctx); err != nil {
			n.log.WithError(err).Error("Failed to initialize state")
		}

		if err := n.subscribeDownstream(ctx); err != nil {
			n.log.WithError(err).Error("Failed to subscribe to downstream")
		}

		if err := n.subscribeToSelf(ctx); err != nil {
			n.log.WithError(err).Error("Failed to subscribe to self")
		}

		//nolint:errcheck // we dont care if this errors out since it runs indefinitely in a goroutine
		go n.ensureBeaconSubscription(ctx)

		if err := n.publishReady(ctx); err != nil {
			n.log.WithError(err).Error("Failed to publish ready")
		}
	}
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
		start := time.Now()

		// Grab the entire block from the beacon node
		block, err := n.getBlock(ctx, fmt.Sprintf("%v", ev.Slot))
		if err != nil {
			return err
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
	if err := n.state.OnEpochChanged(ctx, n.handleStateEpochChanged); err != nil {
		return err
	}

	if err := n.state.OnBlockInserted(ctx, n.handleDownstreamBlockInserted); err != nil {
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

func (n *node) handleStateEpochChanged(ctx context.Context, epoch phase0.Epoch) error {
	n.log.WithFields(logrus.Fields{
		"epoch": epoch,
	}).Info("Current epoch changed")

	for i := epoch; i < epoch+1; i++ {
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

	genesis, err := n.GetGenesis(ctx)
	if err != nil {
		return err
	}

	st := state.NewContainer(ctx, n.log, sp, genesis)

	if err := st.Init(ctx); err != nil {
		return err
	}

	n.state = &st

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

	return &sp, nil
}

func (n *node) getProserDuties(ctx context.Context, epoch phase0.Epoch) ([]*v1.ProposerDuty, error) {
	n.log.WithField("epoch", epoch).Info("Fetching proposer duties")

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
