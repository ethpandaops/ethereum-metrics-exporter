package state

import (
	"context"
	"errors"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/nats-io/nats.go"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/event"
	"github.com/sirupsen/logrus"
)

type Container struct {
	log           logrus.FieldLogger
	events        *event.DecoratedPublisher
	Spec          *Spec
	genesis       *v1.Genesis
	Epochs        Epochs
	subscriptions []*nats.Subscription
}

const (
	SURROUNDING_EPOCH_DISTANCE = 1
)

func NewContainer(ctx context.Context, log logrus.FieldLogger, spec *Spec, genesis *v1.Genesis, events *event.DecoratedPublisher) Container {
	return Container{
		log:    log,
		Spec:   spec,
		events: events,

		genesis: genesis,

		Epochs: NewEpochs(spec, genesis),
	}
}

var (
	ErrSpecNotInitialized = errors.New("spec not initialized")
	ErrGenesisNotFetched  = errors.New("genesis not fetched")
)

func (c *Container) Init(ctx context.Context) error {
	// Create event listeners
	subscription, err := c.events.OnBeaconBlock(ctx, c.handleBeaconBlockEvent)
	if err != nil {
		return err
	}

	c.subscriptions = append(c.subscriptions, subscription)

	go c.ticker(ctx)

	return nil
}

func (c *Container) ticker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 1):
			c.tick(ctx)
		}
	}
}

func (c *Container) tick(ctx context.Context) {
	if err := c.hydrateEpochs(); err != nil {
		c.log.WithError(err).Error("Failed to hydrate epochs")
	}
}

func (c *Container) handleBeaconBlockEvent(ctx context.Context, ev *event.BeaconBlock) error {
	c.log.WithField("slot", ev.RawEvent.Slot).Info("Handling beacon block event")

	if err := c.insertBeaconBlock(ctx, ev.Block); err != nil {
		return err
	}

	return nil
}

func (c *Container) insertBeaconBlock(ctx context.Context, beaconBlock *spec.VersionedSignedBeaconBlock) error {
	// Calculate the epoch
	slot, err := beaconBlock.Slot()
	if err != nil {
		return err
	}

	epochNumber := phase0.Epoch(slot / c.Spec.SlotsPerEpoch)

	if exists := c.Epochs.Exists(epochNumber); !exists {
		if err := c.Epochs.NewInitializedEpoch(epochNumber); err != nil {
			return err
		}
	}

	// Get the epoch
	epoch, err := c.Epochs.GetEpoch(epochNumber)
	if err != nil {
		return err
	}

	// Insert the block
	if err := epoch.AddBlock(beaconBlock); err != nil {
		return err
	}

	delay, err := epoch.Blocks.GetSlotProposerDelay(slot)
	if err != nil {
		return err
	}

	c.log.WithFields(logrus.Fields{
		"epoch":          epochNumber,
		"slot":           slot,
		"proposer_delay": delay.String(),
	}).Info("Inserted beacon block")

	return nil
}

func (c *Container) hydrateEpochs() error {
	epoch, _, err := c.GetCurrentEpochAndSlot()
	if err != nil {
		return err
	}

	// Ensure the state has +-SURROUNDING_EPOCH_DISTANCE epochs created.
	for i := epoch - SURROUNDING_EPOCH_DISTANCE; i <= epoch+SURROUNDING_EPOCH_DISTANCE; i++ {
		if _, err := c.Epochs.GetEpoch(i); err != nil {
			if err := c.Epochs.NewInitializedEpoch(i); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Container) CurrentEpoch() (phase0.Epoch, error) {
	epoch, _, err := c.GetCurrentEpochAndSlot()
	if err != nil {
		return 0, err
	}

	return epoch, nil
}

func (c *Container) CurrentSlot() (phase0.Slot, error) {
	_, slot, err := c.GetCurrentEpochAndSlot()
	if err != nil {
		return 0, err
	}

	return slot, nil
}

func (c *Container) GetCurrentEpochAndSlot() (phase0.Epoch, phase0.Slot, error) {
	if c.Spec == nil {
		return 0, 0, ErrSpecNotInitialized
	}

	if c.genesis == nil {
		return 0, 0, ErrGenesisNotFetched
	}

	if err := c.Spec.Validate(); err != nil {
		return 0, 0, err
	}

	// Calculate the current epoch based on genesis time.
	genesis := c.genesis.GenesisTime

	currentSlot := phase0.Slot(time.Since(genesis).Seconds() / c.Spec.SecondsPerSlot.Seconds())
	currentEpoch := phase0.Epoch(currentSlot / c.Spec.SlotsPerEpoch)

	return currentEpoch, currentSlot, nil
}
