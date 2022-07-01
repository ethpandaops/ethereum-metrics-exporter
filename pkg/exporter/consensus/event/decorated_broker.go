package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// DecordatedBroker is a broker that decorates the underlying broker with additional functionality - commonly
// enriching BeaconAPI events ith additional information.
type DecoratedPublisher struct {
	log logrus.FieldLogger

	conn   *nats.EncodedConn
	client eth2client.Service

	lastEventTime time.Time
}

// NewDecoratedPublisher creates a new DecoratedPublisher.
func NewDecoratedPublisher(ctx context.Context, log logrus.FieldLogger, client eth2client.Service, ec *nats.EncodedConn) *DecoratedPublisher {
	return &DecoratedPublisher{
		client: client,
		log:    log.WithField("module", "event/decorated_broker"),
		conn:   ec,
	}
}

func (p *DecoratedPublisher) StartListeningForEvents(ctx context.Context) {
	go func() {
		for {
			if time.Since(p.lastEventTime) > (5*time.Minute) && p.client != nil {
				p.log.
					WithField("last_event_time", p.lastEventTime.Local().String()).
					Info("Subscribing to upstream events")

				if err := p.startSubscriptions(ctx); err != nil {
					p.log.Errorf("Failed to subscribe to eth2 node: %v", err)
				}
			}

			time.Sleep(60 * time.Second)
		}
	}()
}

func (p *DecoratedPublisher) startSubscriptions(ctx context.Context) error {
	p.log.Info("starting subscriptions")

	provider, isProvider := p.client.(eth2client.EventsProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.Subscriptions")
	}

	topics := []string{}

	for key, supported := range v1.SupportedEventTopics {
		if key == "contribution_and_proof" {
			continue
		}

		if supported {
			topics = append(topics, key)
		}
	}

	if err := provider.Events(ctx, topics, func(event *v1.Event) {
		p.lastEventTime = time.Now()
		if err := p.handleEvent(ctx, event); err != nil {
			p.log.Errorf("Failed to handle event: %v", err)
		}
	}); err != nil {
		return err
	}

	return nil
}

func (p *DecoratedPublisher) handleEvent(ctx context.Context, event *v1.Event) error {
	switch event.Topic {
	case "block":
		return p.handleBlockEvent(ctx, event)
	case "attestation":
		// do nothing
	case "head":
		// do nothing
	default:
		return fmt.Errorf("unknown event: %s", event.Topic)
	}

	return nil
}

func (p *DecoratedPublisher) handleBlockEvent(ctx context.Context, event *v1.Event) error {
	blockEvent, ok := event.Data.(*v1.BlockEvent)
	if !ok {
		return errors.New("invalid block event")
	}

	// Create a decorated block event, enriched with our extra information.
	enriched := &BeaconBlock{
		RawEvent: blockEvent,
	}

	// Fetch the full block from the Beacon node.
	block, err := p.getBlock(ctx, fmt.Sprintf("%v", blockEvent.Slot))
	if err != nil {
		return err
	}

	enriched.Block = block

	// Publish the enriched block event.
	if err := p.PublishBeaconBlock(ctx, enriched); err != nil {
		return err
	}

	return nil
}

func (p *DecoratedPublisher) getBlock(ctx context.Context, blockID string) (*spec.VersionedSignedBeaconBlock, error) {
	provider, isProvider := p.client.(eth2client.SignedBeaconBlockProvider)
	if !isProvider {
		return nil, errors.New("client does not implement eth2client.SignedBeaconBlockProvider")
	}

	signedBeaconBlock, err := provider.SignedBeaconBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	return signedBeaconBlock, nil
}
