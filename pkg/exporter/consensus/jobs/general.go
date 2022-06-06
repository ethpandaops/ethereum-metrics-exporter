package jobs

import (
	"context"
	"errors"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// General reports general information about the node.
type General struct {
	client      eth2client.Service
	log         logrus.FieldLogger
	Slots       prometheus.GaugeVec
	NodeVersion prometheus.GaugeVec
	ClientName  prometheus.GaugeVec
}

const (
	NameGeneral = "general"
)

// NewGeneral creates a new General instance.
func NewGeneralJob(client eth2client.Service, log logrus.FieldLogger, namespace string, constLabels map[string]string) General {
	constLabels["module"] = NameGeneral

	return General{
		client: client,
		log:    log,
		Slots: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "slot_number",
				Help:        "The slot number of the beacon chain.",
				ConstLabels: constLabels,
			},
			[]string{
				"identifier",
			},
		),
		NodeVersion: *prometheus.NewGaugeVec(
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
	}
}

func (g *General) Name() string {
	return NameGeneral
}

func (g *General) Start(ctx context.Context) {
	g.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			g.tick(ctx)
		}
	}
}

func (g *General) HandleEvent(ctx context.Context, event *v1.Event) {

}

func (g *General) tick(ctx context.Context) {
	if err := g.GetNodeVersion(ctx); err != nil {
		g.log.WithError(err).Error("Failed to get node version")
	}

	checkpoints := []string{"head", "justified", "finalized"}

	for _, checkpoint := range checkpoints {
		if err := g.GetBeaconSlot(ctx, checkpoint); err != nil {
			g.log.WithError(err).Error("Failed to get beacon slot: ", checkpoint)
		}
	}
}

func (g *General) GetNodeVersion(ctx context.Context) error {
	provider, isProvider := g.client.(eth2client.NodeVersionProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.NodeVersionProvider")
	}

	version, err := provider.NodeVersion(ctx)
	if err != nil {
		return err
	}

	g.NodeVersion.Reset()
	g.NodeVersion.WithLabelValues(version).Set(1)

	return nil
}

func (g *General) GetBeaconSlot(ctx context.Context, identifier string) error {
	provider, isProvider := g.client.(eth2client.BeaconBlockHeadersProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.BeaconBlockHeadersProvider")
	}

	block, err := provider.BeaconBlockHeader(ctx, identifier)
	if err != nil {
		return err
	}

	if block == nil {
		return errors.New("block is nil")
	}

	if block.Header == nil {
		return errors.New("block header is nil")
	}

	if block.Header.Message == nil {
		return errors.New("block header message is nil")
	}

	g.ObserveSlot(identifier, uint64(block.Header.Message.Slot))

	return nil
}

func (g *General) ObserveSlot(identifier string, slot uint64) {
	g.Slots.WithLabelValues(identifier).Set(float64(slot))
}
