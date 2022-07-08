package beacon

import (
	"context"
	"errors"
	"fmt"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func (n *node) ensureBeaconSubscription(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 5):
			if n.client == nil {
				continue
			}

			if time.Since(n.lastEventTime) < (5 * time.Minute) {
				continue
			}

			n.log.
				WithField("last_event_time", n.lastEventTime.Local().String()).
				Info("Haven't received any events for 5 minutes, re-subscribing")

			if time.Since(n.lastEventTime) > time.Minute*5 {
				if err := n.subscribeToBeaconEvents(ctx); err != nil {
					n.log.WithError(err).Error("Failed to subscribe to beacon")
				}

				time.Sleep(time.Second * 60)
			}
		}
	}
}

func (n *node) subscribeToBeaconEvents(ctx context.Context) error {
	provider, isProvider := n.client.(eth2client.EventsProvider)
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
		n.lastEventTime = time.Now()

		if err := n.handleEvent(ctx, event); err != nil {
			n.log.Errorf("Failed to handle event: %v", err)
		}
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) handleEvent(ctx context.Context, event *v1.Event) error {
	if err := n.publishEvent(ctx, event); err != nil {
		n.log.WithError(err).Error("Failed to publish raw event")
	}

	switch event.Topic {
	case topicAttestation:
		return n.handleAttestation(ctx, event)
	case topicBlock:
		return n.handleBlock(ctx, event)
	case topicChainReorg:
		return n.handleChainReorg(ctx, event)
	case topicFinalizedCheckpoint:
		return n.handleFinalizedCheckpoint(ctx, event)
	case topicHead:
		return n.handleHead(ctx, event)
	case topicVoluntaryExit:
		return n.handleVoluntaryExit(ctx, event)
	case topicContributionAndProof:
		return n.handleContributionAndProof(ctx, event)

	default:
		return fmt.Errorf("unknown event topic %s", event.Topic)
	}
}

func (n *node) handleAttestation(ctx context.Context, event *v1.Event) error {
	attestation, valid := event.Data.(*phase0.Attestation)
	if !valid {
		return errors.New("invalid attestation event")
	}

	if err := n.publishAttestation(ctx, attestation); err != nil {
		return err
	}

	return nil
}

func (n *node) handleBlock(ctx context.Context, event *v1.Event) error {
	block, valid := event.Data.(*v1.BlockEvent)
	if !valid {
		return errors.New("invalid block event")
	}

	if err := n.publishBlock(ctx, block); err != nil {
		return err
	}

	return nil
}

func (n *node) handleChainReorg(ctx context.Context, event *v1.Event) error {
	chainReorg, valid := event.Data.(*v1.ChainReorgEvent)
	if !valid {
		return errors.New("invalid chain reorg event")
	}

	if err := n.publishChainReOrg(ctx, chainReorg); err != nil {
		return err
	}

	return nil
}

func (n *node) handleFinalizedCheckpoint(ctx context.Context, event *v1.Event) error {
	checkpoint, valid := event.Data.(*v1.FinalizedCheckpointEvent)
	if !valid {
		return errors.New("invalid checkpoint event")
	}

	if err := n.publishFinalizedCheckpoint(ctx, checkpoint); err != nil {
		return err
	}

	return nil
}

func (n *node) handleHead(ctx context.Context, event *v1.Event) error {
	head, valid := event.Data.(*v1.HeadEvent)
	if !valid {
		return errors.New("invalid head event")
	}

	if err := n.publishHead(ctx, head); err != nil {
		return err
	}

	return nil
}

func (n *node) handleVoluntaryExit(ctx context.Context, event *v1.Event) error {
	exit, valid := event.Data.(*phase0.VoluntaryExit)
	if !valid {
		return errors.New("invalid voluntary exit event")
	}

	if err := n.publishVoluntaryExit(ctx, exit); err != nil {
		return err
	}

	return nil
}

func (n *node) handleContributionAndProof(ctx context.Context, event *v1.Event) error {
	// Do nothing for now
	return nil
}
