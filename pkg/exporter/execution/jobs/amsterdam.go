package jobs

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
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
// EIP-2780: TX_BASE reduced from 21000 to 12000. Detected via eth_estimateGas
//           on a zero-value self-transfer (no calldata, no code). Returns 12000
//           if EIP-2780 is implemented, 21000 otherwise.
// EIP-7708: ETH transfers emit a Transfer(address,address,uint256) log from the
//           system address 0xffffff...fe. HeadTransferLogCount counts these per block.
type AmsterdamMetrics struct {
	client       *ethclient.Client
	api          api.ExecutionClient
	ethRPCClient *ethrpc.EthRPC
	log          logrus.FieldLogger

	// HeadSlotNumber is the CL slot number embedded in the latest EL block header (EIP-7843).
	HeadSlotNumber prometheus.Gauge

	// HeadBALHashPresent is 1 when the latest block carries a non-zero blockAccessListHash (EIP-7928).
	HeadBALHashPresent prometheus.Gauge

	// HeadGasRefundDelta is block.gasUsed minus the sum of all receipt.gasUsed values (EIP-7778).
	HeadGasRefundDelta prometheus.Gauge

	// TxBaseGas is the observed TX_BASE intrinsic gas (EIP-2780).
	// 12000 = EIP-2780 implemented; 21000 = pre-EIP-2780 (not yet implemented).
	// Detected via eth_estimateGas on a zero-value self-transfer from a zero-balance address.
	TxBaseGas prometheus.Gauge

	// HeadTransferLogCount is the number of EIP-7708 Transfer logs emitted in the latest block.
	// EIP-7708 emits Transfer(address,address,uint256) from 0xffffff...fe on every ETH transfer.
	// A value of 0 indicates either no ETH transfers in the block or EIP-7708 not implemented.
	HeadTransferLogCount prometheus.Gauge

	// HeadMaxTxGas is the highest gas_limit field of any transaction in the latest block (EIP-7825).
	// Amsterdam note: tx.gas is NOT capped at TX_MAX_GAS_LIMIT (2^24=16,777,216) in the mempool.
	// Execution budget is capped at TX_MAX_GAS_LIMIT internally. A value > 16,777,216 means the block
	// contains txs whose effective execution gas is capped (not an error — normal Amsterdam behaviour).
	// Set to 0 when the block contains no transactions.
	HeadMaxTxGas prometheus.Gauge

	currentHeadBlockNumber uint64
	txBaseDetected         bool
}

const (
	NameAmsterdam = "amsterdam"

	// zeroBALHash is the all-zeros hash sentinel for "no BAL data".
	zeroBALHash = "0x0000000000000000000000000000000000000000000000000000000000000000"

	// eip7708SystemAddress is the source address for EIP-7708 ETH Transfer logs.
	eip7708SystemAddress = "0xfffffffffffffffffffffffffffffffffffffffe"

	// eip7708TransferTopic is keccak256("Transfer(address,address,uint256)").
	eip7708TransferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
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
		TxBaseGas: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "tx_base_gas",
				Help:        "Observed TX_BASE intrinsic gas (EIP-2780): 12000 if EIP-2780 is implemented, 21000 otherwise. Detected via eth_estimateGas on a zero-value self-transfer.",
				ConstLabels: constLabels,
			},
		),
		HeadTransferLogCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_transfer_log_count",
				Help:        "Number of EIP-7708 Transfer logs emitted in the latest block (from system address 0xffffff...fe). 0 means no ETH transfers or EIP-7708 not implemented.",
				ConstLabels: constLabels,
			},
		),
		HeadMaxTxGas: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_max_tx_gas",
				Help:        "Highest gas_limit field of any transaction in the latest block (EIP-7825). In Amsterdam, tx.gas is NOT capped; the execution budget is capped at TX_MAX_GAS_LIMIT=16,777,216 (2^24). Values > 16,777,216 indicate txs whose execution is capped. 0 when the block has no transactions.",
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

	// EIP-7825 max tx gas limit in this block.
	// When EIP-7825 is enforced, no tx may have gas > 16,000,000.
	// This gauge lets operators detect non-compliance (value > 16M) in real-time.
	maxTxGas := uint64(0)
	for _, tx := range block.Transactions {
		if tx == nil || tx.Gas == "" {
			continue
		}
		g, err := hexutil.DecodeUint64(tx.Gas)
		if err != nil {
			continue
		}
		if g > maxTxGas {
			maxTxGas = g
		}
	}
	a.HeadMaxTxGas.Set(float64(maxTxGas))

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
		transferLogCount := 0

		for _, r := range receipts {
			if r == nil {
				continue
			}

			if r.GasUsed != "" {
				rGas, ok := new(big.Int).SetString(strings.TrimPrefix(r.GasUsed, "0x"), 16)
				if ok {
					receiptSum.Add(receiptSum, rGas)
				}
			}

			// Count EIP-7708 Transfer logs from the system address.
			for _, l := range r.Logs {
				if l == nil {
					continue
				}
				if !strings.EqualFold(l.Address, eip7708SystemAddress) {
					continue
				}
				if len(l.Topics) > 0 && strings.EqualFold(l.Topics[0], eip7708TransferTopic) {
					transferLogCount++
				}
			}
		}

		a.HeadTransferLogCount.Set(float64(transferLogCount))

		delta := new(big.Int).Sub(blockGasUsed, receiptSum)
		if delta.Sign() >= 0 {
			a.HeadGasRefundDelta.Set(float64(delta.Int64()))
		}
	}

	// EIP-2780 TX_BASE detection (once per restart; value is stable once set).
	// Call eth_estimateGas on a zero-value self-transfer from address zero.
	// A zero-balance address sending value=0 requires no funds, so the estimate
	// succeeds and returns exactly TX_BASE (12000 with EIP-2780, 21000 without).
	if !a.txBaseDetected {
		zeroAddr := common.Address{}
		toAddr := zeroAddr
		gas, err := a.client.EstimateGas(ctx, ethereum.CallMsg{
			From:  zeroAddr,
			To:    &toAddr,
			Value: big.NewInt(0),
		})
		if err != nil {
			a.log.WithError(err).Debug("Amsterdam: EIP-2780 TX_BASE probe failed")
		} else {
			a.TxBaseGas.Set(float64(gas))
			a.txBaseDetected = true
			if gas == 12000 {
				a.log.Info("Amsterdam: EIP-2780 TX_BASE=12000 detected (EIP-2780 implemented)")
			} else {
				a.log.WithField("tx_base", gas).Info("Amsterdam: pre-EIP-2780 TX_BASE detected (EIP-2780 not yet implemented)")
			}
		}
	}
}
