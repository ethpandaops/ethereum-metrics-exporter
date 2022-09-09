package exporter

import (
	"fmt"

	"github.com/savid/ethereum-balance-metrics-exporter/pkg/exporter/jobs"
)

// Config holds the configuration for the ethereum sync status tool.
type Config struct {
	GlobalConfig GlobalConfig `yaml:"global"`
	// Execution is the execution node to use.
	Execution ExecutionNode `yaml:"execution"`
	// Addresses is the list of addresses to monitor.
	Addresses Addresses `yaml:"addresses"`
}

type GlobalConfig struct {
	LoggingLevel string            `yaml:"logging" default:"warn"`
	MetricsAddr  string            `yaml:"metricsAddr" default:":9090"`
	Namespace    string            `yaml:"namespace" default:"eth_address"`
	Labels       map[string]string `yaml:"labels"`
}

// ExecutionNode represents a single ethereum execution client.
type ExecutionNode struct {
	URL string `yaml:"url" default:"http://localhost:8545"`
}

type Addresses struct {
	EOA               []*jobs.AddressEOA               `yaml:"eoa"`
	ERC20             []*jobs.AddressERC20             `yaml:"erc20"`
	ERC721            []*jobs.AddressERC721            `yaml:"erc721"`
	ERC1155           []*jobs.AddressERC1155           `yaml:"erc1155"`
	UniswapPair       []*jobs.AddressUniswapPair       `yaml:"uniswapPair"`
	ChainlinkDataFeed []*jobs.AddressChainlinkDataFeed `yaml:"chainlinkDataFeed"`
}

func (c *Config) Validate() error {
	// Check that all addresses have different names
	duplicates := make(map[string]struct{})
	for _, u := range c.Addresses.EOA {
		// Check that all addresses have different names
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate eoa addresses with the same name: %s", u.Name)
		}

		duplicates[u.Name] = struct{}{}
	}

	duplicates = make(map[string]struct{})
	for _, u := range c.Addresses.ERC20 {
		// Check that all addresses have different names
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate erc20 addresses with the same name: %s", u.Name)
		}

		duplicates[u.Name] = struct{}{}
	}

	duplicates = make(map[string]struct{})
	for _, u := range c.Addresses.ERC721 {
		// Check that all addresses have different names
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate erc721 addresses with the same name: %s", u.Name)
		}

		duplicates[u.Name] = struct{}{}
	}

	duplicates = make(map[string]struct{})
	for _, u := range c.Addresses.ERC1155 {
		// Check that all addresses have different names
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate erc1155 addresses with the same name: %s", u.Name)
		}

		duplicates[u.Name] = struct{}{}
	}

	duplicates = make(map[string]struct{})
	for _, u := range c.Addresses.UniswapPair {
		// Check that all addresses have different names
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate uniswap pair addresses with the same name: %s", u.Name)
		}

		duplicates[u.Name] = struct{}{}
	}

	duplicates = make(map[string]struct{})
	for _, u := range c.Addresses.ChainlinkDataFeed {
		// Check that all addresses have different names
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate chainlink data feed addresses with the same name: %s", u.Name)
		}

		duplicates[u.Name] = struct{}{}
	}

	return nil
}
