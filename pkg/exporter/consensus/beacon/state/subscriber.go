package state

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func (c *Container) OnEpochChanged(ctx context.Context, cb func(ctx context.Context, epoch phase0.Epoch) error) error {
	c.callbacksEpochChanged = append(c.callbacksEpochChanged, cb)

	return nil
}

func (c *Container) OnSlotChanged(ctx context.Context, cb func(ctx context.Context, slot phase0.Slot) error) error {
	c.callbacksSlotChanged = append(c.callbacksSlotChanged, cb)

	return nil
}

func (c *Container) OnEpochSlotChanged(ctx context.Context, cb func(ctx context.Context, epoch phase0.Epoch, slot phase0.Slot) error) error {
	c.callbacksEpochSlotChanged = append(c.callbacksEpochSlotChanged, cb)

	return nil
}

func (c *Container) OnBlockInserted(ctx context.Context, cb func(ctx context.Context, epoch phase0.Epoch, slot Slot) error) error {
	c.callbacksBlockInserted = append(c.callbacksBlockInserted, cb)

	return nil
}
