package exporter

import (
	"time"

	"github.com/ethpandaops/beacon/pkg/human"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/docker"
)

// Config holds the configuration for the ethereum sync status tool.
type Config struct {
	// Execution is the execution node to use.
	Execution ExecutionNode `yaml:"execution"`
	// ConsensusNodes is the consensus node to use.
	Consensus ConsensusNode `yaml:"consensus"`
	// DiskUsage determines if the disk usage metrics should be exported.
	DiskUsage DiskUsage `yaml:"diskUsage"`
	// Docker determines if the docker container metrics should be exported.
	Docker DockerConfig `yaml:"docker"`
	// Pair determines if the pair metrics should be exported.
	Pair PairConfig `yaml:"pair"`
}

// ConsensusNode represents a single ethereum consensus client.
type ConsensusNode struct {
	Enabled     bool        `yaml:"enabled"`
	Name        string      `yaml:"name"`
	URL         string      `yaml:"url"`
	EventStream EventStream `yaml:"eventStream"`
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
}

// DiskUsage configures the exporter to expose disk usage stats for these directories.
type DiskUsage struct {
	Enabled     bool           `yaml:"enabled"`
	Directories []string       `yaml:"directories"`
	Interval    human.Duration `yaml:"interval"`
}

// VolumeConfig defines volume monitoring configuration
// If no volumes are specified, all container volumes will be auto-discovered and monitored.
// Use "*" as name to apply settings to all discovered volumes.
type VolumeConfig struct {
	Name    string `yaml:"name"`    // Volume name or identifier ("*" for all volumes)
	Path    string `yaml:"path"`    // Mount path in container (alternative matcher)
	Type    string `yaml:"type"`    // Volume type filter: "named", "bind", "tmpfs" (optional)
	Monitor bool   `yaml:"monitor"` // Whether to collect usage metrics for this volume
}

// FilesystemConfig defines filesystem monitoring settings
type FilesystemConfig struct {
	Enabled  bool           `yaml:"enabled"`  // Enable container filesystem monitoring
	Interval human.Duration `yaml:"interval"` // Collection interval (default: same as container stats)
}

// ContainerConfig defines a container to monitor with its metadata.
type ContainerConfig struct {
	Name          string                     `yaml:"name"`
	Type          string                     `yaml:"type"`
	Volumes       []VolumeConfig             `yaml:"volumes,omitempty"`        // Volume monitoring configuration (empty = auto-discover all)
	Filesystem    FilesystemConfig           `yaml:"filesystem,omitempty"`     // Filesystem monitoring settings
	PortBandwidth docker.PortBandwidthConfig `yaml:"port_bandwidth,omitempty"` // Port bandwidth monitoring settings
}

// DockerConfig configures the exporter to expose Docker container metrics.
type DockerConfig struct {
	Enabled    bool               `yaml:"enabled"`
	Endpoint   string             `yaml:"endpoint"`
	Containers []ContainerConfig  `yaml:"containers"`
	Interval   human.Duration     `yaml:"interval"`
	Labels     docker.LabelConfig `yaml:"labels"`
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
		Docker: DockerConfig{
			Enabled:    false,
			Endpoint:   "unix:///var/run/docker.sock",
			Containers: []ContainerConfig{},
			Interval:   human.Duration{Duration: 10 * time.Second},
			Labels: docker.LabelConfig{
				IncludeContainerName: true,
				IncludeContainerID:   false,
				IncludeImageName:     false,
				IncludeImageTag:      false,
			},
		},
		Pair: PairConfig{
			Enabled: true,
		},
	}
}
