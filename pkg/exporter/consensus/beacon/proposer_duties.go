package beacon

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

// func (n *Node) getProposerDutiesForEpoch(ctx context.Context, epoch phase0.Epoch) (ProposerDuties, error) {
// 	if b.proposerDutiesForEpoch[epoch] != nil {
// 		return b.proposerDutiesForEpoch[epoch], nil
// 	}

// 	provider, isProvider := b.client.(eth2client.ProposerDutiesProvider)
// 	if !isProvider {
// 		return nil, errors.New("client does not implement eth2client.DutyProvider")
// 	}

// 	duties, err := provider.ProposerDuties(ctx, epoch, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	b.proposerDutiesForEpoch[epoch] = NewProposerDuties(duties)

// 	return b.proposerDutiesForEpoch[epoch], nil
// }
