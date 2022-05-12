package execution

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
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
	name        string
	url         string
	client      *ethclient.Client
	internalApi api.ExecutionClient
	log         logrus.FieldLogger
	metrics     Metrics
}

func NewExecutionNode(ctx context.Context, log logrus.FieldLogger, namespace string, nodeName string, url string, enabledModules []string) (*node, error) {
	internalApi := api.NewExecutionClient(ctx, log, url)
	client, _ := ethclient.Dial(url)
	metrics := NewMetrics(client, internalApi, log, nodeName, namespace, enabledModules)

	node := &node{
		name:        nodeName,
		url:         url,
		log:         log,
		internalApi: internalApi,
		client:      client,
		metrics:     metrics,
	}

	return node, nil
}

func (e *node) Name() string {
	return e.name
}

func (e *node) URL() string {
	return e.url
}

func (e *node) Bootstrapped() bool {
	return e.client != nil
}

func (e *node) Bootstrap(ctx context.Context) error {
	client, err := ethclient.Dial(e.url)
	if err != nil {
		return err
	}

	e.client = client
	return nil
}

func (e *node) StartMetrics(ctx context.Context) {
	for !e.Bootstrapped() {
		e.Bootstrap(ctx)

		time.Sleep(5 * time.Second)
	}

	e.metrics.StartAsync(ctx)
}
