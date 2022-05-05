package cmd

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ethereum-metrics-exporter",
	Short: "A tool to report the sync status of ethereum nodes",
	Run: func(cmd *cobra.Command, args []string) {
		initCommon()

		err := ethClient.Serve(ctx, metricsPort)
		if err != nil {
			logr.Fatal(err)
		}
	},
}

var (
	cfgFile      string
	config       *exporter.Config
	ethClient    exporter.Ethereum
	ctx          context.Context
	logr         logrus.FieldLogger
	executionUrl string
	consensusUrl string
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ethereum-metrics-exporter.yaml)")

	rootCmd.PersistentFlags().IntVarP(&metricsPort, "metrics-port", "", 9090, "Port to serve Prometheus metrics on")
	rootCmd.PersistentFlags().StringVarP(&executionUrl, "execution-url", "", "", "(optional) URL to the execution node")
	rootCmd.PersistentFlags().StringVarP(&consensusUrl, "consensus-url", "", "", "(optional) URL to the consensus node")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func loadConfigFromFile(file string) (*exporter.Config, error) {
	if file == "" {
		return exporter.DefaultConfig(), nil
	}

	var config exporter.Config
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	return &config, nil
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

	if executionUrl != "" {
		config.Execution.Enabled = true
		config.Execution.URL = executionUrl
	}

	if consensusUrl != "" {
		config.Consensus.Enabled = true
		config.Consensus.URL = consensusUrl
	}

	ethClient = exporter.NewEthereum(log, config)
	if err := ethClient.Init(ctx); err != nil {
		logrus.Fatal(err)
	}
}
