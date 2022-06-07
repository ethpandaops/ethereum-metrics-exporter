package jobs

import (
	"context"
	"errors"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/sirupsen/logrus"
)

// Beacon reports Beacon information about the beacon chain.
type Beacon struct {
	client              eth2client.Service
	log                 logrus.FieldLogger
	Slot                prometheus.GaugeVec
	Transactions        prometheus.GaugeVec
	Slashings           prometheus.GaugeVec
	Attestations        prometheus.GaugeVec
	Deposits            prometheus.GaugeVec
	VoluntaryExits      prometheus.GaugeVec
	FinalityCheckpoints prometheus.GaugeVec
	ReOrgs              prometheus.Counter
	ReOrgDepth          prometheus.Counter

	currentVersion string
}

const (
	NameBeacon = "beacon"
)

// NewBeacon creates a new Beacon instance.
func NewBeaconJob(client eth2client.Service, ap api.ConsensusClient, log logrus.FieldLogger, namespace string, constLabels map[string]string) Beacon {
	constLabels["module"] = NameBeacon
	namespace += "_beacon"

	return Beacon{
		client: client,
		log:    log,
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
	}
}

func (b *Beacon) Name() string {
	return NameBeacon
}

func (b *Beacon) Start(ctx context.Context) {
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

func (b *Beacon) tick(ctx context.Context) {
	for _, id := range []string{"head", "finalized", "justified"} {
		if id != "justified" {
			if err := b.GetFinality(ctx, id); err != nil {
				b.log.WithError(err).Error("Failed to get finality")
			}
		}

		if err := b.GetSignedBeaconBlock(ctx, id); err != nil {
			b.log.WithError(err).Error("Failed to get signed beacon block")
		}
	}
}

func (b *Beacon) HandleEvent(ctx context.Context, event *v1.Event) {
	if event.Topic == EventTopicBlock {
		if err := b.GetSignedBeaconBlock(ctx, "head"); err != nil {
			b.log.WithError(err).Error("Failed to get signed beacon block")
		}
	}

	if event.Topic == EventTopicChainReorg {
		b.handleChainReorg(event)
	}
}

func (b *Beacon) handleChainReorg(event *v1.Event) {
	reorg, ok := event.Data.(*v1.ChainReorgEvent)
	if !ok {
		return
	}

	b.ReOrgs.Inc()
	b.ReOrgDepth.Add(float64(reorg.Depth))
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

	return nil
}

func (b *Beacon) handleSingleBlock(blockID string, block *spec.VersionedSignedBeaconBlock) error {
	if b.currentVersion != block.Version.String() {
		b.Transactions.Reset()
		b.Slashings.Reset()
		b.Attestations.Reset()
		b.Deposits.Reset()
		b.VoluntaryExits.Reset()

		b.currentVersion = block.Version.String()
	}

	var beaconBlock BeaconBlock

	switch block.Version {
	case spec.DataVersionPhase0:
		beaconBlock = NewBeaconBlockFromPhase0(block)
	case spec.DataVersionAltair:
		beaconBlock = NewBeaconBlockFromAltair(block)
	case spec.DataVersionBellatrix:
		beaconBlock = NewBeaconBlockFromBellatrix(block)
	default:
		return errors.New("received beacon block of unknown spec version")
	}

	b.recordNewBeaconBlock(blockID, block.Version.String(), beaconBlock)

	return nil
}

func (b *Beacon) recordNewBeaconBlock(blockID, version string, block BeaconBlock) {
	b.Slot.WithLabelValues(blockID, version).Set(float64(block.Slot))
	b.Slashings.WithLabelValues(blockID, version, "proposer").Set(float64(block.ProposerSlashings))
	b.Slashings.WithLabelValues(blockID, version, "attester").Set(float64(block.ProposerSlashings))
	b.Attestations.WithLabelValues(blockID, version).Set(float64(block.Attestations))
	b.Deposits.WithLabelValues(blockID, version).Set(float64(block.Deposits))
	b.VoluntaryExits.WithLabelValues(blockID, version).Set(float64(block.VoluntaryExits))
	b.Transactions.WithLabelValues(blockID, version).Set(float64(block.Transactions))
}
