package exporter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution"
	"github.com/sirupsen/logrus"
)

type Ethereum interface {
	Init(ctx context.Context) error
	Config(ctx context.Context) *Config
	Serve(ctx context.Context, port int) error
	GetSyncStatus(ctx context.Context) (*SyncStatus, error)
}

func NewEthereum(log logrus.FieldLogger, conf *Config) Ethereum {
	return &ethereum{
		log:    log.WithField("component", "exporter"),
		config: conf,
	}
}

type ethereum struct {
	log       logrus.FieldLogger
	config    *Config
	consensus consensus.Node
	execution execution.Node
	metrics   Metrics
}

func (e *ethereum) Init(ctx context.Context) error {
	e.log.Info("Initializing...")
	e.metrics = NewMetrics(e.config.Execution.Name, e.config.Consensus.Name, "")
	e.log.Info("metrics done")

	if e.config.Consensus.Enabled {
		consensus, err := consensus.NewConsensusNode(ctx, e.log, e.config.Consensus.Name, e.config.Consensus.URL, e.metrics.Consensus())
		if err != nil {
			return err
		}

		e.consensus = consensus
	}

	if e.config.Execution.Enabled {
		execution, err := execution.NewExecutionNode(ctx, e.log, e.config.Execution.Name, e.config.Execution.URL, e.metrics.Execution())
		if err != nil {
			return err
		}

		e.execution = execution
	}

	return nil
}

func (e *ethereum) Config(ctx context.Context) *Config {
	return e.config
}

func (e *ethereum) ticker(ctx context.Context) {
	for {
		e.Tick(ctx)
		time.Sleep(time.Second * time.Duration(e.config.PollingFrequencySeconds))
	}
}

func (e *ethereum) Serve(ctx context.Context, port int) error {
	go e.ticker(ctx)
	e.log.Info(fmt.Sprintf("Starting metrics server on :%v", port))

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	return err
}

func (e *ethereum) Tick(ctx context.Context) {
	if _, err := e.GetSyncStatus(ctx); err != nil {
		e.log.Error(err)
	}
}

func (e *ethereum) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	status := &SyncStatus{}
	consensus, err := e.consensus.SyncStatus(ctx)
	if err == nil {
		status.Consensus = consensus
	} else {
		e.log.WithError(err).Error("Failed to fetch consensus client sync status")
	}

	execution, err := e.execution.SyncStatus(ctx)
	if err == nil {
		status.Execution = execution
	} else {
		e.log.WithError(err).Error("Failed to fetch execution client sync status")
	}

	return status, nil
}
