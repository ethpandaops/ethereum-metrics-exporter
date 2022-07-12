package state

import (
	"errors"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// ForkEpoch is a beacon fork that activates at a specific epoch.
type ForkEpoch struct {
	Epoch phase0.Epoch
	Name  string
}

// Active returns true if the fork is active at the given slot.
func (f *ForkEpoch) Active(slot, slotsPerEpoch phase0.Slot) bool {
	return phase0.Epoch(int(slot)/int(slotsPerEpoch)) > f.Epoch
}

// ForkEpochs is a list of forks that activate at specific epochs.
type ForkEpochs []ForkEpoch

// Active returns a list of forks that are active at the given slot.
func (f *ForkEpochs) Active(slot, slotsPerEpoch phase0.Slot) []ForkEpoch {
	activated := []ForkEpoch{}

	for _, fork := range *f {
		if fork.Active(slot, slotsPerEpoch) {
			activated = append(activated, fork)
		}
	}

	return activated
}

// CurrentFork returns the current fork at the given slot.
func (f *ForkEpochs) Scheduled(slot, slotsPerEpoch phase0.Slot) []ForkEpoch {
	scheduled := []ForkEpoch{}

	for _, fork := range *f {
		if !fork.Active(slot, slotsPerEpoch) {
			scheduled = append(scheduled, fork)
		}
	}

	return scheduled
}

// CurrentFork returns the current fork at the given slot.
func (f *ForkEpochs) CurrentFork(slot, slotsPerEpoch phase0.Slot) (ForkEpoch, error) {
	largest := ForkEpoch{
		Epoch: 0,
	}

	for _, fork := range f.Active(slot, slotsPerEpoch) {
		if fork.Active(slot, slotsPerEpoch) && fork.Epoch > largest.Epoch {
			largest = fork
		}
	}

	if largest.Epoch == 0 {
		return ForkEpoch{}, errors.New("no active fork")
	}

	return largest, nil
}
