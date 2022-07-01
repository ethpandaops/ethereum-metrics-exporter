package state

import (
	"errors"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type Epoch struct {
	ProposerDuties ProposerDuties
	Blocks         MapOfSlotToBlock
	Number         phase0.Epoch
	Start          phase0.Slot
	End            phase0.Slot
	bundle         BlockTimeCalculatorBundle
}

func NewEpoch(epochNumber phase0.Epoch, slotsPerEpoch phase0.Slot, bundle BlockTimeCalculatorBundle) Epoch {
	start := uint64(epochNumber) * uint64(slotsPerEpoch)
	end := (start + uint64(slotsPerEpoch)) - 1

	e := Epoch{
		Number:         epochNumber,
		ProposerDuties: make(ProposerDuties),
		Blocks:         NewMapOfSlotToBlock(bundle),
		Start:          phase0.Slot(start),
		End:            phase0.Slot(end),
		bundle:         bundle,
	}

	e.InitializeSlots(epochNumber, slotsPerEpoch)

	return e
}

func (e *Epoch) SetProposerDuties(duties []*v1.ProposerDuty) {
	for _, duty := range duties {
		e.ProposerDuties[duty.Slot] = duty
	}
}

func (e *Epoch) AddBlock(block *spec.VersionedSignedBeaconBlock) error {
	if block == nil {
		return errors.New("block is nil")
	}

	return e.Blocks.AddBlock(&TimedBlock{
		Block:  block,
		SeenAt: time.Now(),
	})
}

func (e *Epoch) GetProserDutyAtSlot(slot phase0.Slot) (*v1.ProposerDuty, error) {
	if e.ProposerDuties[slot] == nil {
		return nil, errors.New("no proposer duty at slot")
	}

	return e.ProposerDuties[slot], nil
}

func (e *Epoch) InitializeSlots(epoch phase0.Epoch, slots phase0.Slot) {
	start := uint64(e.Start)
	end := uint64(e.End)

	for i := start; i <= end; i++ {
		e.Blocks.AddEmptySlot(phase0.Slot(i))
	}
}
