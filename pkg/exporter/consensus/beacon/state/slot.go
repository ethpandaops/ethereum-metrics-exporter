package state

import (
	"errors"
	"sync"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type Slots struct {
	blocks         map[phase0.Slot]*TimedBlock
	proposerDuties map[phase0.Slot]*v1.ProposerDuty
	bundle         BlockTimeCalculatorBundle
	mu             *sync.Mutex
}

func NewSlots(bundle BlockTimeCalculatorBundle) Slots {
	return Slots{
		blocks:         make(map[phase0.Slot]*TimedBlock),
		proposerDuties: make(map[phase0.Slot]*v1.ProposerDuty),
		mu:             &sync.Mutex{},
		bundle:         bundle,
	}
}

func (m *Slots) GetBlockAtSlot(slot phase0.Slot) (*TimedBlock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.blocks[slot] == nil {
		return nil, errors.New("no block at slot")
	}

	return m.blocks[slot], nil
}

func (m *Slots) AddBlock(timedBlock *TimedBlock) error {
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

func (m *Slots) AddEmptySlot(slot phase0.Slot) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[slot] = nil
}

func (m *Slots) GetSlotProposerDelay(slot phase0.Slot) (time.Duration, error) {
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

func (m *Slots) GetProposerDuty(slot phase0.Slot) (*v1.ProposerDuty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.proposerDuties[slot] == nil {
		return nil, errors.New("no proposer duty at slot")
	}

	return m.proposerDuties[slot], nil
}

func (m *Slots) AddProposerDuties(proposerDuties []*v1.ProposerDuty) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, proposerDuty := range proposerDuties {
		m.proposerDuties[proposerDuty.Slot] = proposerDuty
	}

	return nil
}
