package exporter

import (
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution"
)

type Metrics interface {
	Consensus() consensus.Metrics
	Execution() execution.Metrics
}

type metrics struct {
	consensus consensus.Metrics
	execution execution.Metrics
}

func NewMetrics(executionName, consensusName, namespace string) Metrics {
	return &metrics{
		consensus: consensus.NewMetrics(consensusName, namespace),
		execution: execution.NewMetrics(executionName, namespace),
	}
}

func (m *metrics) Consensus() consensus.Metrics {
	return m.consensus
}
func (m *metrics) Execution() execution.Metrics {
	return m.execution
}
