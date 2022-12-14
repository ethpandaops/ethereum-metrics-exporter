package exporter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/disk"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/execution"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samcm/beacon"
	"github.com/sirupsen/logrus"
)

// Exporter defines the Ethereum Metrics Exporter interface
type Exporter interface {
	// Init initialises the exporter
	Init(ctx context.Context) error
	// Config returns the configuration of the exporter
	Config(ctx context.Context) *Config
	// Serve starts the metrics server
	Serve(ctx context.Context, port int) error
}

// NewExporter returns a new Exporter instance
func NewExporter(log logrus.FieldLogger, conf *Config) Exporter {
	return &exporter{
		log:       log.WithField("component", "exporter"),
		config:    conf,
		namespace: "eth",
	}
}

type exporter struct {
	// Helpers
	namespace string
	log       logrus.FieldLogger
	config    *Config

	// Exporters
	execution execution.Node
	diskUsage disk.UsageMetrics

	// Clients
	beacon beacon.Node
}

func (e *exporter) Init(ctx context.Context) error {
	e.log.Info("Initializing...")

	if e.config.Execution.Enabled {
		e.log.WithField("modules", strings.Join(e.config.Execution.Modules, ", ")).Info("Initializing execution...")

		executionNode, err := execution.NewExecutionNode(ctx, e.log.WithField("exporter", "execution"), fmt.Sprintf("%s_exe", e.namespace), e.config.Execution.Name, e.config.Execution.URL, e.config.Execution.Modules)
		if err != nil {
			return err
		}

		if err := executionNode.Bootstrap(ctx); err != nil {
			e.log.WithError(err).Error("failed to bootstrap execution node")
		}

		e.execution = executionNode
	}

	if e.config.DiskUsage.Enabled {
		e.log.Info("Initializing disk usage...")

		diskUsage, err := disk.NewUsage(ctx, e.log.WithField("exporter", "disk"), fmt.Sprintf("%s_disk", e.namespace), e.config.DiskUsage.Directories)
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

func (e *exporter) Serve(ctx context.Context, port int) error {
	e.log.
		WithField("consensus_url", e.config.Consensus.URL).
		WithField("execution_url", e.config.Execution.URL).
		Info(fmt.Sprintf("Starting metrics server on :%v", port))

	s := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		ReadHeaderTimeout: 30 * time.Second,
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			e.log.Fatal(err)
		}
	}()

	if e.config.Execution.Enabled {
		e.log.WithField("execution_url", e.execution.URL()).Info("Starting execution metrics...")

		go e.execution.StartMetrics(ctx)
	}

	if e.config.DiskUsage.Enabled {
		e.log.Info("Starting disk usage metrics...")

		go e.diskUsage.StartAsync(ctx)
	}

	if e.config.Consensus.Enabled {
		e.log.WithField("consensus_url", e.config.Consensus.URL).Info("Starting consensus metrics...")

		if err := e.bootstrapConsensusClients(ctx); err != nil {
			e.log.WithError(err).Error("failed to bootstrap consensus clients")

			return err
		}

		go e.beacon.StartAsync(ctx)
	}

	return nil
}

func (e *exporter) bootstrapConsensusClients(ctx context.Context) error {
	opts := *beacon.DefaultOptions().
		EnableDefaultBeaconSubscription().
		EnablePrometheusMetrics().
		DisableFetchingProposerDuties()

	e.beacon = beacon.NewNode(e.log, &beacon.Config{
		Addr: e.config.Consensus.URL,
		Name: e.config.Consensus.Name,
	}, "eth_con", opts)

	return nil
}
