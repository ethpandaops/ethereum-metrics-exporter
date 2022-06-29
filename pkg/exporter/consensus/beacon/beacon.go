package beacon

import (
	"context"
	"errors"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/davecgh/go-spew/spew"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/sirupsen/logrus"
)

// Node represents an Ethereum beacon node. It computes values based on the spec.
type Node struct {
	// Helpers
	log logrus.FieldLogger

	// Clients
	api    api.ConsensusClient
	client eth2client.Service

	// Internal data stores
	spec    *Spec
	genesis *v1.Genesis

	// Misc
	specFetchedAt time.Time
}

func NewNode(ctx context.Context, log logrus.FieldLogger, ap api.ConsensusClient, client eth2client.Service) *Node {
	return &Node{
		log:    log,
		api:    ap,
		client: client,
	}
}

func (n *Node) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 5):
			n.tick(ctx)
		}
	}
}

func (n *Node) tick(ctx context.Context) {
	if time.Since(n.specFetchedAt) > 15*time.Minute {
		if err := n.fetchSpec(ctx); err != nil {
			n.log.Errorf("failed to fetch spec: %v", err)
		}

		if _, err := n.GetGenesis(ctx); err != nil {
			n.log.Errorf("failed to fetch genesis: %v", err)
		}
	}

	if _, _, err := n.CurrentSlot(ctx); err != nil {
		n.log.Errorf("failed to get current slot: %v", err)
	}
}

func (n *Node) fetchSpec(ctx context.Context) error {
	provider, isProvider := n.client.(eth2client.SpecProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.SpecProvider")
	}

	data, err := provider.Spec(ctx)
	if err != nil {
		return err
	}

	spec := NewSpec(data)

	n.spec = &spec

	n.specFetchedAt = time.Now()

	spew.Dump(spec)

	return nil
}
