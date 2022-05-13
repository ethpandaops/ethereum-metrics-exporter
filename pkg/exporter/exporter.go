package exporter

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
}

func (e *exporter) Init(ctx context.Context) error {
	e.log.Info("Initializing...")

	namespace := "eth"

	if e.config.Consensus.Enabled {
		e.log.Info("Initializing consensus...")
		consensus, err := consensus.NewConsensusNode(ctx, e.log.WithField("exporter", "consensus"), fmt.Sprintf("%s_con", namespace), e.config.Consensus.Name, e.config.Consensus.URL)
		if err != nil {
			return err
		}

		consensus.Bootstrap(ctx)

		e.consensus = consensus
	}

	if e.config.Execution.Enabled {
		e.log.WithField("modules", strings.Join(e.config.Execution.Modules, ", ")).Info("Initializing execution...")
		execution, err := execution.NewExecutionNode(ctx, e.log.WithField("exporter", "execution"), fmt.Sprintf("%s_exe", namespace), e.config.Execution.Name, e.config.Execution.URL, e.config.Execution.Modules)
		if err != nil {
			return err
		}

		execution.Bootstrap(ctx)

		e.execution = execution
	}

	if e.config.DiskUsage.Enabled {
		e.log.Info("Initializing disk usage...")
		diskUsage, err := disk.NewDiskUsage(ctx, e.log.WithField("exporter", "disk"), fmt.Sprintf("%s_disk", namespace), e.config.DiskUsage.Directories)
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
		go e.execution.StartMetrics(ctx)
	}

	if e.config.DiskUsage.Enabled {
		go e.diskUsage.StartAsync(ctx)
	}

	if e.config.Consensus.Enabled {
		go e.consensus.StartMetrics(ctx)
	}

	e.log.
		WithField("consensus_url", e.consensus.URL()).
		WithField("execution_url", e.execution.URL()).
		Info(fmt.Sprintf("Starting metrics server on :%v", port))

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	return err
}
