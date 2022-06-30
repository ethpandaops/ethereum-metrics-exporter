package state

import (
	"errors"
	"sync"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type Epochs struct {
	spec  *Spec
	state map[phase0.Epoch]*Epoch

	mu *sync.Mutex
}

func NewEpochs(spec *Spec) Epochs {
	return Epochs{
		spec:  spec,
		state: make(map[phase0.Epoch]*Epoch),

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

func (e *Epochs) NewInitializedEpoch(number phase0.Epoch) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	epoch := NewEpoch(e.spec.SlotsPerEpoch)

	return e.AddEpoch(number, &epoch)
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
