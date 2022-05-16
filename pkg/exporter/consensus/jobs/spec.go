package jobs

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// Spec reports metrics about the configured consensus spec.
type Spec struct {
	MetricExporter
	client                           eth2client.Service
	log                              logrus.FieldLogger
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
	TerminalTotalDifficultyTrillions prometheus.Gauge
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

const (
	NameSpec = "spec"
)

// NewSpecJob returns a new Spec instance.
func NewSpecJob(client eth2client.Service, log logrus.FieldLogger, namespace string, constLabels map[string]string) Spec {
	constLabels["module"] = NameSpec
	namespace = namespace + "_spec"
	return Spec{
		client: client,
		log:    log,
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
		TerminalTotalDifficultyTrillions: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "terminal_total_difficulty_trillions",
				Help:        "The terminal total difficulty in trillions.",
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

func (s *Spec) Name() string {
	return NameSpec
}

func (s *Spec) Start(ctx context.Context) {
	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 600):
			s.tick(ctx)
		}
	}
}

func (s *Spec) tick(ctx context.Context) {
	if err := s.GetSpec(ctx); err != nil {
		s.log.WithError(err).Error("Failed to fetch spec")
	}
}

func (s *Spec) GetSpec(ctx context.Context) error {
	provider, isProvider := s.client.(eth2client.SpecProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.SpecProvider")
	}

	spec, err := provider.Spec(ctx)
	if err != nil {
		return err
	}

	s.Update(spec)

	return nil
}

func (s *Spec) Update(spec map[string]interface{}) {
	if safeSlotsToUpdateJustified, exists := spec["SAFE_SLOTS_TO_UPDATE_JUSTIFIED"]; exists {
		s.SafeSlotsToUpdateJustified.Set(cast.ToFloat64(safeSlotsToUpdateJustified))
	}

	if depositChainId, exists := spec["DEPOSIT_CHAIN_ID"]; exists {
		s.DepositChainID.Set(cast.ToFloat64(depositChainId))
	}

	if configName, exists := spec["CONFIG_NAME"]; exists {
		s.ConfigName.WithLabelValues(cast.ToString(configName)).Set(1)
	}

	if maxValidatorsPerCommittee, exists := spec["MAX_VALIDATORS_PER_COMMITTEE"]; exists {
		s.MaxValidatorsPerCommittee.Set(cast.ToFloat64(maxValidatorsPerCommittee))
	}

	if secondsPerEth1Block, exists := spec["SECONDS_PER_ETH1_BLOCK"]; exists {
		s.SecondsPerEth1Block.Set(float64(cast.ToDuration(secondsPerEth1Block)))
	}

	if baseRewardFactor, exists := spec["BASE_REWARD_FACTOR"]; exists {
		s.BaseRewardFactor.Set(cast.ToFloat64(baseRewardFactor))
	}

	if epochsPerSyncComitteePeriod, exists := spec["EPOCHS_PER_SYNC_COMMITTEE_PERIOD"]; exists {
		s.EpochsPerSyncCommitteePeriod.Set(cast.ToFloat64(epochsPerSyncComitteePeriod))
	}

	if effectiveBalanceIncrement, exists := spec["EFFECTIVE_BALANCE_INCREMENT"]; exists {
		s.EffectiveBalanceIncrement.Set(cast.ToFloat64(effectiveBalanceIncrement))
	}

	if maxAttestations, exists := spec["MAX_ATTESTATIONS"]; exists {
		s.MaxAttestations.Set(cast.ToFloat64(maxAttestations))
	}

	if minSyncCommitteeParticipants, exists := spec["MIN_SYNC_COMMITTEE_PARTICIPANTS"]; exists {
		s.MinSyncCommitteeParticipants.Set(cast.ToFloat64(minSyncCommitteeParticipants))
	}

	if genesisDelay, exists := spec["GENESIS_DELAY"]; exists {
		s.GenesisDelay.Set(float64(cast.ToDuration(genesisDelay)))
	}

	if secondsPerSlot, exists := spec["SECONDS_PER_SLOT"]; exists {
		s.SecondsPerSlot.Set(float64(cast.ToDuration(secondsPerSlot)))
	}

	if maxEffectiveBalance, exists := spec["MAX_EFFECTIVE_BALANCE"]; exists {
		s.MaxEffectiveBalance.Set(cast.ToFloat64(maxEffectiveBalance))
	}

	if terminalTotalDifficulty, exists := spec["TERMINAL_TOTAL_DIFFICULTY"]; exists {
		ttd := cast.ToString(fmt.Sprintf("%v", terminalTotalDifficulty))
		asBigInt, success := big.NewInt(0).SetString(ttd, 10)
		if success {
			trillion := big.NewInt(1e12)
			divided := new(big.Int).Div(asBigInt, trillion)
			asFloat, _ := new(big.Float).SetInt(divided).Float64()
			s.TerminalTotalDifficultyTrillions.Set(asFloat)
			s.TerminalTotalDifficulty.Set(float64(asBigInt.Uint64()))
		}
	}

	if maxDeposits, exists := spec["MAX_DEPOSITS"]; exists {
		s.MaxDeposits.Set(cast.ToFloat64((maxDeposits)))
	}

	if minGenesisActiveValidatorCount, exists := spec["MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"]; exists {
		s.MinGenesisActiveValidatorCount.Set(cast.ToFloat64(minGenesisActiveValidatorCount))
	}

	if targetCommitteeSize, exists := spec["TARGET_COMMITTEE_SIZE"]; exists {
		s.TargetCommitteeSize.Set(cast.ToFloat64(targetCommitteeSize))
	}

	if syncCommitteeSize, exists := spec["SYNC_COMMITTEE_SIZE"]; exists {
		s.SyncCommitteeSize.Set(cast.ToFloat64(syncCommitteeSize))
	}

	if eth1FollowDistance, exists := spec["ETH1_FOLLOW_DISTANCE"]; exists {
		s.Eth1FollowDistance.Set(cast.ToFloat64(eth1FollowDistance))
	}

	if terminalBlockHashActivationEpoch, exists := spec["TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH"]; exists {
		s.TerminalBlockHashActivationEpoch.Set(cast.ToFloat64(terminalBlockHashActivationEpoch))
	}

	if minDepositAmount, exists := spec["MIN_DEPOSIT_AMOUNT"]; exists {
		s.MinDepositAmount.Set(cast.ToFloat64(minDepositAmount))
	}

	if slotsPerEpoch, exists := spec["SLOTS_PER_EPOCH"]; exists {
		s.SlotsPerEpoch.Set(cast.ToFloat64(slotsPerEpoch))
	}

	if presetBase, exists := spec["PRESET_BASE"]; exists {
		s.PresetBase.WithLabelValues(cast.ToString(presetBase)).Set(1)
	}

}
