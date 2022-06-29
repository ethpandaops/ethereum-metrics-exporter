package consensus

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/rs/zerolog"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon"
	"github.com/sirupsen/logrus"
)

// Node represents a single consensus node in the ethereum network.
type Node interface {
	// Name returns the name of the node.
	Name() string
	// URL returns the url of the node.
	URL() string
	// Bootstrapped returns whether the node has been bootstrapped and is ready to be used.
	Bootstrapped() bool
	// Bootstrap attempts to bootstrap the node (i.e. configuring clients)
	Bootstrap(ctx context.Context) error
	// StartMetrics starts the metrics collection.
	StartMetrics(ctx context.Context)
}

type node struct {
	name       string
	url        string
	namespace  string
	client     eth2client.Service
	api        api.ConsensusClient
	beaconNode *beacon.Node
	log        logrus.FieldLogger
	metrics    Metrics
}

// NewConsensusNode returns a new Node instance.
func NewConsensusNode(ctx context.Context, log logrus.FieldLogger, namespace, name, url string) (Node, error) {
	return &node{
		name:      name,
		url:       url,
		log:       log,
		namespace: namespace,
	}, nil
}

func (c *node) Name() string {
	return c.name
}

func (c *node) URL() string {
	return c.url
}

func (c *node) Bootstrap(ctx context.Context) error {
	client, err := http.New(ctx,
		http.WithAddress(c.url),
		http.WithLogLevel(zerolog.Disabled),
	)
	if err != nil {
		return err
	}

	c.client = client
	c.api = api.NewConsensusClient(ctx, c.log, c.url)
	c.beaconNode = beacon.NewNode(ctx, c.log, c.api, c.client)

	return nil
}

func (c *node) Bootstrapped() bool {
	_, isProvider := c.client.(eth2client.NodeSyncingProvider)
	return isProvider
}

func (c *node) StartMetrics(ctx context.Context) {
	for !c.Bootstrapped() {
		if err := c.Bootstrap(ctx); err != nil {
			c.log.WithError(err).Error("Failed to bootstrap consensus client")
		}

		time.Sleep(5 * time.Second)
	}

	if err := c.beaconNode.Start(ctx); err != nil {
		c.log.WithError(err).Error("Failed to start beacon state")
	}

	c.metrics = NewMetrics(c.client, c.api, c.beaconNode, c.log, c.name, c.namespace)
	c.metrics.StartAsync(ctx)
}
