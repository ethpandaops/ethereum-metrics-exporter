package jobs

import (
	"context"
	"math/big"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon/state"
	"github.com/sirupsen/logrus"
)

// Spec reports metrics about the configured consensus spec.
type Spec struct {
	beacon                           beacon.Node
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
func NewSpecJob(bc beacon.Node, log logrus.FieldLogger, namespace string, constLabels map[string]string) Spec {
	constLabels["module"] = NameSpec

	namespace += "_spec"

	return Spec{
		log:    log,
		beacon: bc,
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

func (s *Spec) Start(ctx context.Context) error {
	if _, err := s.beacon.OnSpecUpdated(ctx, func(ctx context.Context, event *beacon.SpecUpdatedEvent) error {
		return s.observeSpec(ctx, event.Spec)
	}); err != nil {
		return err
	}

	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Minute * 5):
			s.tick(ctx)
		}
	}
}

func (s *Spec) tick(ctx context.Context) {
	if err := s.getSpec(ctx); err != nil {
		s.log.WithError(err).Error("Failed to fetch spec")
	}
}

func (s *Spec) observeSpec(ctx context.Context, spec *state.Spec) error {
	s.ConfigName.Reset()
	s.ConfigName.WithLabelValues(spec.ConfigName).Set(1)

	s.PresetBase.Reset()
	s.PresetBase.WithLabelValues(spec.PresetBase).Set(1)

	s.SafeSlotsToUpdateJustified.Set(float64(spec.SafeSlotsToUpdateJustified))
	s.DepositChainID.Set(float64(spec.DepositChainID))
	s.MaxValidatorsPerCommittee.Set(float64(spec.MaxValidatorsPerCommittee))
	// nolint:unconvert // false positive
	s.SecondsPerEth1Block.Set(float64(spec.SecondsPerEth1Block.Seconds()))
	s.BaseRewardFactor.Set(float64(spec.BaseRewardFactor))
	s.EpochsPerSyncCommitteePeriod.Set(float64(spec.EpochsPerSyncCommitteePeriod))
	s.EffectiveBalanceIncrement.Set(float64(spec.EffectiveBalanceIncrement))
	s.MaxAttestations.Set(float64(spec.MaxAttestations))
	s.MinSyncCommitteeParticipants.Set(float64(spec.MinSyncCommitteeParticipants))
	// nolint:unconvert // false positive
	s.GenesisDelay.Set(float64(spec.GenesisDelay.Seconds()))
	// nolint:unconvert // false positive
	s.SecondsPerSlot.Set(float64(spec.SecondsPerSlot.Seconds()))
	s.MaxEffectiveBalance.Set(float64(spec.MaxEffectiveBalance))
	s.MaxDeposits.Set(float64(spec.MaxDeposits))
	s.MinGenesisActiveValidatorCount.Set(float64(spec.MinGenesisActiveValidatorCount))
	s.TargetCommitteeSize.Set(float64(spec.TargetCommitteeSize))
	s.SyncCommitteeSize.Set(float64(spec.SyncCommitteeSize))
	s.Eth1FollowDistance.Set(float64(spec.Eth1FollowDistance))
	s.TerminalBlockHashActivationEpoch.Set(float64(spec.TerminalBlockHashActivationEpoch))
	s.MinDepositAmount.Set(float64(spec.MinDepositAmount))
	s.SlotsPerEpoch.Set(float64(spec.SlotsPerEpoch))

	trillion := big.NewInt(1e12)
	divided := new(big.Int).Div(&spec.TerminalTotalDifficulty, trillion)
	asFloat, _ := new(big.Float).SetInt(divided).Float64()
	s.TerminalTotalDifficultyTrillions.Set(asFloat)
	s.TerminalTotalDifficulty.Set(float64(spec.TerminalTotalDifficulty.Uint64()))

	return nil
}

func (s *Spec) getSpec(ctx context.Context) error {
	spec, err := s.beacon.GetSpec(ctx)
	if err != nil {
		return err
	}

	return s.observeSpec(ctx, spec)
}
