package jobs

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api/types"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon"
	"github.com/sirupsen/logrus"
)

// General reports general information about the node.
type General struct {
	beacon      beacon.Node
	log         logrus.FieldLogger
	NodeVersion prometheus.GaugeVec
	ClientName  prometheus.GaugeVec
	Peers       prometheus.GaugeVec
}

const (
	NameGeneral = "general"
)

// NewGeneral creates a new General instance.
func NewGeneralJob(beac beacon.Node, log logrus.FieldLogger, namespace string, constLabels map[string]string) General {
	constLabels["module"] = NameGeneral

	return General{
		beacon: beac,
		log:    log,
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
		Peers: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "peers",
				Help:        "The count of peers connected to beacon node.",
				ConstLabels: constLabels,
			},
			[]string{
				"state",
				"direction",
			},
		),
	}
}

func (g *General) Name() string {
	return NameGeneral
}

func (g *General) Start(ctx context.Context) error {
	if _, err := g.beacon.OnNodeVersionUpdated(ctx, func(ctx context.Context, event *beacon.NodeVersionUpdatedEvent) error {
		g.log.WithField("version", event.Version).Info("Got node version")

		g.NodeVersion.Reset()
		g.NodeVersion.WithLabelValues(event.Version).Set(1)

		return nil
	}); err != nil {
		return err
	}

	if _, err := g.beacon.OnPeersUpdated(ctx, func(ctx context.Context, event *beacon.PeersUpdatedEvent) error {
		g.Peers.Reset()

		for _, state := range types.PeerStates {
			for _, direction := range types.PeerDirections {
				g.Peers.WithLabelValues(state, direction).Set(float64(len(event.Peers.ByStateAndDirection(state, direction))))
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
