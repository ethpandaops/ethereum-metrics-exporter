package state

import (
	"errors"
	"sync"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// Slot is a slot in the beacon chain.
type Slot struct {
	block        *TimedBlock
	proposerDuty *v1.ProposerDuty
	number       phase0.Slot
	bundle       BlockTimeCalculatorBundle
	mu           *sync.Mutex
}

// NewSlot returns a new slot.
func NewSlot(number phase0.Slot, bundle BlockTimeCalculatorBundle) Slot {
	return Slot{
		block:        nil,
		proposerDuty: nil,
		number:       number,
		mu:           &sync.Mutex{},
		bundle:       bundle,
	}
}

// Number returns the slot number.
func (m *Slot) Number() phase0.Slot {
	return m.number
}

// Block returns the block for the slot (if it exists).
func (m *Slot) Block() (*TimedBlock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.block == nil {
		return nil, errors.New("block does not exist")
	}

	return m.block, nil
}

// AddBlock adds a block to the slot.
func (m *Slot) AddBlock(timedBlock *TimedBlock) error {
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

	if slot != m.number {
		return errors.New("block slot does not match slot")
	}

	m.block = timedBlock

	return nil
}

// ProposerDelay calculates the amount of time it took for the proposer to publish the block.
func (m *Slot) ProposerDelay() (time.Duration, error) {
	if m.block == nil {
		return 0, errors.New("block does not exist")
	}

	// Calculate the proposer delay for the block.
	expected := m.bundle.Genesis.GenesisTime.Add(time.Duration(uint64(m.number)) * m.bundle.SecondsPerSlot)
	delay := m.block.SeenAt.Sub(expected)

	return delay, nil
}

// ProposerDuty returns the proposer duty for the slot (if it exists).
func (m *Slot) ProposerDuty() (*v1.ProposerDuty, error) {
	if m.proposerDuty == nil {
		return nil, errors.New("proposer duty does not exist")
	}

	return m.proposerDuty, nil
}

// SetProposerDuty sets the proposer duty for the slot.
func (m *Slot) SetProposerDuty(proposerDuty *v1.ProposerDuty) error {
	if proposerDuty.Slot != m.number {
		return errors.New("proposer duty slot does not match slot")
	}

	m.proposerDuty = proposerDuty

	return nil
}

func (m *Slot) MissingBlock() bool {
	return m.block == nil
}
