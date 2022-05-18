package execution

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/onrik/ethrpc"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api"
	"github.com/sirupsen/logrus"
)

// Node represents an execution node.
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
	name         string
	url          string
	client       *ethclient.Client
	internalAPI  api.ExecutionClient
	ethrpcClient *ethrpc.EthRPC
	log          logrus.FieldLogger
	metrics      Metrics
}

// NewExecutionNode returns a new execution node.
func NewExecutionNode(ctx context.Context, log logrus.FieldLogger, namespace, nodeName, url string, enabledModules []string) (Node, error) {
	internalAPI := api.NewExecutionClient(ctx, log, url)
	client, _ := ethclient.Dial(url)
	ethrpcClient := ethrpc.New(url)
	metrics := NewMetrics(client, internalAPI, ethrpcClient, log, nodeName, namespace, enabledModules)

	node := &node{
		name:         nodeName,
		url:          url,
		log:          log,
		ethrpcClient: ethrpcClient,
		internalAPI:  internalAPI,
		client:       client,
		metrics:      metrics,
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
		if err := e.Bootstrap(ctx); err != nil {
			e.log.WithError(err).Error("Failed to bootstrap node")
		}

		time.Sleep(5 * time.Second)
	}

	e.metrics.StartAsync(ctx)
}
