package state

import (
	"errors"
	"sync"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type MapOfSlotToBlock struct {
	blocks map[phase0.Slot]*TimedBlock

	mu *sync.Mutex
}

func NewMapOfSlotToBlock() MapOfSlotToBlock {
	return MapOfSlotToBlock{
		blocks: make(map[phase0.Slot]*TimedBlock),
		mu:     &sync.Mutex{},
	}
}

func (m *MapOfSlotToBlock) GetBlockAtSlot(slot phase0.Slot) (*TimedBlock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.blocks[slot] == nil {
		return nil, errors.New("no block at slot")
	}

	return m.blocks[slot], nil
}

func (m *MapOfSlotToBlock) AddBlock(timedBlock *TimedBlock) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if timedBlock == nil {
		return errors.New("timed_block is nil")
	}

	if timedBlock.Block == nil {
		return errors.New("block is nil")
	}

	slot, err := timedBlock.Block.Slot()
	if err != nil {
		return err
	}

	m.blocks[slot] = timedBlock

	return nil
}

func (m *MapOfSlotToBlock) InitializeSlots(slots phase0.Slot) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := uint64(0); i < uint64(slots); i++ {
		m.blocks[phase0.Slot(i)] = nil
	}
}
