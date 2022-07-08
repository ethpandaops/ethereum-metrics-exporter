package state

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// OnEpochChanged is called when the current epoch changes.
func (c *Container) OnEpochChanged(ctx context.Context, cb func(ctx context.Context, epoch phase0.Epoch) error) error {
	c.callbacksEpochChanged = append(c.callbacksEpochChanged, cb)

	return nil
}

// OnSlotChanged is called when the current slot changes.
func (c *Container) OnSlotChanged(ctx context.Context, cb func(ctx context.Context, slot phase0.Slot) error) error {
	c.callbacksSlotChanged = append(c.callbacksSlotChanged, cb)

	return nil
}

// OnEpochSlotChanged is called when the current epoch or slot changes.
func (c *Container) OnEpochSlotChanged(ctx context.Context, cb func(ctx context.Context, epoch phase0.Epoch, slot phase0.Slot) error) error {
	c.callbacksEpochSlotChanged = append(c.callbacksEpochSlotChanged, cb)

	return nil
}

// OnBlockInserted is called when a block is inserted in to a slot.
func (c *Container) OnBlockInserted(ctx context.Context, cb func(ctx context.Context, epoch phase0.Epoch, slot Slot) error) error {
	c.callbacksBlockInserted = append(c.callbacksBlockInserted, cb)

	return nil
}

// OnEmptySlot is called when a slot expires without an associated block.
func (c *Container) OnEmptySlot(ctx context.Context, cb func(ctx context.Context, epoch phase0.Epoch, slot Slot) error) error {
	c.callbacksEmptySlot = append(c.callbacksEmptySlot, cb)

	return nil
}
