package jobs

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// BlockMetrics exposes metrics on the head/safest block.
type BlockMetrics struct {
	client       *ethclient.Client
	api          api.ExecutionClient
	ethRPCClient *ethrpc.EthRPC
	log          logrus.FieldLogger

	MostRecentBlockNumber prometheus.GaugeVec

	HeadGasUsed                  prometheus.Gauge
	HeadGasLimit                 prometheus.Gauge
	HeadBaseFeePerGas            prometheus.Gauge
	HeadBlockSize                prometheus.Gauge
	HeadTransactionCount         prometheus.Gauge
	HeadTotalDifficulty          prometheus.Gauge
	HeadTotalDifficultyTrillions prometheus.Gauge

	SafeGasUsed          prometheus.Counter
	SafeGasLimit         prometheus.Counter
	SafeBaseFeePerGas    prometheus.Counter
	SafeBlockSize        prometheus.Counter
	SafeTransactionCount prometheus.Counter

	safeDistanceBlocks     uint64
	currentHeadBlockNumber uint64
	currentSafeBlockNumber uint64
}

const (
	NameBlock = "block"

	SafeDistanceBlocks = 6
)

func (b *BlockMetrics) Name() string {
	return NameGeneral
}

func (b *BlockMetrics) RequiredModules() []string {
	return []string{"eth", "net"}
}

// NewBlockMetrics returns a new Block metrics instance.
func NewBlockMetrics(client *ethclient.Client, internalAPI api.ExecutionClient, ethRPCClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) BlockMetrics {
	constLabels["module"] = NameBlock

	namespace = namespace + "_" + NameBlock

	return BlockMetrics{
		client:       client,
		api:          internalAPI,
		ethRPCClient: ethRPCClient,
		log:          log.WithField("module", NameBlock),

		MostRecentBlockNumber: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "most_recent_number",
				Help:        "The most recent block number.",
				ConstLabels: constLabels,
			},
			[]string{
				"identifier",
			},
		),

		HeadGasUsed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_gas_used",
				Help:        "The gas used in the most recent block.",
				ConstLabels: constLabels,
			},
		),
		HeadGasLimit: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_gas_limit",
				Help:        "The gas limit of the most recent block.",
				ConstLabels: constLabels,
			},
		),
		HeadBaseFeePerGas: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_base_fee_per_gas",
				Help:        "The base fee per gas in the most recent block.",
				ConstLabels: constLabels,
			},
		),
		HeadBlockSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_block_size_bytes",
				Help:        "The size of the most recent block (in bytes).",
				ConstLabels: constLabels,
			},
		),
		HeadTransactionCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_transactions_in_block",
				Help:        "The number of transactions in the most recent block.",
				ConstLabels: constLabels,
			},
		),
		HeadTotalDifficulty: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_total_difficulty",
				Help:        "The total difficulty of the head block.",
				ConstLabels: constLabels,
			},
		),
		HeadTotalDifficultyTrillions: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_total_difficulty_trillions",
				Help:        "The total difficulty of the head block (in trillions).",
				ConstLabels: constLabels,
			},
		),

		SafeGasUsed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "safe_gas_used",
				Help:        "The gas usedin the most recent safe block .",
				ConstLabels: constLabels,
			},
		),
		SafeGasLimit: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "safe_gas_limit",
				Help:        "The gas limit in the most recent safe block .",
				ConstLabels: constLabels,
			},
		),
		SafeBaseFeePerGas: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "safe_base_fee_per_gas",
				Help:        "The base fee per gas in the most recent safe block .",
				ConstLabels: constLabels,
			},
		),
		SafeBlockSize: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "safe_block_size_bytes",
				Help:        "The size of the most recent safe block (in bytes).",
				ConstLabels: constLabels,
			},
		),
		SafeTransactionCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "safe_transaction_count",
				Help:        "The number of transactions in the most recent safe block.",
				ConstLabels: constLabels,
			},
		),

		safeDistanceBlocks: SafeDistanceBlocks,

		currentHeadBlockNumber: 0,
		currentSafeBlockNumber: 0,
	}
}

func (b *BlockMetrics) Start(ctx context.Context) {
	b.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
			b.tick(ctx)
		}
	}
}

func (b *BlockMetrics) tick(ctx context.Context) {
	if err := b.getHeadBlockStats(ctx); err != nil {
		b.log.WithError(err).Error("Failed to get head block stats")
	}
}

func (b *BlockMetrics) getHeadBlockStats(ctx context.Context) error {
	mostRecentBlockNumber, err := b.client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	// No-op if we've already reported this block number.
	if mostRecentBlockNumber == b.currentHeadBlockNumber {
		return nil
	}

	b.currentHeadBlockNumber = mostRecentBlockNumber
	b.MostRecentBlockNumber.WithLabelValues("head").Set(float64(mostRecentBlockNumber))

	block, err := b.ethRPCClient.EthGetBlockByNumber(int(mostRecentBlockNumber), false)
	if err != nil {
		return err
	}

	if block == nil {
		return errors.New("block is nil")
	}

	b.HeadGasUsed.Set(float64(block.GasUsed))
	b.HeadGasLimit.Set(float64(block.GasLimit))
	b.HeadBlockSize.Set(float64(block.Size))
	b.HeadTransactionCount.Set(float64(len(block.Transactions)))
	// b.HeadBaseFeePerGas.Set(float64(block.BaseFee().Int64())) TODO(sam.calder-mason): Fix me

	b.HeadTotalDifficulty.Set(float64(block.TotalDifficulty.Uint64()))
	// Since we can't represent a big.Int as a float64, and the TD on mainnet is beyond float64, we'll divide the number by a trillion
	trillion := big.NewInt(1e12)
	divided := new(big.Int).Quo(&block.TotalDifficulty, trillion)
	b.HeadTotalDifficultyTrillions.Set(float64(divided.Uint64()))

	return nil
}

func (b *BlockMetrics) getSafeBlockStats(ctx context.Context) error {
	mostRecentBlockNumber, err := b.client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	newSafeBlockNumber := mostRecentBlockNumber - b.safeDistanceBlocks
	if newSafeBlockNumber < 1 {
		return errors.New("safe block does not exist yet")
	}

	// If the new safest block has already been recorded, no-op and wait a little longer.
	if newSafeBlockNumber == b.currentSafeBlockNumber {
		b.log.WithField("new", newSafeBlockNumber).Info("No-op safe block")
		return nil
	}

	// Skip to the most recent safe block if we don't have a safe block yet.
	if b.currentSafeBlockNumber == 0 {
		b.currentSafeBlockNumber = newSafeBlockNumber
	}

	// Loop over all blocks that we haven't recorded for.
	for newSafeBlockNumber < b.currentSafeBlockNumber {
		if err := b.getSafeBlock(ctx, newSafeBlockNumber); err != nil {
			b.log.WithError(err).WithField("block_number", newSafeBlockNumber).Error("failed to get block stats for safe block")
		}

		b.currentSafeBlockNumber++
	}

	b.MostRecentBlockNumber.WithLabelValues("safe").Set(float64(b.currentSafeBlockNumber))

	return nil
}

func (b *BlockMetrics) getSafeBlock(ctx context.Context, number uint64) error {
	b.log.WithField("block_number", fmt.Sprintf("%d", number)).Info("getting safe block stats")

	block, err := b.client.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		return err
	}

	b.SafeBlockSize.Add(float64(block.Size()))
	b.SafeBaseFeePerGas.Add(float64(block.BaseFee().Int64()))
	b.SafeGasLimit.Add(float64(block.GasLimit()))
	b.SafeGasUsed.Add(float64(block.GasUsed()))
	b.SafeTransactionCount.Add(float64(len(block.Transactions())))

	b.log.WithField("block_number", fmt.Sprintf("%d", number)).Info("got safe block stats")

	return nil
}
