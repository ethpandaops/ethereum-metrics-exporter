package state

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func (c *Container) publishEpochChanged(ctx context.Context, epoch phase0.Epoch) {
	for _, cb := range c.callbacksEpochChanged {
		//nolint:errcheck // we dont care if the callback fails
		cb(ctx, epoch)
	}
}

func (c *Container) publishSlotChanged(ctx context.Context, slot phase0.Slot) {
	for _, cb := range c.callbacksSlotChanged {
		//nolint:errcheck // we dont care if the callback fails
		cb(ctx, slot)
	}
}

func (c *Container) publishEpochSlotChanged(ctx context.Context, epoch phase0.Epoch, slot phase0.Slot) {
	for _, cb := range c.callbacksEpochSlotChanged {
		//nolint:errcheck // we dont care if the callback fails
		cb(ctx, epoch, slot)
	}
}

func (c *Container) publishBlockInserted(ctx context.Context, epoch phase0.Epoch, slot Slot) {
	for _, cb := range c.callbacksBlockInserted {
		//nolint:errcheck // we dont care if the callback fails
		cb(ctx, epoch, slot)
	}
}
