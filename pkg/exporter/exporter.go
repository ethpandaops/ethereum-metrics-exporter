package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/disk"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution"
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
		log:    log.WithField("component", "exporter"),
		config: conf,
	}
}

type exporter struct {
	log       logrus.FieldLogger
	config    *Config
	consensus consensus.Node
	execution execution.Node
	diskUsage disk.UsageMetrics
	broker    *server.Server
}

func (e *exporter) Init(ctx context.Context) error {
	e.log.Info("Initializing...")

	namespace := "eth"

	natsServer, err := server.NewServer(&server.Options{})
	if err != nil {
		return err
	}

	e.broker = natsServer

	// Start the nats server via goroutine
	go e.broker.Start()

	if !e.broker.ReadyForConnections(15 * time.Second) {
		return errors.New("nats server failed to start")
	}

	nc, err := nats.Connect(e.broker.ClientURL())

	if err != nil {
		return err
	}

	// Create a NATS encoded connection to the nats server
	conn, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	if e.config.Consensus.Enabled {
		e.log.Info("Initializing consensus...")

		consensusNode, err := consensus.NewConsensusNode(ctx, e.log.WithField("exporter", "consensus"), fmt.Sprintf("%s_con", namespace), e.config.Consensus.Name, e.config.Consensus.URL, conn)
		if err != nil {
			return err
		}

		if err := consensusNode.Bootstrap(ctx); err != nil {
			e.log.WithError(err).Error("failed to bootstrap consnesus node")
		}

		e.consensus = consensusNode
	}

	if e.config.Execution.Enabled {
		e.log.WithField("modules", strings.Join(e.config.Execution.Modules, ", ")).Info("Initializing execution...")

		executionNode, err := execution.NewExecutionNode(ctx, e.log.WithField("exporter", "execution"), fmt.Sprintf("%s_exe", namespace), e.config.Execution.Name, e.config.Execution.URL, e.config.Execution.Modules)
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

		diskUsage, err := disk.NewUsage(ctx, e.log.WithField("exporter", "disk"), fmt.Sprintf("%s_disk", namespace), e.config.DiskUsage.Directories)
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
	if e.config.Execution.Enabled {
		e.log.WithField("execution_url", e.execution.URL()).Info("Starting execution metrics...")

		go e.execution.StartMetrics(ctx)
	}

	if e.config.DiskUsage.Enabled {
		e.log.Info("Starting disk usage metrics...")

		go e.diskUsage.StartAsync(ctx)
	}

	if e.config.Consensus.Enabled {
		e.log.WithField("consensus_url", e.consensus.URL()).Info("Starting consensus metrics...")

		go e.consensus.StartMetrics(ctx)
	}

	e.log.Info(fmt.Sprintf("Starting metrics server on :%v", port))

	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)

	return err
}
