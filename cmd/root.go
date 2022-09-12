package cmd

import (
	"os"

	"github.com/creasty/defaults"
	"github.com/savid/ethereum-address-metrics-exporter/pkg/exporter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ethereum-address-metrics-exporter",
	Short: "A tool to export the ethereum address state",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := initCommon()

		export := exporter.NewExporter(log, cfg)
		if err := export.Start(cmd.Context()); err != nil {
			log.WithError(err).Fatal("failed to init")
		}
	},
}

var (
	cfgFile string
	log     = logrus.New()
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file (default is config.yaml)")
}

func loadConfigFromFile(file string) (*exporter.Config, error) {
	if file == "" {
		file = "config.yaml"
	}

	cfg := &exporter.Config{}

	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	yamlFile, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	type plain exporter.Config

	if err := yaml.Unmarshal(yamlFile, (*plain)(cfg)); err != nil {
		return nil, err
	}

	return cfg, nil
}

func initCommon() *exporter.Config {
	log.SetFormatter(&logrus.TextFormatter{})

	log.WithField("cfgFile", cfgFile).Info("loading config")

	cfg, err := loadConfigFromFile(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	logLevel, err := logrus.ParseLevel(cfg.GlobalConfig.LoggingLevel)
	if err != nil {
		log.WithField("logLevel", cfg.GlobalConfig.LoggingLevel).Fatal("invalid logging level")
	}

	log.SetLevel(logLevel)

	return cfg
}
