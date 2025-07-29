package cmd

import (
	"context"
	"os"
	"time"

	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ethereum-metrics-exporter",
	Short: "A tool to export the state of ethereum nodes",
	Run: func(cmd *cobra.Command, args []string) {
		initCommon()

		err := export.Serve(cmd.Context(), metricsPort)
		if err != nil {
			logr.Fatal(err)
		}
	},
}

var (
	metricsPort          int
	cfgFile              string
	config               *exporter.Config //nolint:deadcode // False positive
	export               exporter.Exporter
	ctx                  context.Context
	logr                 logrus.FieldLogger
	executionURL         string
	consensusURL         string
	monitoredDirectories []string
	executionModules     []string
	diskUsageInterval    string
)

const (
	DefaultMetricsPort = 9090
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ethereum-metrics-exporter.yaml)")
	rootCmd.PersistentFlags().IntVarP(&metricsPort, "metrics-port", "", DefaultMetricsPort, "Port to serve Prometheus metrics on")
	rootCmd.PersistentFlags().StringVarP(&executionURL, "execution-url", "", "", "(optional) URL to the execution node")
	rootCmd.PersistentFlags().StringVarP(&consensusURL, "consensus-url", "", "", "(optional) URL to the consensus node")
	rootCmd.PersistentFlags().StringSliceVarP(&monitoredDirectories, "monitored-directories", "", []string{}, "(optional) directories to monitor for disk usage")
	rootCmd.PersistentFlags().StringSliceVarP(&executionModules, "execution-modules", "", []string{}, "(optional) execution modules that are enabled on the node")
	rootCmd.PersistentFlags().StringVar(&diskUsageInterval, "disk-usage-interval", "", "(optional) interval for disk usage metrics collection (e.g. 1h, 5m, 30s)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func loadConfigFromFile(file string) (*exporter.Config, error) {
	if file == "" {
		return exporter.DefaultConfig(), nil
	}

	config := exporter.DefaultConfig()

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func initCommon() {
	ctx = context.Background()
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	logr = log

	log.WithField("cfgFile", cfgFile).Info("Loading config")

	config, err := loadConfigFromFile(cfgFile)
	if err != nil {
		logr.Fatal(err)
	}

	if executionURL != "" {
		config.Execution.Enabled = true
		config.Execution.URL = executionURL
	}

	if consensusURL != "" {
		config.Consensus.Enabled = true
		config.Consensus.URL = consensusURL
	}

	if len(monitoredDirectories) > 0 {
		config.DiskUsage.Enabled = true
		config.DiskUsage.Directories = monitoredDirectories
	}

	if len(executionModules) > 0 {
		config.Execution.Modules = executionModules
	}

	if diskUsageInterval != "" {
		duration, err := time.ParseDuration(diskUsageInterval)
		if err != nil {
			logr.WithError(err).Fatalf("Invalid disk usage interval format: %s", diskUsageInterval)
		}

		config.DiskUsage.Interval.Duration = duration
		config.DiskUsage.Enabled = true
	}

	export = exporter.NewExporter(log, config)
	if err := export.Init(ctx); err != nil {
		logrus.Fatal(err)
	}
}
