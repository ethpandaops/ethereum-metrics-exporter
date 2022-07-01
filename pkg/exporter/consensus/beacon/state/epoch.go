package state

import (
	"errors"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type Epoch struct {
	slots     Slots
	Number    phase0.Epoch
	FirstSlot phase0.Slot
	LastSlot  phase0.Slot
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	bundle    BlockTimeCalculatorBundle
}

func NewEpoch(epochNumber phase0.Epoch, slotsPerEpoch phase0.Slot, bundle BlockTimeCalculatorBundle) Epoch {
	firstSlot := uint64(epochNumber) * uint64(slotsPerEpoch)
	lastSlot := (firstSlot + uint64(slotsPerEpoch)) - 1

	e := Epoch{
		slots: NewSlots(bundle),

		Number:    epochNumber,
		FirstSlot: phase0.Slot(firstSlot),
		LastSlot:  phase0.Slot(lastSlot),
		StartTime: bundle.Genesis.GenesisTime.Add(time.Duration(firstSlot) * bundle.SecondsPerSlot),
		EndTime:   bundle.Genesis.GenesisTime.Add((time.Duration(lastSlot) * bundle.SecondsPerSlot)).Add(bundle.SecondsPerSlot),
		Duration:  bundle.SecondsPerSlot * time.Duration(slotsPerEpoch),
		bundle:    bundle,
	}

	e.InitializeSlots(epochNumber, slotsPerEpoch)

	return e
}

func (e *Epoch) AddBlock(block *spec.VersionedSignedBeaconBlock) error {
	if block == nil {
		return errors.New("block is nil")
	}

	return e.slots.AddBlock(&TimedBlock{
		Block:  block,
		SeenAt: time.Now(),
	})
}

func (e *Epoch) GetSlotProposer(slot phase0.Slot) (*v1.ProposerDuty, error) {
	return e.slots.GetProposerDuty(slot)
}

func (e *Epoch) SetProposerDuties(duties []*v1.ProposerDuty) error {
	return e.slots.AddProposerDuties(duties)
}

func (e *Epoch) HasProposerDuties() bool {
	return len(e.slots.proposerDuties) > 0
}

func (e *Epoch) InitializeSlots(epoch phase0.Epoch, slots phase0.Slot) {
	start := uint64(e.FirstSlot)
	end := uint64(e.LastSlot)

	for i := start; i <= end; i++ {
		e.slots.AddEmptySlot(phase0.Slot(i))
	}
}
