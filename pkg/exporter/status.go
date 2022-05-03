package exporter

import (
	"fmt"
	"strings"

	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution"
)

type SyncStatus struct {
	Consensus *consensus.SyncStatus
	Execution *execution.SyncStatus
}

func (s *SyncStatus) String() string {
	status := []string{}
	if s.Consensus != nil {
		status = append(status, fmt.Sprintf("Consensus: %v", s.Consensus.Percent()))
	}

	if s.Execution != nil {
		status = append(status, fmt.Sprintf("Execution: %v", s.Execution.Percent()))
	}

	return strings.Join(status, ", ")
}
