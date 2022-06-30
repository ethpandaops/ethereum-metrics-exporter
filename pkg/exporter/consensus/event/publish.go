package event

import (
	"context"
)

func (p *DecoratedPublisher) PublishBeaconBlock(ctx context.Context, event *BeaconBlock) error {
	return p.conn.Publish(TopicBeaconBlock, event)
}
