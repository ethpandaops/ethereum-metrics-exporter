package consensus

import (
	"context"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/jobs"
	"github.com/sirupsen/logrus"
)

type Metrics interface {
	StartAsync(ctx context.Context)
}

type metrics struct {
	log logrus.FieldLogger

	generalMetrics jobs.General
	syncMetrics    jobs.Sync
	specMetrics    jobs.Spec
	forkMetrics    jobs.Forks
}

func NewMetrics(client eth2client.Service, log logrus.FieldLogger, nodeName, namespace string) Metrics {
	constLabels := make(prometheus.Labels)
	constLabels["ethereum_role"] = "consensus"
	constLabels["node_name"] = nodeName

	m := &metrics{
		log:            log,
		generalMetrics: jobs.NewGeneralJob(client, log, namespace, constLabels),
		specMetrics:    jobs.NewSpecJob(client, log, namespace, constLabels),
		syncMetrics:    jobs.NewSyncJob(client, log, namespace, constLabels),
		forkMetrics:    jobs.NewForksJob(client, log, namespace, constLabels),
	}

	prometheus.MustRegister(m.generalMetrics.Slots)
	prometheus.MustRegister(m.generalMetrics.NodeVersion)
	prometheus.MustRegister(m.generalMetrics.NetworkdID)
	prometheus.MustRegister(m.generalMetrics.ReOrgs)
	prometheus.MustRegister(m.generalMetrics.ReOrgDepth)

	prometheus.MustRegister(m.syncMetrics.Percentage)
	prometheus.MustRegister(m.syncMetrics.EstimatedHighestSlot)
	prometheus.MustRegister(m.syncMetrics.HeadSlot)
	prometheus.MustRegister(m.syncMetrics.Distance)
	prometheus.MustRegister(m.syncMetrics.IsSyncing)

	prometheus.MustRegister(m.specMetrics.SafeSlotsToUpdateJustified)
	prometheus.MustRegister(m.specMetrics.DepositChainID)
	prometheus.MustRegister(m.specMetrics.ConfigName)
	prometheus.MustRegister(m.specMetrics.MaxValidatorsPerCommittee)
	prometheus.MustRegister(m.specMetrics.SecondsPerEth1Block)
	prometheus.MustRegister(m.specMetrics.BaseRewardFactor)
	prometheus.MustRegister(m.specMetrics.EpochsPerSyncCommitteePeriod)
	prometheus.MustRegister(m.specMetrics.EffectiveBalanceIncrement)
	prometheus.MustRegister(m.specMetrics.MaxAttestations)
	prometheus.MustRegister(m.specMetrics.MinSyncCommitteeParticipants)
	prometheus.MustRegister(m.specMetrics.GenesisDelay)
	prometheus.MustRegister(m.specMetrics.SecondsPerSlot)
	prometheus.MustRegister(m.specMetrics.MaxEffectiveBalance)
	prometheus.MustRegister(m.specMetrics.TerminalTotalDifficulty)
	prometheus.MustRegister(m.specMetrics.TerminalTotalDifficultyTrillions)
	prometheus.MustRegister(m.specMetrics.MaxDeposits)
	prometheus.MustRegister(m.specMetrics.MinGenesisActiveValidatorCount)
	prometheus.MustRegister(m.specMetrics.TargetCommitteeSize)
	prometheus.MustRegister(m.specMetrics.SyncCommitteeSize)
	prometheus.MustRegister(m.specMetrics.Eth1FollowDistance)
	prometheus.MustRegister(m.specMetrics.TerminalBlockHashActivationEpoch)
	prometheus.MustRegister(m.specMetrics.MinDepositAmount)
	prometheus.MustRegister(m.specMetrics.SlotsPerEpoch)
	prometheus.MustRegister(m.specMetrics.PresetBase)

	prometheus.MustRegister(m.forkMetrics.Epochs)
	prometheus.MustRegister(m.forkMetrics.Current)
	prometheus.MustRegister(m.forkMetrics.Activated)
	return m
}

func (m *metrics) StartAsync(ctx context.Context) {
	go m.generalMetrics.Start(ctx)
	go m.specMetrics.Start(ctx)
	go m.syncMetrics.Start(ctx)
	go m.forkMetrics.Start(ctx)
}
