package state

import (
	"errors"
	"sync"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type Epochs struct {
	spec    *Spec
	state   map[phase0.Epoch]*Epoch
	genesis *v1.Genesis
	bundle  BlockTimeCalculatorBundle
	mu      *sync.Mutex
}

func NewEpochs(spec *Spec, genesis *v1.Genesis) Epochs {
	return Epochs{
		spec:    spec,
		state:   make(map[phase0.Epoch]*Epoch),
		genesis: genesis,
		bundle: BlockTimeCalculatorBundle{
			Genesis:        genesis,
			SecondsPerSlot: spec.SecondsPerSlot,
		},
		mu: &sync.Mutex{},
	}
}

func (e *Epochs) GetEpoch(epoch phase0.Epoch) (*Epoch, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state[epoch] == nil {
		return nil, errors.New("epoch not found")
	}

	return e.state[epoch], nil
}

func (e *Epochs) Exists(number phase0.Epoch) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.state[number]

	return ok
}

func (e *Epochs) NewInitializedEpoch(number phase0.Epoch) (*Epoch, error) {
	epoch := NewEpoch(number, e.spec.SlotsPerEpoch, e.bundle)

	if err := epoch.InitializeSlots(); err != nil {
		return nil, err
	}

	if err := e.AddEpoch(number, &epoch); err != nil {
		return nil, err
	}

	return &epoch, nil
}

func (e *Epochs) AddEpoch(number phase0.Epoch, epoch *Epoch) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if epoch == nil {
		return errors.New("epoch is nil")
	}

	e.state[number] = epoch

	return nil
}

func (e *Epochs) RemoveEpoch(number phase0.Epoch) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.state, number)

	return nil
}
