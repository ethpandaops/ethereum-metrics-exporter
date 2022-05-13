package consensus

import (
	"context"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
)

type Node interface {
	Name() string
	URL() string
	Bootstrapped() bool
	Bootstrap(ctx context.Context) error
	StartMetrics(ctx context.Context)
}

type node struct {
	name      string
	url       string
	namespace string
	client    eth2client.Service
	log       logrus.FieldLogger
	metrics   Metrics
}

func NewConsensusNode(ctx context.Context, log logrus.FieldLogger, namespace string, name string, url string) (*node, error) {
	node := &node{
		name:      name,
		url:       url,
		log:       log,
		namespace: namespace,
	}
	return node, nil
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
		http.WithLogLevel(zerolog.WarnLevel),
	)
	if err != nil {
		return err
	}

	c.client = client

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
	}

	c.metrics = NewMetrics(c.client, c.log, c.name, c.namespace)
	c.metrics.StartAsync(ctx)
}
