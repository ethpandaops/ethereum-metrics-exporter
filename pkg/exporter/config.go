package exporter

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
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
}

// ExecutionNode represents a single ethereum execution client.
type ExecutionNode struct {
	Enabled bool     `yaml:"enabled"`
	Name    string   `yaml:"name"`
	URL     string   `yaml:"url"`
	Modules []string `yaml:"modules"`
}

// DiskUsage configures the exporter to expose disk usage stats for these directories.
type DiskUsage struct {
	Enabled     bool     `yaml:"enabled"`
	Directories []string `yaml:"directories"`
}

// PairConfig holds the config for a Pair of Execution and Consensus Clients
type PairConfig struct {
	Enabled bool `yaml:"enabled"`
}

// DefaultConfig represents a sane-default configuration.
func DefaultConfig() *Config {
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
		},
		DiskUsage: DiskUsage{
			Enabled:     false,
			Directories: []string{},
		},
		Pair: PairConfig{
			Enabled: true,
		},
	}
}
