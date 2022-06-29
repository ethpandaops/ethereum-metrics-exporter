package beacon

import (
	"context"
	"errors"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus"
)

var (
	ErrSpecNotInitialized = errors.New("spec not initialized")
	ErrGenesisNotFetched  = errors.New("genesis not fetched")
)

func (n *Node) CurrentSlot(ctx context.Context) (phase0.Slot, phase0.Epoch, error) {
	if n.spec == nil {
		return 0, 0, ErrSpecNotInitialized
	}

	if n.genesis == nil {
		return 0, 0, ErrGenesisNotFetched
	}

	if err := n.spec.Validate(); err != nil {
		return 0, 0, err
	}

	// Calculate the current epoch based on genesis time.
	genesis := n.genesis.GenesisTime

	currentSlot := phase0.Slot(time.Since(genesis).Seconds() / n.spec.SecondsPerSlot.Seconds())
	currentEpoch := phase0.Epoch(currentSlot / n.spec.SlotsPerEpoch)

	n.log.WithFields(logrus.Fields{
		"slot":  currentSlot,
		"epoch": currentEpoch,
	}).Info("Calculated current epoch and slot")

	return currentSlot, currentEpoch, nil
}
