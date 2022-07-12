package state

import (
	"errors"
	"sync"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// Slots is a collection of slots.
type Slots struct {
	state  map[phase0.Slot]*Slot
	bundle BlockTimeCalculatorBundle
	mu     *sync.Mutex
}

// NewSlots returns a new slots instance.
func NewSlots(bundle BlockTimeCalculatorBundle) Slots {
	return Slots{
		state:  make(map[phase0.Slot]*Slot),
		mu:     &sync.Mutex{},
		bundle: bundle,
	}
}

// Add adds a slot to the collection.
func (m *Slots) Add(slot *Slot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if slot == nil {
		return errors.New("slot is nil")
	}

	m.state[slot.Number()] = slot

	return nil
}

// Get returns a slot from the collection.
func (m *Slots) Get(slot phase0.Slot) (*Slot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state[slot] == nil {
		return nil, errors.New("slot does not exist")
	}

	return m.state[slot], nil
}

// Delete deletes a slot from the collection.
func (m *Slots) Delete(slot phase0.Slot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.state, slot)

	return nil
}
