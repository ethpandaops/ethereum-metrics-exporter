package state

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type ProposerDuties = map[phase0.Slot]*v1.ProposerDuty

func NewProposerDuties(duties []*v1.ProposerDuty) ProposerDuties {
	out := make(ProposerDuties)

	for _, duty := range duties {
		out[duty.Slot] = duty
	}

	return out
}

type ProposerDutiesForEpoch = map[phase0.Epoch]ProposerDuties
