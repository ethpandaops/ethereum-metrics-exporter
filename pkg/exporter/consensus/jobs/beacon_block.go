package jobs

import (
	"github.com/attestantio/go-eth2-client/spec"
)

type BeaconBlock struct {
	AttesterSlashings int
	ProposerSlashings int
	Transactions      int
	Deposits          int
	VoluntaryExits    int
	Attestations      int
	Slot              uint64
}

func NewBeaconBlockFromPhase0(block *spec.VersionedSignedBeaconBlock) BeaconBlock {
	return BeaconBlock{
		AttesterSlashings: len(block.Phase0.Message.Body.AttesterSlashings),
		ProposerSlashings: len(block.Phase0.Message.Body.ProposerSlashings),
		Transactions:      0,
		Deposits:          len(block.Phase0.Message.Body.Deposits),
		VoluntaryExits:    len(block.Phase0.Message.Body.VoluntaryExits),
		Attestations:      len(block.Phase0.Message.Body.Attestations),
		Slot:              uint64(block.Phase0.Message.Slot),
	}
}

func NewBeaconBlockFromAltair(block *spec.VersionedSignedBeaconBlock) BeaconBlock {
	return BeaconBlock{
		AttesterSlashings: len(block.Altair.Message.Body.AttesterSlashings),
		ProposerSlashings: len(block.Altair.Message.Body.ProposerSlashings),
		Transactions:      0,
		Deposits:          len(block.Altair.Message.Body.Deposits),
		VoluntaryExits:    len(block.Altair.Message.Body.VoluntaryExits),
		Attestations:      len(block.Altair.Message.Body.Attestations),
		Slot:              uint64(block.Altair.Message.Slot),
	}
}

func NewBeaconBlockFromBellatrix(block *spec.VersionedSignedBeaconBlock) BeaconBlock {
	return BeaconBlock{
		AttesterSlashings: len(block.Bellatrix.Message.Body.AttesterSlashings),
		ProposerSlashings: len(block.Bellatrix.Message.Body.ProposerSlashings),
		Transactions:      len(block.Bellatrix.Message.Body.ExecutionPayload.Transactions),
		Deposits:          len(block.Bellatrix.Message.Body.Deposits),
		VoluntaryExits:    len(block.Bellatrix.Message.Body.VoluntaryExits),
		Attestations:      len(block.Bellatrix.Message.Body.Attestations),
		Slot:              uint64(block.Bellatrix.Message.Slot),
	}
}
