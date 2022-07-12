package state

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func (c *Container) handleCallbackError(err error, topic string) {
	if err != nil {
		c.log.WithError(err).WithField("topic", topic).Error("Receieved error from subscriber callback")
	}
}

func (c *Container) publishEpochChanged(ctx context.Context, epoch phase0.Epoch) {
	for _, cb := range c.callbacksEpochChanged {
		c.handleCallbackError(cb(ctx, epoch), "epochs_changed")
	}
}

func (c *Container) publishSlotChanged(ctx context.Context, slot phase0.Slot) {
	for _, cb := range c.callbacksSlotChanged {
		c.handleCallbackError(cb(ctx, slot), "slots_changed")
	}
}

func (c *Container) publishEpochSlotChanged(ctx context.Context, epoch phase0.Epoch, slot phase0.Slot) {
	for _, cb := range c.callbacksEpochSlotChanged {
		c.handleCallbackError(cb(ctx, epoch, slot), "epoch_slots_changed")
	}
}

func (c *Container) publishBlockInserted(ctx context.Context, epoch phase0.Epoch, slot Slot) {
	for _, cb := range c.callbacksBlockInserted {
		c.handleCallbackError(cb(ctx, epoch, slot), "block_inserted")
	}
}

func (c *Container) publishEmptySlot(ctx context.Context, epoch phase0.Epoch, slot Slot) {
	for _, cb := range c.callbacksEmptySlot {
		c.handleCallbackError(cb(ctx, epoch, slot), "empty_slot")
	}
}
