package exporter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ethpandaops/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Exporter defines the Ethereum Metrics Exporter interface
type Exporter interface {
	// Init initialises the exporter
	Start(ctx context.Context) error
}

// NewExporter returns a new Exporter instance
func NewExporter(log logrus.FieldLogger, conf *Config) Exporter {
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	return &exporter{
		log: log.WithField("component", "exporter"),
		Cfg: conf,
	}
}

type exporter struct {
	// Helpers
	log logrus.FieldLogger
	Cfg *Config

	execution api.ExecutionClient
	// Metrics
	metrics Metrics
}

func (e *exporter) Start(ctx context.Context) error {
	e.log.Info("Initializing...")

	e.execution = api.NewExecutionClient(e.log, e.Cfg.Execution.URL, e.Cfg.Execution.Headers, e.Cfg.Execution.Timeout)

	e.metrics = NewMetrics(e.execution, e.log, e.Cfg.GlobalConfig.Namespace, e.Cfg.GlobalConfig.Labels, &e.Cfg.Addresses)

	e.log.
		WithField("execution_url", e.Cfg.Execution.URL).
		Info(fmt.Sprintf("Starting metrics server on %v", e.Cfg.GlobalConfig.MetricsAddr))

	http.Handle("/metrics", promhttp.Handler())

	if err := e.ServeMetrics(ctx); err != nil {
		return err
	}

	go e.metrics.StartAsync(ctx)

	return nil
}

func (e *exporter) ServeMetrics(ctx context.Context) error {
	go func() {
		server := &http.Server{
			Addr:              e.Cfg.GlobalConfig.MetricsAddr,
			ReadHeaderTimeout: 15 * time.Second,
		}

		server.Handler = promhttp.Handler()

		e.log.Infof("Serving metrics at %s", e.Cfg.GlobalConfig.MetricsAddr)

		if err := server.ListenAndServe(); err != nil {
			e.log.Fatal(err)
		}
	}()

	return nil
}
