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
}

func NewEpoch(slotsPerEpoch phase0.Slot) Epoch {
	e := Epoch{
		ProposerDuties: make(ProposerDuties),
		Blocks:         NewMapOfSlotToBlock(),
	}

	e.Blocks.InitializeSlots(slotsPerEpoch)

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
