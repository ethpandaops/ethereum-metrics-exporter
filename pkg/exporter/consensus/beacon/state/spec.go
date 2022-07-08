package state

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/spf13/cast"
)

// Spec represents the state of the spec.
type Spec struct {
	PresetBase string
	ConfigName string

	DepositChainID uint64

	SafeSlotsToUpdateJustified phase0.Slot
	SlotsPerEpoch              phase0.Slot

	EpochsPerSyncCommitteePeriod phase0.Epoch
	MinSyncCommitteeParticipants uint64
	TargetCommitteeSize          uint64
	SyncCommitteeSize            uint64

	TerminalBlockHashActivationEpoch phase0.Epoch
	TerminalTotalDifficulty          big.Int

	MaxValidatorsPerCommittee uint64
	BaseRewardFactor          uint64
	EffectiveBalanceIncrement phase0.Gwei
	MaxEffectiveBalance       phase0.Gwei
	MinDepositAmount          phase0.Gwei
	MaxAttestations           uint64

	SecondsPerEth1Block            time.Duration
	GenesisDelay                   time.Duration
	SecondsPerSlot                 time.Duration
	MaxDeposits                    uint64
	MinGenesisActiveValidatorCount uint64
	Eth1FollowDistance             uint64

	ForkEpochs ForkEpochs
}

// NewSpec creates a new spec instance.
func NewSpec(data map[string]interface{}) Spec {
	spec := Spec{
		ForkEpochs: ForkEpochs{},
	}

	if safeSlotsToUpdateJustified, exists := data["SAFE_SLOTS_TO_UPDATE_JUSTIFIED"]; exists {
		spec.SafeSlotsToUpdateJustified = phase0.Slot(cast.ToUint64(safeSlotsToUpdateJustified))
	}

	if depositChainID, exists := data["DEPOSIT_CHAIN_ID"]; exists {
		spec.DepositChainID = cast.ToUint64(depositChainID)
	}

	if configName, exists := data["CONFIG_NAME"]; exists {
		spec.ConfigName = cast.ToString(configName)
	}

	if maxValidatorsPerCommittee, exists := data["MAX_VALIDATORS_PER_COMMITTEE"]; exists {
		spec.MaxValidatorsPerCommittee = cast.ToUint64(maxValidatorsPerCommittee)
	}

	if secondsPerEth1Block, exists := data["SECONDS_PER_ETH1_BLOCK"]; exists {
		spec.SecondsPerEth1Block = cast.ToDuration(secondsPerEth1Block)
	}

	if baseRewardFactor, exists := data["BASE_REWARD_FACTOR"]; exists {
		spec.BaseRewardFactor = cast.ToUint64(baseRewardFactor)
	}

	if epochsPerSyncComitteePeriod, exists := data["EPOCHS_PER_SYNC_COMMITTEE_PERIOD"]; exists {
		spec.EpochsPerSyncCommitteePeriod = phase0.Epoch(cast.ToUint64(epochsPerSyncComitteePeriod))
	}

	if effectiveBalanceIncrement, exists := data["EFFECTIVE_BALANCE_INCREMENT"]; exists {
		spec.EffectiveBalanceIncrement = phase0.Gwei(cast.ToUint64(effectiveBalanceIncrement))
	}

	if maxAttestations, exists := data["MAX_ATTESTATIONS"]; exists {
		spec.MaxAttestations = cast.ToUint64(maxAttestations)
	}

	if minSyncCommitteeParticipants, exists := data["MIN_SYNC_COMMITTEE_PARTICIPANTS"]; exists {
		spec.MinSyncCommitteeParticipants = cast.ToUint64(minSyncCommitteeParticipants)
	}

	if genesisDelay, exists := data["GENESIS_DELAY"]; exists {
		spec.GenesisDelay = cast.ToDuration(genesisDelay)
	}

	if secondsPerSlot, exists := data["SECONDS_PER_SLOT"]; exists {
		spec.SecondsPerSlot = cast.ToDuration(secondsPerSlot)
	}

	if maxEffectiveBalance, exists := data["MAX_EFFECTIVE_BALANCE"]; exists {
		spec.MaxEffectiveBalance = phase0.Gwei(cast.ToUint64(maxEffectiveBalance))
	}

	if terminalTotalDifficulty, exists := data["TERMINAL_TOTAL_DIFFICULTY"]; exists {
		ttd := cast.ToString(fmt.Sprintf("%v", terminalTotalDifficulty))

		casted, _ := (*big.NewInt(0)).SetString(ttd, 10)
		spec.TerminalTotalDifficulty = *casted
	}

	if maxDeposits, exists := data["MAX_DEPOSITS"]; exists {
		spec.MaxDeposits = cast.ToUint64(maxDeposits)
	}

	if minGenesisActiveValidatorCount, exists := data["MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"]; exists {
		spec.MinGenesisActiveValidatorCount = cast.ToUint64(minGenesisActiveValidatorCount)
	}

	if targetCommitteeSize, exists := data["TARGET_COMMITTEE_SIZE"]; exists {
		spec.TargetCommitteeSize = cast.ToUint64(targetCommitteeSize)
	}

	if syncCommitteeSize, exists := data["SYNC_COMMITTEE_SIZE"]; exists {
		spec.SyncCommitteeSize = cast.ToUint64(syncCommitteeSize)
	}

	if eth1FollowDistance, exists := data["ETH1_FOLLOW_DISTANCE"]; exists {
		spec.Eth1FollowDistance = cast.ToUint64(eth1FollowDistance)
	}

	if terminalBlockHashActivationEpoch, exists := data["TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH"]; exists {
		spec.TerminalBlockHashActivationEpoch = phase0.Epoch(cast.ToUint64(terminalBlockHashActivationEpoch))
	}

	if minDepositAmount, exists := data["MIN_DEPOSIT_AMOUNT"]; exists {
		spec.MinDepositAmount = phase0.Gwei(cast.ToUint64(minDepositAmount))
	}

	if slotsPerEpoch, exists := data["SLOTS_PER_EPOCH"]; exists {
		spec.SlotsPerEpoch = phase0.Slot(cast.ToUint64(slotsPerEpoch))
	}

	if presetBase, exists := data["PRESET_BASE"]; exists {
		spec.PresetBase = cast.ToString(presetBase)
	}

	for k, v := range data {
		if strings.Contains(k, "_FORK_EPOCH") {
			forkName := strings.ReplaceAll(k, "_FORK_EPOCH", "")

			spec.ForkEpochs = append(spec.ForkEpochs, ForkEpoch{
				Epoch: phase0.Epoch(cast.ToUint64(v)),
				Name:  forkName,
			})
		}
	}

	return spec
}

// Validate performs basic validation of the spec.
func (s *Spec) Validate() error {
	return nil
}
