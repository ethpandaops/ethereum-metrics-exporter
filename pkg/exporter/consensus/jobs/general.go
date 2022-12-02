package jobs

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/beacon"
	"github.com/samcm/beacon/api/types"
	"github.com/sirupsen/logrus"
)

// General reports general information about the node.
type General struct {
	beacon      beacon.Node
	log         logrus.FieldLogger
	NodeVersion prometheus.GaugeVec
	ClientName  prometheus.GaugeVec
	Peers       prometheus.GaugeVec
	PeerAgents  prometheus.GaugeVec
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
		PeerAgents: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "peer_agents",
				Help:        "The count of peers connected to beacon node by agent.",
				ConstLabels: constLabels,
			},
			[]string{
				"agent",
			},
		),
	}
}

func (g *General) Name() string {
	return NameGeneral
}

func (g *General) Start(ctx context.Context) error {
	g.beacon.OnNodeVersionUpdated(ctx, func(ctx context.Context, event *beacon.NodeVersionUpdatedEvent) error {
		g.observeNodeVersion(ctx, event.Version)
		return nil
	})

	g.beacon.OnPeersUpdated(ctx, func(ctx context.Context, event *beacon.PeersUpdatedEvent) error {
		g.Peers.Reset()

		for _, state := range types.PeerStates {
			for _, direction := range types.PeerDirections {
				g.Peers.WithLabelValues(state, direction).Set(float64(len(event.Peers.ByStateAndDirection(state, direction))))
			}
		}

		connected := event.Peers.ByState("connected")

		for _, agent := range types.AllAgents {
			g.PeerAgents.WithLabelValues(string(agent)).Set(float64(len(connected.ByAgent(agent))))
		}

		return nil
	})

	if err := g.initialFetch(ctx); err != nil {
		return err
	}

	return nil
}

func (g *General) initialFetch(ctx context.Context) error {
	version, err := g.beacon.GetNodeVersion(ctx)
	if err != nil {
		return err
	}

	g.observeNodeVersion(ctx, version)

	return nil
}

func (g *General) observeNodeVersion(ctx context.Context, version string) {
	g.log.WithField("version", version).Debug("Got node version")

	g.NodeVersion.Reset()
	g.NodeVersion.WithLabelValues(version).Set(1)
}
