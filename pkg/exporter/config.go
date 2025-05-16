package exporter

import (
	"time"

	"github.com/ethpandaops/beacon/pkg/human"
)

// Config holds the configuration for the ethereum sync status tool.
type Config struct {
	// Execution is the execution node to use.
	Execution ExecutionNode `yaml:"execution"`
	// ConsensusNodes is the consensus node to use.
	Consensus ConsensusNode `yaml:"consensus"`
	// DiskUsage determines if the disk usage metrics should be exported.
	DiskUsage DiskUsage `yaml:"diskUsage"`
	// Pair determines if the pair metrics should be exported.
	Pair PairConfig `yaml:"pair"`
}

// ConsensusNode represents a single ethereum consensus client.
type ConsensusNode struct {
	Enabled     bool        `yaml:"enabled"`
	Name        string      `yaml:"name"`
	URL         string      `yaml:"url"`
	EventStream EventStream `yaml:"eventStream"`
	DBPath      string      `yaml:"dbPath"` // Path to the consensus layer database
}

type EventStream struct {
	Enabled *bool    `yaml:"enabled"`
	Topics  []string `yaml:"topics"`
}

// ExecutionNode represents a single ethereum execution client.
type ExecutionNode struct {
	Enabled bool     `yaml:"enabled"`
	Name    string   `yaml:"name"`
	URL     string   `yaml:"url"`
	Modules []string `yaml:"modules"`
	DBPath  string   `yaml:"dbPath"` // Path to the execution layer database
}

// DiskUsage configures the exporter to expose disk usage stats for these directories.
type DiskUsage struct {
	Enabled     bool           `yaml:"enabled"`
	Directories []string       `yaml:"directories"`
	Interval    human.Duration `yaml:"interval"`
}

// PairConfig holds the config for a Pair of Execution and Consensus Clients
type PairConfig struct {
	Enabled bool `yaml:"enabled"`
}

// DefaultConfig represents a sane-default configuration.
func DefaultConfig() *Config {
	f := false

	return &Config{
		Execution: ExecutionNode{
			Enabled: true,
			Name:    "execution",
			URL:     "http://localhost:8545",
			Modules: []string{"eth", "net", "web3"},
		},
		Consensus: ConsensusNode{
			Enabled: true,
			Name:    "consensus",
			URL:     "http://localhost:5052",
			EventStream: EventStream{
				Enabled: &f,
				Topics:  []string{},
			},
		},
		DiskUsage: DiskUsage{
			Enabled:     false,
			Directories: []string{},
			Interval: human.Duration{
				Duration: 60 * time.Minute,
			},
		},
		Pair: PairConfig{
			Enabled: true,
		},
	}
}
