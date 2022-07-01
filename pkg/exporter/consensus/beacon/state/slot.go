package state

import (
	"errors"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type MapOfSlotToBlock struct {
	blocks map[phase0.Slot]*TimedBlock
	bundle BlockTimeCalculatorBundle
	mu     *sync.Mutex
}

func NewMapOfSlotToBlock(bundle BlockTimeCalculatorBundle) MapOfSlotToBlock {
	return MapOfSlotToBlock{
		blocks: make(map[phase0.Slot]*TimedBlock),
		mu:     &sync.Mutex{},
		bundle: bundle,
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

func (m *MapOfSlotToBlock) AddEmptySlot(slot phase0.Slot) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[slot] = nil
}

func (m *MapOfSlotToBlock) GetSlotProposerDelay(slot phase0.Slot) (time.Duration, error) {
	block, err := m.GetBlockAtSlot(slot)
	if err != nil {
		return 0, err
	}

	// Calculate the proposer delay for the block.
	// A negative delay means the block was proposed in-time.
	expected := m.bundle.Genesis.GenesisTime.Add(time.Duration(uint64(slot)) * m.bundle.SecondsPerSlot)
	delay := expected.Sub(block.SeenAt)

	return delay, nil
}
