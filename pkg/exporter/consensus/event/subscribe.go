package event

import (
	"context"

	"github.com/nats-io/nats.go"
)

func (p *DecoratedPublisher) OnBeaconBlock(ctx context.Context, handler func(ctx context.Context, event *BeaconBlock) error) (*nats.Subscription, error) {
	return p.conn.Subscribe(TopicBeaconBlock, func(event *BeaconBlock) {
		//nolint:errcheck // safe to ignore)
		handler(ctx, event)
	})
}
