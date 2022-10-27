package execution

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/execution/jobs"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Metrics exposes Execution layer metrics
type Metrics interface {
	// StartAsync starts all the metrics jobs
	StartAsync(ctx context.Context)
}

type metrics struct {
	log            logrus.FieldLogger
	syncMetrics    jobs.SyncStatus
	generalMetrics jobs.GeneralMetrics
	txpoolMetrics  jobs.TXPool
	adminMetrics   jobs.Admin
	blockMetrics   jobs.BlockMetrics
	web3Metrics    jobs.Web3
	netMetrics     jobs.Net

	enabledJobs map[string]bool
}

// NewMetrics creates a new execution Metrics instance
func NewMetrics(client *ethclient.Client, internalAPI api.ExecutionClient, ethRPCClient *ethrpc.EthRPC, log logrus.FieldLogger, nodeName, namespace string, enabledModules []string) Metrics {
	constLabels := make(prometheus.Labels)
	constLabels["ethereum_role"] = "execution"
	constLabels["node_name"] = nodeName

	m := &metrics{
		log:            log,
		generalMetrics: jobs.NewGeneralMetrics(client, internalAPI, ethRPCClient, log, namespace, constLabels),
		syncMetrics:    jobs.NewSyncStatus(client, internalAPI, ethRPCClient, log, namespace, constLabels),
		txpoolMetrics:  jobs.NewTXPool(client, internalAPI, ethRPCClient, log, namespace, constLabels),
		adminMetrics:   jobs.NewAdmin(client, internalAPI, ethRPCClient, log, namespace, constLabels),
		blockMetrics:   jobs.NewBlockMetrics(client, internalAPI, ethRPCClient, log, namespace, constLabels),
		web3Metrics:    jobs.NewWeb3(client, internalAPI, ethRPCClient, log, namespace, constLabels),
		netMetrics:     jobs.NewNet(client, internalAPI, ethRPCClient, log, namespace, constLabels),

		enabledJobs: make(map[string]bool),
	}

	if able := jobs.ExporterCanRun(enabledModules, m.syncMetrics.RequiredModules()); able {
		m.log.Info("Enabling sync status metrics")
		m.enabledJobs[m.syncMetrics.Name()] = true

		prometheus.MustRegister(m.syncMetrics.Percentage)
		prometheus.MustRegister(m.syncMetrics.StartingBlock)
		prometheus.MustRegister(m.syncMetrics.CurrentBlock)
		prometheus.MustRegister(m.syncMetrics.IsSyncing)
		prometheus.MustRegister(m.syncMetrics.HighestBlock)
	}

	if able := jobs.ExporterCanRun(enabledModules, m.generalMetrics.RequiredModules()); able {
		m.log.Info("Enabling general metrics")
		m.enabledJobs[m.generalMetrics.Name()] = true

		prometheus.MustRegister(m.generalMetrics.NetworkID)
		prometheus.MustRegister(m.generalMetrics.GasPrice)
		prometheus.MustRegister(m.generalMetrics.ChainID)
	}

	if able := jobs.ExporterCanRun(enabledModules, m.blockMetrics.RequiredModules()); able {
		m.log.Info("Enabling block metrics")
		m.enabledJobs[m.blockMetrics.Name()] = true

		prometheus.MustRegister(m.blockMetrics.MostRecentBlockNumber)

		prometheus.MustRegister(m.blockMetrics.HeadBlockSize)
		prometheus.MustRegister(m.blockMetrics.HeadGasLimit)
		prometheus.MustRegister(m.blockMetrics.HeadGasUsed)
		prometheus.MustRegister(m.blockMetrics.HeadTransactionCount)
		prometheus.MustRegister(m.blockMetrics.HeadBaseFeePerGas)
		prometheus.MustRegister(m.blockMetrics.HeadTotalDifficulty)
		prometheus.MustRegister(m.blockMetrics.HeadTotalDifficultyTrillions)

		prometheus.MustRegister(m.blockMetrics.SafeBaseFeePerGas)
		prometheus.MustRegister(m.blockMetrics.SafeBlockSize)
		prometheus.MustRegister(m.blockMetrics.SafeGasLimit)
		prometheus.MustRegister(m.blockMetrics.SafeGasUsed)
		prometheus.MustRegister(m.blockMetrics.SafeTransactionCount)
	}

	if able := jobs.ExporterCanRun(enabledModules, m.txpoolMetrics.RequiredModules()); able {
		m.log.Info("Enabling txpool metrics")
		m.enabledJobs[m.txpoolMetrics.Name()] = true

		prometheus.MustRegister(m.txpoolMetrics.Transactions)
	}

	if able := jobs.ExporterCanRun(enabledModules, m.adminMetrics.RequiredModules()); able {
		m.log.Info("Enabling admin metrics")
		m.enabledJobs[m.adminMetrics.Name()] = true

		prometheus.MustRegister(m.adminMetrics.TotalDifficulty)
		prometheus.MustRegister(m.adminMetrics.TotalDifficultyTrillions)
		prometheus.MustRegister(m.adminMetrics.NodeInfo)
		prometheus.MustRegister(m.adminMetrics.Port)
		prometheus.MustRegister(m.adminMetrics.Peers)
	}

	if able := jobs.ExporterCanRun(enabledModules, m.web3Metrics.RequiredModules()); able {
		m.log.Info("Enabling web3 metrics")
		m.enabledJobs[m.web3Metrics.Name()] = true

		prometheus.MustRegister(m.web3Metrics.ClientVersion)
	}

	if able := jobs.ExporterCanRun(enabledModules, m.netMetrics.RequiredModules()); able {
		m.log.Info("Enabling net metrics")
		m.enabledJobs[m.netMetrics.Name()] = true

		prometheus.MustRegister(m.netMetrics.PeerCount)
	}

	return m
}

func (m *metrics) StartAsync(ctx context.Context) {
	if m.enabledJobs[m.syncMetrics.Name()] {
		go m.syncMetrics.Start(ctx)
	}

	if m.enabledJobs[m.generalMetrics.Name()] {
		go m.generalMetrics.Start(ctx)
	}

	if m.enabledJobs[m.txpoolMetrics.Name()] {
		go m.txpoolMetrics.Start(ctx)
	}

	if m.enabledJobs[m.adminMetrics.Name()] {
		go m.adminMetrics.Start(ctx)
	}

	if m.enabledJobs[m.blockMetrics.Name()] {
		go m.blockMetrics.Start(ctx)
	}

	if m.enabledJobs[m.web3Metrics.Name()] {
		go m.web3Metrics.Start(ctx)
	}

	if m.enabledJobs[m.netMetrics.Name()] {
		go m.netMetrics.Start(ctx)
	}

	m.log.Info("Started metrics exporter jobs")
}
