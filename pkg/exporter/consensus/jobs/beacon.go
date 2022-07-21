package jobs

import (
	"context"
	"errors"
	"hash/fnv"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon"
	"github.com/sirupsen/logrus"
)

// Beacon reports Beacon information about the beacon chain.
type Beacon struct {
	client                 eth2client.Service
	log                    logrus.FieldLogger
	beaconNode             beacon.Node
	Slot                   prometheus.GaugeVec
	Transactions           prometheus.GaugeVec
	Slashings              prometheus.GaugeVec
	Attestations           prometheus.GaugeVec
	Deposits               prometheus.GaugeVec
	VoluntaryExits         prometheus.GaugeVec
	FinalityCheckpoints    prometheus.GaugeVec
	ReOrgs                 prometheus.Counter
	ReOrgDepth             prometheus.Counter
	HeadSlotHash           prometheus.Gauge
	FinalityCheckpointHash prometheus.GaugeVec
	EmptySlots             prometheus.Counter
	ProposerDelay          prometheus.Histogram
	currentVersion         string
}

const (
	NameBeacon = "beacon"
	// NumRootHashShards defines the range of values that the *hash metrics are moduloed by. That is to say,
	// this number defines the value range of those metrics from 0 -> NumRootHashShards.
	NumRootHashShards = 65536
)

// NewBeacon creates a new Beacon instance.
func NewBeaconJob(client eth2client.Service, ap api.ConsensusClient, beac beacon.Node, log logrus.FieldLogger, namespace string, constLabels map[string]string) Beacon {
	constLabels["module"] = NameBeacon
	namespace += "_beacon"

	return Beacon{
		client:     client,
		beaconNode: beac,
		log:        log,
		Slot: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "slot",
				Help:        "The slot number in the block.",
				ConstLabels: constLabels,
			},
			[]string{
				"block_id",
				"version",
			},
		),
		Transactions: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "transactions",
				Help:        "The amount of transactions in the block.",
				ConstLabels: constLabels,
			},
			[]string{
				"block_id",
				"version",
			},
		),
		Slashings: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "slashings",
				Help:        "The amount of slashings in the block.",
				ConstLabels: constLabels,
			},
			[]string{
				"block_id",
				"version",
				"type",
			},
		),
		Attestations: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "attestations",
				Help:        "The amount of attestations in the block.",
				ConstLabels: constLabels,
			},
			[]string{
				"block_id",
				"version",
			},
		),
		Deposits: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "deposits",
				Help:        "The amount of deposits in the block.",
				ConstLabels: constLabels,
			},
			[]string{
				"block_id",
				"version",
			},
		),
		VoluntaryExits: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "voluntary_exits",
				Help:        "The amount of voluntary exits in the block.",
				ConstLabels: constLabels,
			},
			[]string{
				"block_id",
				"version",
			},
		),
		FinalityCheckpoints: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "finality_checkpoint_epochs",
				Help:        "That epochs of the finality checkpoints.",
				ConstLabels: constLabels,
			},
			[]string{
				"state_id",
				"checkpoint",
			},
		),
		ReOrgs: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "reorg_count",
				Help:        "The count of reorgs.",
				ConstLabels: constLabels,
			},
		),
		ReOrgDepth: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "reorg_depth",
				Help:        "The number of reorgs.",
				ConstLabels: constLabels,
			},
		),
		HeadSlotHash: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "head_slot_hash",
				Help:        "The hash of the head slot (ranges from 0-15).",
				ConstLabels: constLabels,
			},
		),
		FinalityCheckpointHash: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "finality_checkpoint_hash",
				Help:        "The hash of the finality checkpoint.",
				ConstLabels: constLabels,
			},
			[]string{
				"state_id",
				"checkpoint",
			},
		),
		ProposerDelay: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Name:        "proposer_delay",
				Help:        "The delay of the proposer.",
				ConstLabels: constLabels,
				Buckets:     prometheus.LinearBuckets(0, 1000, 13),
			},
		),
		EmptySlots: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "empty_slots_count",
				Help:        "The number of slots that have expired without a block proposed.",
				ConstLabels: constLabels,
			},
		),
	}
}

func (b *Beacon) Name() string {
	return NameBeacon
}

func (b *Beacon) Start(ctx context.Context) error {
	b.tick(ctx)

	if err := b.setupSubscriptions(ctx); err != nil {
		return err
	}

	go b.getInitialData(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 5):
			b.tick(ctx)
		}
	}
}

func (b *Beacon) tick(ctx context.Context) {

}

func (b *Beacon) setupSubscriptions(ctx context.Context) error {
	if _, err := b.beaconNode.OnBlockInserted(ctx, b.handleBlockInserted); err != nil {
		return err
	}

	if _, err := b.beaconNode.OnChainReOrg(ctx, b.handleChainReorg); err != nil {
		return err
	}

	if _, err := b.beaconNode.OnEmptySlot(ctx, b.handleEmptySlot); err != nil {
		return err
	}

	if _, err := b.beaconNode.OnFinalizedCheckpoint(ctx, b.handleFinalizedCheckpointEvent); err != nil {
		return err
	}

	return nil
}

func (b *Beacon) handleEmptySlot(ctx context.Context, event *beacon.EmptySlotEvent) error {
	b.log.WithField("slot", event.Slot).Debug("Empty slot detected")

	b.EmptySlots.Inc()

	return nil
}

func (b *Beacon) handleBlockInserted(ctx context.Context, event *beacon.BlockInsertedEvent) error {
	// Fetch the slot
	slot, err := b.beaconNode.GetSlot(ctx, event.Slot)
	if err != nil {
		return err
	}

	timedBlock, err := slot.Block()
	if err != nil {
		return err
	}

	// nolint:gocritic // False positive
	if err = b.handleSingleBlock("head", timedBlock.Block); err != nil {
		return err
	}

	delay, err := slot.ProposerDelay()
	if err != nil {
		return err
	}

	b.ProposerDelay.Observe(float64(delay.Milliseconds()))

	return nil
}

func (b *Beacon) getInitialData(ctx context.Context) {
	for {
		if b.client == nil {
			time.Sleep(time.Second * 5)
			continue
		}

		b.updateFinalizedCheckpoint(ctx)

		break
	}
}

func (b *Beacon) handleChainReorg(ctx context.Context, event *v1.ChainReorgEvent) error {
	b.ReOrgs.Inc()
	b.ReOrgDepth.Add(float64(event.Depth))

	return nil
}

func (b *Beacon) handleFinalizedCheckpointEvent(ctx context.Context, event *v1.FinalizedCheckpointEvent) error {
	b.updateFinalizedCheckpoint(ctx)

	return nil
}

func (b *Beacon) updateFinalizedCheckpoint(ctx context.Context) {
	if err := b.GetFinality(ctx, "head"); err != nil {
		b.log.WithError(err).Error("Failed to get finality")
	}

	if err := b.GetSignedBeaconBlock(ctx, "finalized"); err != nil {
		b.log.WithError(err).Error("Failed to get signed beacon block")
	}
}

func (b *Beacon) GetSignedBeaconBlock(ctx context.Context, blockID string) error {
	provider, isProvider := b.client.(eth2client.SignedBeaconBlockProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.SignedBeaconBlockProvider")
	}

	signedBeaconBlock, err := provider.SignedBeaconBlock(ctx, blockID)
	if err != nil {
		return err
	}

	if err := b.handleSingleBlock(blockID, signedBeaconBlock); err != nil {
		return err
	}

	return nil
}

func (b *Beacon) GetFinality(ctx context.Context, stateID string) error {
	provider, isProvider := b.client.(eth2client.FinalityProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.FinalityProvider")
	}

	finality, err := provider.Finality(ctx, stateID)
	if err != nil {
		return err
	}

	b.FinalityCheckpoints.
		WithLabelValues(stateID, "previous_justified").
		Set(float64(finality.PreviousJustified.Epoch))

	b.FinalityCheckpoints.
		WithLabelValues(stateID, "justified").
		Set(float64(finality.Justified.Epoch))

	b.FinalityCheckpoints.
		WithLabelValues(stateID, "finalized").
		Set(float64(finality.Finalized.Epoch))

	b.recordFinalityCheckpointHash(stateID, finality)

	return nil
}

func (b *Beacon) handleSingleBlock(blockID string, block *spec.VersionedSignedBeaconBlock) error {
	if block == nil {
		return errors.New("block is nil")
	}

	if b.currentVersion != block.Version.String() {
		b.Transactions.Reset()
		b.Slashings.Reset()
		b.Attestations.Reset()
		b.Deposits.Reset()
		b.VoluntaryExits.Reset()
		b.Slot.Reset()

		b.currentVersion = block.Version.String()
	}

	b.recordNewBeaconBlock(blockID, block)

	return nil
}

func (b *Beacon) recordNewBeaconBlock(blockID string, block *spec.VersionedSignedBeaconBlock) {
	version := block.Version.String()

	slot, err := block.Slot()
	if err != nil {
		b.log.WithError(err).WithField("block_id", blockID).Error("Failed to get slot from block")
	} else {
		b.Slot.WithLabelValues(blockID, version).Set(float64(slot))
	}

	attesterSlashing, err := block.AttesterSlashings()
	if err != nil {
		b.log.WithError(err).WithField("block_id", blockID).Error("Failed to get attester slashing from block")
	} else {
		b.Slashings.WithLabelValues(blockID, version, "attester").Set(float64(len(attesterSlashing)))
	}

	proposerSlashing, err := block.ProposerSlashings()
	if err != nil {
		b.log.WithError(err).WithField("block_id", blockID).Error("Failed to get proposer slashing from block")
	} else {
		b.Slashings.WithLabelValues(blockID, version, "proposer").Set(float64(len(proposerSlashing)))
	}

	attestations, err := block.Attestations()
	if err != nil {
		b.log.WithError(err).WithField("block_id", blockID).Error("Failed to get attestations from block")
	} else {
		b.Attestations.WithLabelValues(blockID, version).Set(float64(len(attestations)))
	}

	deposits := GetDepositCountsFromBeaconBlock(block)
	b.Deposits.WithLabelValues(blockID, version).Set(float64(deposits))

	voluntaryExits := GetVoluntaryExitsFromBeaconBlock(block)
	b.VoluntaryExits.WithLabelValues(blockID, version).Set(float64(voluntaryExits))

	transactions := GetTransactionsCountFromBeaconBlock(block)
	b.Transactions.WithLabelValues(blockID, version).Set(float64(transactions))

	if blockID == "head" {
		stateRoot, err := block.StateRoot()
		if err != nil {
			b.log.WithError(err).Error("Failed to get state root for head block")

			return
		}

		compressedHash := float64(b.getModulo(string(stateRoot[:])))

		b.log.WithField("compressed_hash", compressedHash).Debug("Calculated state root of head block")

		b.HeadSlotHash.Set(compressedHash)
	}
}

func (b *Beacon) recordFinalityCheckpointHash(stateID string, finality *v1.Finality) {
	finalized := float64(b.getModulo(string(finality.Finalized.Root[:])))

	justified := float64(b.getModulo(string(finality.Justified.Root[:])))

	previousJustified := float64(b.getModulo(string(finality.PreviousJustified.Root[:])))

	b.log.WithFields(logrus.Fields{
		"state_id":           stateID,
		"finalized":          finalized,
		"justified":          justified,
		"previous_justified": previousJustified,
	}).Debug("Recorded finality checkpoint hash")

	b.FinalityCheckpointHash.WithLabelValues(stateID, "finalized").Set(finalized)
	b.FinalityCheckpointHash.WithLabelValues(stateID, "justified").Set(justified)
	b.FinalityCheckpointHash.WithLabelValues(stateID, "previous_justified").Set(previousJustified)
}

func (b *Beacon) getModulo(hash string) int {
	h := fnv.New32a()
	h.Write([]byte(hash))

	return int(float64(h.Sum32() % NumRootHashShards))
}
