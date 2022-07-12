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

	haveProposerDuties bool
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

		haveProposerDuties: false,
	}

	return e
}

func (e *Epoch) AddBlock(block *spec.VersionedSignedBeaconBlock, seenAt time.Time) error {
	if block == nil {
		return errors.New("block is nil")
	}

	slotNumber, err := block.Slot()
	if err != nil {
		return err
	}

	slot, err := e.slots.Get(slotNumber)
	if err != nil {
		return err
	}

	return slot.AddBlock(&TimedBlock{
		Block:  block,
		SeenAt: seenAt,
	})
}

func (e *Epoch) GetSlotProposer(slotNumber phase0.Slot) (*v1.ProposerDuty, error) {
	slot, err := e.slots.Get(slotNumber)
	if err != nil {
		return nil, err
	}

	return slot.ProposerDuty()
}

func (e *Epoch) SetProposerDuties(duties []*v1.ProposerDuty) error {
	for _, duty := range duties {
		slot, err := e.slots.Get(duty.Slot)
		if err != nil {
			return err
		}

		if err := slot.SetProposerDuty(duty); err != nil {
			return err
		}
	}

	e.haveProposerDuties = true

	return nil
}

func (e *Epoch) HaveProposerDuties() bool {
	return e.haveProposerDuties
}

func (e *Epoch) GetSlot(slotNumber phase0.Slot) (*Slot, error) {
	return e.slots.Get(slotNumber)
}

func (e *Epoch) InitializeSlots() error {
	start := uint64(e.FirstSlot)
	end := uint64(e.LastSlot)

	for i := start; i <= end; i++ {
		slot := NewSlot(phase0.Slot(i), e.bundle)

		if err := e.slots.Add(&slot); err != nil {
			return err
		}
	}

	return nil
}
