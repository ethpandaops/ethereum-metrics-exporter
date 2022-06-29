package beacon

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/nats-io/nats.go"
)

const (
	TopicComputedEpochUpdated = "computed_epoch_updated"
	TopicComputedSlotUpdated  = "computed_slot_updated"
)

type ComputedEpochUpdatedEvent struct {
	Epoch phase0.Epoch
}

type ComputedSlotUpdatedEvent struct {
	Slot  phase0.Slot
	Epoch phase0.Epoch
}

func OnComputedEpochUpdated(ctx context.Context, broker *nats.EncodedConn, handler func(ctx context.Context, event *ComputedEpochUpdatedEvent) error) (*nats.Subscription, error) {
	return broker.Subscribe(TopicComputedEpochUpdated, func(event *ComputedEpochUpdatedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}

func OnComputedSlotUpdated(ctx context.Context, broker *nats.EncodedConn, handler func(ctx context.Context, event *ComputedSlotUpdatedEvent) error) (*nats.Subscription, error) {
	return broker.Subscribe(TopicComputedSlotUpdated, func(event *ComputedSlotUpdatedEvent) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}
