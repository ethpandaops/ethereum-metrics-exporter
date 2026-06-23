package jobs

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// AmsterdamMetrics tracks Amsterdam-fork EIP-specific Prometheus metrics.
//
// EIP-7843: slotNumber embedded in the execution payload header.
// EIP-7928: blockAccessListHash in the execution payload header.
// EIP-7778: gas refund delta = block.gasUsed − sum(receipt.gasUsed).
//           This value is always non-negative in Amsterdam because refunds are
//           credited to the originating address directly (not deducted from
//           gasUsed in receipts).
type AmsterdamMetrics struct {
	client       *ethclient.Client
	api          api.ExecutionClient
	ethRPCClient *ethrpc.EthRPC
	log          logrus.FieldLogger

	// HeadSlotNumber is the CL slot number embedded in the latest EL block header (EIP-7843).
	HeadSlotNumber prometheus.Gauge

	// HeadBALHashPresent is 1 when blockAccessListHash is non-zero (EIP-7928).
	// A zero hash indicates no BAL data (pre-Amsterdam or empty access list).
	HeadBALHashPresent prometheus.Gauge

	// HeadGasRefundDelta is block.gasUsed minus the sum of all receipt.gasUsed values (EIP-7778).
	// In Amsterdam, source-based refunds are added back to the originating address and NOT
	// subtracted from receipt gasUsed, so the delta is always ≥ 0.
	HeadGasRefundDelta prometheus.Gauge

	currentHeadBlockNumber uint64
}

const (
	NameAmsterdam = "amsterdam"

	// zeroBALHash is the all-zeros hash sentinel for "no BAL data".
	zeroBALHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

func (a *AmsterdamMetrics) Name() string {
	return NameAmsterdam
}

func (a *AmsterdamMetrics) RequiredModules() []string {
	return []string{"eth"}
}

// NewAmsterdamMetrics returns a new AmsterdamMetrics instance.
func NewAmsterdamMetrics(client *ethclient.Client, internalAPI api.ExecutionClient, ethRPCClient *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string) AmsterdamMetrics {
	constLabels["module"] = NameAmsterdam
	namespace = namespace + "_" + NameAmsterdam

	return AmsterdamMetrics{
		client:       client,
		api:          internalAPI,
		ethRPCClient: ethRPCClient,
		log:          log.WithField("module", NameAmsterdam),

		HeadSlotNumber: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_slot_number",
				Help:        "The CL slot number embedded in the latest EL block header (EIP-7843 SLOTNUM opcode).",
				ConstLabels: constLabels,
			},
		),
		HeadBALHashPresent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_bal_hash_present",
				Help:        "1 when the latest block carries a non-zero blockAccessListHash (EIP-7928), 0 otherwise.",
				ConstLabels: constLabels,
			},
		),
		HeadGasRefundDelta: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_gas_refund_delta",
				Help:        "Gas units refunded to originating addresses: block.gasUsed minus sum(receipt.gasUsed) (EIP-7778).",
				ConstLabels: constLabels,
			},
		),
	}
}

func (a *AmsterdamMetrics) Start(ctx context.Context) {
	a.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
			a.tick(ctx)
		}
	}
}

func (a *AmsterdamMetrics) tick(ctx context.Context) {
	headNumber, err := a.client.BlockNumber(ctx)
	if err != nil {
		a.log.WithError(err).Error("Amsterdam: failed to get head block number")
		return
	}

	if headNumber == a.currentHeadBlockNumber {
		return
	}

	a.currentHeadBlockNumber = headNumber

	block, err := a.api.GetAmsterdamBlock(ctx, "latest")
	if err != nil {
		a.log.WithError(err).Error("Amsterdam: failed to get Amsterdam block fields")
		return
	}

	if block == nil {
		// Pre-Amsterdam block; skip silently.
		return
	}

	// EIP-7843 slot number.
	if block.SlotNumber != "" {
		slot, err := hexutil.DecodeUint64(block.SlotNumber)
		if err != nil {
			a.log.WithError(err).Error("Amsterdam: failed to parse slotNumber")
		} else {
			a.HeadSlotNumber.Set(float64(slot))
		}
	}

	// EIP-7928 BAL hash presence.
	balPresent := 0.0
	if block.BlockAccessListHash != "" &&
		!strings.EqualFold(block.BlockAccessListHash, zeroBALHash) {
		balPresent = 1.0
	}

	a.HeadBALHashPresent.Set(balPresent)

	// EIP-7778 gas refund delta.
	if block.GasUsed != "" {
		blockGasUsed, ok := new(big.Int).SetString(strings.TrimPrefix(block.GasUsed, "0x"), 16)
		if !ok {
			a.log.Error("Amsterdam: failed to parse block gasUsed")
			return
		}

		receipts, err := a.api.GetBlockReceipts(ctx, "latest")
		if err != nil {
			a.log.WithError(err).Error("Amsterdam: failed to get block receipts for gas refund delta")
			return
		}

		receiptSum := new(big.Int)

		for _, r := range receipts {
			if r == nil || r.GasUsed == "" {
				continue
			}

			rGas, ok := new(big.Int).SetString(strings.TrimPrefix(r.GasUsed, "0x"), 16)
			if !ok {
				continue
			}

			receiptSum.Add(receiptSum, rGas)
		}

		delta := new(big.Int).Sub(blockGasUsed, receiptSum)
		if delta.Sign() >= 0 {
			a.HeadGasRefundDelta.Set(float64(delta.Int64()))
		}
	}
}
