package consensus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

type SpecMetrics struct {
	SafeSlotsToUpdateJustified       prometheus.Gauge
	DepositChainID                   prometheus.Gauge
	ConfigName                       prometheus.GaugeVec
	MaxValidatorsPerCommittee        prometheus.Gauge
	SecondsPerEth1Block              prometheus.Gauge
	BaseRewardFactor                 prometheus.Gauge
	EpochsPerSyncCommitteePeriod     prometheus.Gauge
	EffectiveBalanceIncrement        prometheus.Gauge
	MaxAttestations                  prometheus.Gauge
	MinSyncCommitteeParticipants     prometheus.Gauge
	GenesisDelay                     prometheus.Gauge
	SecondsPerSlot                   prometheus.Gauge
	MaxEffectiveBalance              prometheus.Gauge
	TerminalTotalDifficulty          prometheus.Gauge
	MaxDeposits                      prometheus.Gauge
	MinGenesisActiveValidatorCount   prometheus.Gauge
	TargetCommitteeSize              prometheus.Gauge
	SyncCommitteeSize                prometheus.Gauge
	Eth1FollowDistance               prometheus.Gauge
	TerminalBlockHashActivationEpoch prometheus.Gauge
	MinDepositAmount                 prometheus.Gauge
	SlotsPerEpoch                    prometheus.Gauge
	PresetBase                       prometheus.GaugeVec
}

func NewSpecMetrics(namespace string, constLabels map[string]string) SpecMetrics {
	namespace = namespace + "_spec"
	return SpecMetrics{
		SafeSlotsToUpdateJustified: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "safe_slots_to_update_justified",
				Help:        "The number of slots to wait before updating the justified checkpoint.",
				ConstLabels: constLabels,
			},
		),
		DepositChainID: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "deposit_chain_id",
				Help:        "The chain ID of the deposit contract.",
				ConstLabels: constLabels,
			},
		),
		ConfigName: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "config_name",
				Help:        "The name of the config.",
				ConstLabels: constLabels,
			},
			[]string{"name"},
		),
		MaxValidatorsPerCommittee: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "max_validators_per_committee",
				Help:        "The maximum number of validators per committee.",
				ConstLabels: constLabels,
			},
		),
		SecondsPerEth1Block: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "seconds_per_eth1_block",
				Help:        "The number of seconds per ETH1 block.",
				ConstLabels: constLabels,
			},
		),
		BaseRewardFactor: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "base_reward_factor",
				Help:        "The base reward factor.",
				ConstLabels: constLabels,
			},
		),
		EpochsPerSyncCommitteePeriod: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "epochs_per_sync_committee_period",
				Help:        "The number of epochs per sync committee period.",
				ConstLabels: constLabels,
			},
		),
		EffectiveBalanceIncrement: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "effective_balance_increment",
				Help:        "The effective balance increment.",
				ConstLabels: constLabels,
			},
		),
		MaxAttestations: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "max_attestations",
				Help:        "The maximum number of attestations.",
				ConstLabels: constLabels,
			},
		),
		MinSyncCommitteeParticipants: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "min_sync_committee_participants",
				Help:        "The minimum number of sync committee participants.",
				ConstLabels: constLabels,
			},
		),
		GenesisDelay: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "genesis_delay",
				Help:        "The number of epochs to wait before processing the genesis block.",
				ConstLabels: constLabels,
			},
		),
		SecondsPerSlot: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "seconds_per_slot",
				Help:        "The number of seconds per slot.",
				ConstLabels: constLabels,
			},
		),
		MaxEffectiveBalance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "max_effective_balance",
				Help:        "The maximum effective balance.",
				ConstLabels: constLabels,
			},
		),
		TerminalTotalDifficulty: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "terminal_total_difficulty",
				Help:        "The terminal total difficulty.",
				ConstLabels: constLabels,
			},
		),
		MaxDeposits: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "max_deposits",
				Help:        "The maximum number of deposits.",
				ConstLabels: constLabels,
			},
		),
		MinGenesisActiveValidatorCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "min_genesis_active_validator_count",
				Help:        "The minimum number of active validators at genesis.",
				ConstLabels: constLabels,
			},
		),
		TargetCommitteeSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "target_committee_size",
				Help:        "The target committee size.",
				ConstLabels: constLabels,
			},
		),
		SyncCommitteeSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_committee_size",
				Help:        "The sync committee size.",
				ConstLabels: constLabels,
			},
		),
		Eth1FollowDistance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "eth1_follow_distance",
				Help:        "The number of blocks to follow behind the head of the eth1 chain.",
				ConstLabels: constLabels,
			},
		),
		TerminalBlockHashActivationEpoch: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "terminal_block_hash_activation_epoch",
				Help:        "The epoch at which the terminal block hash is activated.",
				ConstLabels: constLabels,
			},
		),
		MinDepositAmount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "min_deposit_amount",
				Help:        "The minimum deposit amount.",
				ConstLabels: constLabels,
			},
		),
		SlotsPerEpoch: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "slots_per_epoch",
				Help:        "The number of slots per epoch.",
				ConstLabels: constLabels,
			},
		),
		PresetBase: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "preset_base",
				Help:        "The base of the preset.",
				ConstLabels: constLabels,
			},
			[]string{"preset"},
		),
	}
}

func (c *SpecMetrics) Update(spec map[string]interface{}) {
	if safeSlotsToUpdateJustified, exists := spec["SAFE_SLOTS_TO_UPDATE_JUSTIFIED"]; exists {
		c.SafeSlotsToUpdateJustified.Set(cast.ToFloat64(safeSlotsToUpdateJustified))
	}

	if depositChainId, exists := spec["DEPOSIT_CHAIN_ID"]; exists {
		c.DepositChainID.Set(cast.ToFloat64(depositChainId))
	}

	if configName, exists := spec["CONFIG_NAME"]; exists {
		c.ConfigName.WithLabelValues(cast.ToString(configName)).Set(1)
	}

	if maxValidatorsPerCommittee, exists := spec["MAX_VALIDATORS_PER_COMMITTEE"]; exists {
		c.MaxValidatorsPerCommittee.Set(cast.ToFloat64(maxValidatorsPerCommittee))
	}

	if secondsPerEth1Block, exists := spec["SECONDS_PER_ETH1_BLOCK"]; exists {
		c.SecondsPerEth1Block.Set(float64(cast.ToDuration(secondsPerEth1Block)))
	}

	if baseRewardFactor, exists := spec["BASE_REWARD_FACTOR"]; exists {
		c.BaseRewardFactor.Set(cast.ToFloat64(baseRewardFactor))
	}

	if epochsPerSyncComitteePeriod, exists := spec["EPOCHS_PER_SYNC_COMMITTEE_PERIOD"]; exists {
		c.EpochsPerSyncCommitteePeriod.Set(cast.ToFloat64(epochsPerSyncComitteePeriod))
	}

	if effectiveBalanceIncrement, exists := spec["EFFECTIVE_BALANCE_INCREMENT"]; exists {
		c.EffectiveBalanceIncrement.Set(cast.ToFloat64(effectiveBalanceIncrement))
	}

	if maxAttestations, exists := spec["MAX_ATTESTATIONS"]; exists {
		c.MaxAttestations.Set(cast.ToFloat64(maxAttestations))
	}

	if minSyncCommitteeParticipants, exists := spec["MIN_SYNC_COMMITTEE_PARTICIPANTS"]; exists {
		c.MinSyncCommitteeParticipants.Set(cast.ToFloat64(minSyncCommitteeParticipants))
	}

	if genesisDelay, exists := spec["GENESIS_DELAY"]; exists {
		c.GenesisDelay.Set(float64(cast.ToDuration(genesisDelay)))
	}

	if secondsPerSlot, exists := spec["SECONDS_PER_SLOT"]; exists {
		c.SecondsPerSlot.Set(float64(cast.ToDuration(secondsPerSlot)))
	}

	if maxEffectiveBalance, exists := spec["MAX_EFFECTIVE_BALANCE"]; exists {
		c.MaxEffectiveBalance.Set(cast.ToFloat64(maxEffectiveBalance))
	}

	if terminalTotalDifficulty, exists := spec["TERMINAL_TOTAL_DIFFICULTY"]; exists {
		c.TerminalTotalDifficulty.Set(cast.ToFloat64(terminalTotalDifficulty))
	}

	if maxDeposits, exists := spec["MAX_DEPOSITS"]; exists {
		c.MaxDeposits.Set(cast.ToFloat64((maxDeposits)))
	}

	if minGenesisActiveValidatorCount, exists := spec["MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"]; exists {
		c.MinGenesisActiveValidatorCount.Set(cast.ToFloat64(minGenesisActiveValidatorCount))
	}

	if targetCommitteeSize, exists := spec["TARGET_COMMITTEE_SIZE"]; exists {
		c.TargetCommitteeSize.Set(cast.ToFloat64(targetCommitteeSize))
	}

	if syncCommitteeSize, exists := spec["SYNC_COMMITTEE_SIZE"]; exists {
		c.SyncCommitteeSize.Set(cast.ToFloat64(syncCommitteeSize))
	}

	if eth1FollowDistance, exists := spec["ETH1_FOLLOW_DISTANCE"]; exists {
		c.Eth1FollowDistance.Set(cast.ToFloat64(eth1FollowDistance))
	}

	if terminalBlockHashActivationEpoch, exists := spec["TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH"]; exists {
		c.TerminalBlockHashActivationEpoch.Set(cast.ToFloat64(terminalBlockHashActivationEpoch))
	}

	if minDepositAmount, exists := spec["MIN_DEPOSIT_AMOUNT"]; exists {
		c.MinDepositAmount.Set(cast.ToFloat64(minDepositAmount))
	}

	if slotsPerEpoch, exists := spec["SLOTS_PER_EPOCH"]; exists {
		c.SlotsPerEpoch.Set(cast.ToFloat64(slotsPerEpoch))
	}

	if presetBase, exists := spec["PRESET_BASE"]; exists {
		c.PresetBase.WithLabelValues(cast.ToString(presetBase)).Set(1)
	}

}
