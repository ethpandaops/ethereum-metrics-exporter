package exporter

import (
	"context"

	"github.com/onrik/ethrpc"
	"github.com/savid/ethereum-balance-metrics-exporter/pkg/exporter/jobs"
	"github.com/sirupsen/logrus"
)

// Metrics exposes Execution layer metrics
type Metrics interface {
	// StartAsync starts all the metrics jobs
	StartAsync(ctx context.Context)
}

type metrics struct {
	log                      logrus.FieldLogger
	eoaMetrics               jobs.EOA
	erc20Metrics             jobs.ERC20
	erc721Metrics            jobs.ERC721
	erc1155Metrics           jobs.ERC1155
	uniswapPairMetrics       jobs.UniswapPair
	chainlinkDataFeedMetrics jobs.ChainlinkDataFeed

	enabledJobs map[string]bool
}

// NewMetrics creates a new execution Metrics instance
func NewMetrics(client *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses *Addresses) Metrics {
	m := &metrics{
		log:                      log,
		eoaMetrics:               jobs.NewEOA(client, log, namespace, constLabels, addresses.EOA),
		erc20Metrics:             jobs.NewERC20(client, log, namespace, constLabels, addresses.ERC20),
		erc721Metrics:            jobs.NewERC721(client, log, namespace, constLabels, addresses.ERC721),
		erc1155Metrics:           jobs.NewERC1155(client, log, namespace, constLabels, addresses.ERC1155),
		uniswapPairMetrics:       jobs.NewUniswapPair(client, log, namespace, constLabels, addresses.UniswapPair),
		chainlinkDataFeedMetrics: jobs.NewChainlinkDataFeed(client, log, namespace, constLabels, addresses.ChainlinkDataFeed),

		enabledJobs: make(map[string]bool),
	}

	m.log.Info("Enabling address metrics")

	if len(addresses.EOA) > 0 {
		m.enabledJobs[m.eoaMetrics.Name()] = true
	}

	if len(addresses.ERC20) > 0 {
		m.enabledJobs[m.erc20Metrics.Name()] = true
	}

	if len(addresses.ERC721) > 0 {
		m.enabledJobs[m.erc721Metrics.Name()] = true
	}

	if len(addresses.ERC1155) > 0 {
		m.enabledJobs[m.erc1155Metrics.Name()] = true
	}

	if len(addresses.UniswapPair) > 0 {
		m.enabledJobs[m.uniswapPairMetrics.Name()] = true
	}

	if len(addresses.ChainlinkDataFeed) > 0 {
		m.enabledJobs[m.chainlinkDataFeedMetrics.Name()] = true
	}

	return m
}

func (m *metrics) StartAsync(ctx context.Context) {
	if m.enabledJobs[m.eoaMetrics.Name()] {
		go m.eoaMetrics.Start(ctx)
	}

	if m.enabledJobs[m.erc20Metrics.Name()] {
		go m.erc20Metrics.Start(ctx)
	}

	if m.enabledJobs[m.erc721Metrics.Name()] {
		go m.erc721Metrics.Start(ctx)
	}

	if m.enabledJobs[m.erc1155Metrics.Name()] {
		go m.erc1155Metrics.Start(ctx)
	}

	if m.enabledJobs[m.uniswapPairMetrics.Name()] {
		go m.uniswapPairMetrics.Start(ctx)
	}

	if m.enabledJobs[m.chainlinkDataFeedMetrics.Name()] {
		go m.chainlinkDataFeedMetrics.Start(ctx)
	}

	m.log.Info("Started metrics exporter jobs")
}
