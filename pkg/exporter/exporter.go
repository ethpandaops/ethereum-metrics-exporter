package exporter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/disk"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution"
	"github.com/sirupsen/logrus"
)

type Exporter interface {
	Init(ctx context.Context) error
	Config(ctx context.Context) *Config
	Serve(ctx context.Context, port int) error
	GetSyncStatus(ctx context.Context) (*SyncStatus, error)
}

func NewExporter(log logrus.FieldLogger, conf *Config) Exporter {
	return &exporter{
		log:    log.WithField("component", "exporter"),
		config: conf,
	}
}

type exporter struct {
	log       logrus.FieldLogger
	config    *Config
	consensus consensus.Node
	execution execution.Node
	diskUsage disk.DiskUsage
	metrics   Metrics
}

func (e *exporter) Init(ctx context.Context) error {
	e.log.Info("Initializing...")
	e.metrics = NewMetrics(e.config.Execution.Name, e.config.Consensus.Name, "eth")
	e.log.Info("metrics done")

	if e.config.Consensus.Enabled {
		consensus, err := consensus.NewConsensusNode(ctx, e.log, e.config.Consensus.Name, e.config.Consensus.URL, e.metrics.Consensus())
		if err != nil {
			return err
		}

		consensus.Bootstrap(ctx)

		e.consensus = consensus
	}

	if e.config.Execution.Enabled {
		execution, err := execution.NewExecutionNode(ctx, e.log, e.config.Execution.Name, e.config.Execution.URL, e.metrics.Execution())
		if err != nil {
			return err
		}

		execution.Bootstrap(ctx)

		e.execution = execution
	}

	if e.config.DiskUsage.Enabled {
		diskUsage, err := disk.NewDiskUsage(ctx, e.log, e.metrics.Disk())
		if err != nil {
			return err
		}

		e.diskUsage = diskUsage
	}

	return nil
}

func (e *exporter) Config(ctx context.Context) *Config {
	return e.config
}

func (e *exporter) ticker(ctx context.Context) {
	for {
		e.Tick(ctx)
		time.Sleep(time.Second * time.Duration(e.config.PollingFrequencySeconds))
	}
}

func (e *exporter) Serve(ctx context.Context, port int) error {
	go e.ticker(ctx)
	e.log.
		WithField("consensus_url", e.consensus.URL()).
		WithField("execution_url", e.execution.URL()).
		Info(fmt.Sprintf("Starting metrics server on :%v", port))

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	return err
}

func (e *exporter) Tick(ctx context.Context) {
	if err := e.PollConsensus(ctx); err != nil {
		e.log.Error(err)
	}
	if err := e.PollExecution(ctx); err != nil {
		e.log.Error(err)
	}
	if err := e.PollDiskUsage(ctx); err != nil {
		e.log.Error(err)
	}
}

func (e *exporter) PollConsensus(ctx context.Context) error {
	if !e.config.Consensus.Enabled {
		return nil
	}

	if !e.consensus.Bootstrapped() {
		if err := e.consensus.Bootstrap(ctx); err != nil {
			return err
		}
	}

	// TODO(sam.calder-mason): Parallelize this
	if _, err := e.consensus.SyncStatus(ctx); err != nil {
		e.log.WithError(err).Error("failed to get sync status")
	}

	if _, err := e.consensus.NodeVersion(ctx); err != nil {
		e.log.WithError(err).Error("failed to get node version")
	}

	if _, err := e.consensus.Spec(ctx); err != nil {
		e.log.WithError(err).Error("failed to get chain id")
	}

	return nil
}

func (e *exporter) PollExecution(ctx context.Context) error {
	if !e.config.Execution.Enabled {
		return nil
	}

	if !e.execution.Bootstrapped() {
		if err := e.execution.Bootstrap(ctx); err != nil {
			return err
		}
	}

	// TODO(sam.calder-mason): Parallelize this
	if _, err := e.execution.SyncStatus(ctx); err != nil {
		e.log.WithError(err).Error("failed to get network id")

	}

	if _, err := e.execution.NetworkID(ctx); err != nil {
		e.log.WithError(err).Error("failed to get network id")
	}

	return nil
}

func (e *exporter) PollDiskUsage(ctx context.Context) error {
	if !e.config.DiskUsage.Enabled {
		return nil
	}

	_, err := e.diskUsage.GetUsage(ctx, e.config.DiskUsage.Directories)
	return err
}

func (e *exporter) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
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
