package consensus

import "github.com/prometheus/client_golang/prometheus"

type Metrics interface {
	ObserveSyncPercentage(percent float64)
	ObserveSyncEstimatedHighestSlot(slot uint64)
	ObserveSyncHeadSlot(slot uint64)
	ObserveSyncDistance(slots uint64)
	ObserveSyncIsSyncing(syncing bool)
	ObserveNodeVersion(version string)
	ObserveSpec(spec map[string]interface{})
}

type metrics struct {
	syncPercentage           prometheus.Gauge
	syncEstimatedHighestSlot prometheus.Gauge
	syncHeadSlot             prometheus.Gauge
	syncDistance             prometheus.Gauge
	syncIsSyncing            prometheus.Gauge
	nodeVersion              *prometheus.GaugeVec

	specMetrics SpecMetrics
}

func NewMetrics(nodeName, namespace string) Metrics {
	constLabels := make(prometheus.Labels)
	constLabels["ethereum_role"] = "consensus"
	constLabels["node_name"] = nodeName

	m := &metrics{
		syncPercentage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_percentage",
				Help:        "How synced the node is with the network (0-100%).",
				ConstLabels: constLabels,
			},
		),
		syncEstimatedHighestSlot: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_estimated_highest_slot",
				Help:        "The estimated highest slot of the network.",
				ConstLabels: constLabels,
			},
		),
		syncHeadSlot: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_head_slot",
				Help:        "The current slot of the node.",
				ConstLabels: constLabels,
			},
		),
		syncDistance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_distance",
				Help:        "The sync distance of the node.",
				ConstLabels: constLabels,
			},
		),
		syncIsSyncing: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "sync_is_syncing",
				Help:        "1 if the node is in syncing state.",
				ConstLabels: constLabels,
			},
		),
		nodeVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "node_version",
				Help:        "The version of the running beacon node.",
				ConstLabels: constLabels,
			},
			[]string{
				"version",
			},
		),
		specMetrics: NewSpecMetrics(namespace, constLabels),
	}

	prometheus.MustRegister(m.syncPercentage)
	prometheus.MustRegister(m.syncEstimatedHighestSlot)
	prometheus.MustRegister(m.syncHeadSlot)
	prometheus.MustRegister(m.syncDistance)
	prometheus.MustRegister(m.syncIsSyncing)
	prometheus.MustRegister(m.nodeVersion)

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
	prometheus.MustRegister(m.specMetrics.MaxDeposits)
	prometheus.MustRegister(m.specMetrics.MinGenesisActiveValidatorCount)
	prometheus.MustRegister(m.specMetrics.TargetCommitteeSize)
	prometheus.MustRegister(m.specMetrics.SyncCommitteeSize)
	prometheus.MustRegister(m.specMetrics.Eth1FollowDistance)
	prometheus.MustRegister(m.specMetrics.TerminalBlockHashActivationEpoch)
	prometheus.MustRegister(m.specMetrics.MinDepositAmount)
	prometheus.MustRegister(m.specMetrics.SlotsPerEpoch)
	prometheus.MustRegister(m.specMetrics.PresetBase)

	return m
}

func (m *metrics) ObserveSyncPercentage(percent float64) {
	m.syncPercentage.Set(percent)
}

func (m *metrics) ObserveSyncEstimatedHighestSlot(slot uint64) {
	m.syncEstimatedHighestSlot.Set(float64(slot))
}

func (m *metrics) ObserveSyncHeadSlot(slot uint64) {
	m.syncHeadSlot.Set(float64(slot))
}

func (m *metrics) ObserveSyncDistance(slots uint64) {
	m.syncDistance.Set(float64(slots))
}

func (m *metrics) ObserveSyncIsSyncing(syncing bool) {
	if syncing {
		m.syncIsSyncing.Set(1)
		return
	}

	m.syncIsSyncing.Set(0)
}

func (m *metrics) ObserveNodeVersion(version string) {
	m.nodeVersion.WithLabelValues(version).Set(float64(1))
}

func (m *metrics) ObserveSpec(spec map[string]interface{}) {
	m.specMetrics.Update(spec)
}
