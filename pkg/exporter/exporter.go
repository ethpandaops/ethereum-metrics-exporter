package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	ehttp "github.com/attestantio/go-eth2-client/http"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/beacon"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/disk"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/pair"
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
	consensus   consensus.Metrics
	execution   execution.Node
	diskUsage   disk.UsageMetrics
	pairMetrics pair.Metrics

	// Nats
	broker     *server.Server
	brokerConn *nats.EncodedConn

	// Clients
	beacon beacon.Node
	client eth2client.Service
	api    api.ConsensusClient
}

func (e *exporter) Init(ctx context.Context) error {
	e.log.Info("Initializing...")

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

	e.brokerConn = conn

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

	if e.config.Pair.Enabled && e.config.Execution.Enabled && e.config.Consensus.Enabled {
		if err := e.ensureConsensusClients(ctx); err != nil {
			e.log.Fatal(err)
		}

		if _, err := e.beacon.OnReady(ctx, func(ctx context.Context, event *beacon.ReadyEvent) error {
			e.pairMetrics.StartAsync(ctx)
			return nil
		}); err != nil {
			e.log.WithError(err).Error("Failed to subscribe to beacon node ready event")
		}

		if err := e.startPairExporter(ctx); err != nil {
			e.log.WithError(err).Error("failed to start pair metrics")

			e.log.Fatal(err)
		}
	}

	if e.config.Consensus.Enabled {
		if err := e.ensureConsensusClients(ctx); err != nil {
			e.log.Fatal(err)
		}

		if err := e.startConsensusExporter(ctx); err != nil {
			e.log.WithError(err).Error("failed to start consensus")

			e.log.Fatal(err)
		}

		if _, err := e.beacon.OnReady(ctx, func(ctx context.Context, event *beacon.ReadyEvent) error {
			e.consensus.StartAsync(ctx)

			return nil
		}); err != nil {
			e.log.WithError(err).Error("Failed to subscribe to beacon node ready event")
		}

		e.beacon.StartAsync(ctx)
	}

	return nil
}

func (e *exporter) bootstrapConsensusClients(ctx context.Context) error {
	client, err := ehttp.New(ctx,
		ehttp.WithAddress(e.config.Consensus.URL),
		ehttp.WithLogLevel(zerolog.Disabled),
	)
	if err != nil {
		return err
	}

	e.client = client
	e.api = api.NewConsensusClient(ctx, e.log, e.config.Consensus.URL)
	e.beacon = beacon.NewNode(ctx, e.log, e.api, e.client, e.brokerConn)

	return nil
}

func (e *exporter) ensureConsensusClients(ctx context.Context) error {
	for {
		if e.client != nil {
			_, isProvider := e.client.(eth2client.NodeSyncingProvider)
			if isProvider {
				break
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(1 * time.Second)

			if err := e.bootstrapConsensusClients(ctx); err != nil {
				e.log.WithError(err).Error("failed to bootstrap consensus node")

				continue
			}

			break
		}
	}

	return nil
}

func (e *exporter) startConsensusExporter(ctx context.Context) error {
	if err := e.ensureConsensusClients(ctx); err != nil {
		return err
	}

	e.log.Info("Starting consensus metrics...")

	conMetrics := consensus.NewMetrics(e.client, e.api, e.beacon, e.log.WithField("exporter", "consensus"), e.config.Consensus.Name, fmt.Sprintf("%s_con", e.namespace))

	e.consensus = conMetrics

	return nil
}

func (e *exporter) startPairExporter(ctx context.Context) error {
	if err := e.ensureConsensusClients(ctx); err != nil {
		return err
	}

	pairMetrics, err := pair.NewMetrics(ctx, e.log.WithField("exporter", "pair"), fmt.Sprintf("%s_pair", e.namespace), e.beacon, e.config.Execution.URL)
	if err != nil {
		return err
	}

	e.pairMetrics = pairMetrics

	return nil
}
