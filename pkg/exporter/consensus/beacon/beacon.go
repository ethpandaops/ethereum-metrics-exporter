package beacon

import (
	"context"
	"errors"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon/state"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/event"
	"github.com/sirupsen/logrus"
)

// Node represents an Ethereum beacon node. It computes values based on the spec.
type Node struct {
	// Helpers
	log logrus.FieldLogger

	// Clients
	api    api.ConsensusClient
	client eth2client.Service
	events *event.DecoratedPublisher

	// Internal data stores
	genesis *v1.Genesis
	state   *state.Container

	currentEpoch phase0.Epoch
}

func NewNode(ctx context.Context, log logrus.FieldLogger, ap api.ConsensusClient, client eth2client.Service, events *event.DecoratedPublisher) *Node {
	return &Node{
		log:    log,
		api:    ap,
		client: client,
		events: events,

		currentEpoch: phase0.Epoch(0),
	}
}

func (n *Node) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 1):
			n.tick(ctx)
		}
	}
}

func (n *Node) StartAsync(ctx context.Context) {
	go func() {
		if err := n.Start(ctx); err != nil {
			n.log.WithError(err).Error("Failed to start beacon node")
		}
	}()
}

func (n *Node) tick(ctx context.Context) {
	if n.state == nil {
		if err := n.InitializeState(ctx); err != nil {
			n.log.WithError(err).Error("Failed to initialize state")
		}
	} else {
		if err := n.checkForNewEpoch(ctx); err != nil {
			n.log.WithError(err).Error("Failed to check for new epoch")
		}
	}
}

func (n *Node) checkForNewEpoch(ctx context.Context) error {
	currentEpoch, err := n.state.CurrentEpoch()
	if err != nil {
		return err
	}

	if n.currentEpoch != currentEpoch {
		n.log.WithField("epoch", currentEpoch).Info("Epoch changed")

		for i := currentEpoch; i < currentEpoch+1; i++ {
			if err := n.fetchEpochProposerDuties(ctx, phase0.Epoch(i)); err != nil {
				return err
			}
		}

		n.currentEpoch = currentEpoch
	}

	return nil
}

func (n *Node) fetchEpochProposerDuties(ctx context.Context, epoch phase0.Epoch) error {
	n.log.WithField("epoch", epoch).Info("Fetching proposer duties")

	duties, err := n.GetProserDuties(ctx, epoch)
	if err != nil {
		return err
	}

	if err := n.state.SetProposerDuties(ctx, epoch, duties); err != nil {
		return err
	}

	return nil
}

func (n *Node) InitializeState(ctx context.Context) error {
	n.log.Info("Initializing beacon state")

	spec, err := n.GetSpec(ctx)
	if err != nil {
		return err
	}

	genesis, err := n.GetGenesis(ctx)
	if err != nil {
		return err
	}

	st := state.NewContainer(ctx, n.log, spec, genesis, n.events)

	if err := st.Init(ctx); err != nil {
		return err
	}

	n.state = &st

	return nil
}

func (n *Node) GetSpec(ctx context.Context) (*state.Spec, error) {
	provider, isProvider := n.client.(eth2client.SpecProvider)
	if !isProvider {
		return nil, errors.New("client does not implement eth2client.SpecProvider")
	}

	data, err := provider.Spec(ctx)
	if err != nil {
		return nil, err
	}

	spec := state.NewSpec(data)

	return &spec, nil
}

func (n *Node) GetProserDuties(ctx context.Context, epoch phase0.Epoch) ([]*v1.ProposerDuty, error) {
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
