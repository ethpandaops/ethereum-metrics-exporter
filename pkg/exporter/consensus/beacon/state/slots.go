package state

import (
	"errors"
	"sync"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type Slots struct {
	state  map[phase0.Slot]*Slot
	bundle BlockTimeCalculatorBundle
	mu     *sync.Mutex
}

func NewSlots(bundle BlockTimeCalculatorBundle) Slots {
	return Slots{
		state:  make(map[phase0.Slot]*Slot),
		mu:     &sync.Mutex{},
		bundle: bundle,
	}
}

func (m *Slots) Add(slot *Slot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if slot == nil {
		return errors.New("slot is nil")
	}

	m.state[slot.Number()] = slot

	return nil
}

func (m *Slots) Get(slot phase0.Slot) (*Slot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state[slot] == nil {
		return nil, errors.New("slot does not exist")
	}

	return m.state[slot], nil
}

func (m *Slots) Delete(slot phase0.Slot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.state, slot)

	return nil
}
