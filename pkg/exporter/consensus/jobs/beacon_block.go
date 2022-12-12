package jobs

import (
	"github.com/attestantio/go-eth2-client/spec"
)

func GetDepositCountsFromBeaconBlock(block *spec.VersionedSignedBeaconBlock) int {
	switch block.Version {
	case spec.DataVersionPhase0:
		return len(block.Phase0.Message.Body.Deposits)
	case spec.DataVersionAltair:
		return len(block.Altair.Message.Body.Deposits)
	case spec.DataVersionBellatrix:
		return len(block.Bellatrix.Message.Body.Deposits)
	case spec.DataVersionCapella:
		return len(block.Capella.Message.Body.Deposits)
	default:
		return 0
	}
}

func GetVoluntaryExitsFromBeaconBlock(block *spec.VersionedSignedBeaconBlock) int {
	switch block.Version {
	case spec.DataVersionPhase0:
		return len(block.Phase0.Message.Body.VoluntaryExits)
	case spec.DataVersionAltair:
		return len(block.Altair.Message.Body.VoluntaryExits)
	case spec.DataVersionBellatrix:
		return len(block.Bellatrix.Message.Body.VoluntaryExits)
	case spec.DataVersionCapella:
		return len(block.Capella.Message.Body.VoluntaryExits)
	default:
		return 0
	}
}

func GetTransactionsCountFromBeaconBlock(block *spec.VersionedSignedBeaconBlock) int {
	switch block.Version {
	case spec.DataVersionPhase0:
		return 0
	case spec.DataVersionAltair:
		return 0
	case spec.DataVersionBellatrix:
		return len(block.Bellatrix.Message.Body.ExecutionPayload.Transactions)
	case spec.DataVersionCapella:
		return len(block.Capella.Message.Body.ExecutionPayload.Transactions)
	default:
		return 0
	}
}
