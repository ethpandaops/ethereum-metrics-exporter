package docker

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/prometheus/client_golang/prometheus"
)

type LabelConfig struct {
	IncludeContainerName bool `yaml:"containerName"`
	IncludeContainerID   bool `yaml:"containerID"`
	IncludeImageName     bool `yaml:"imageName"`
	IncludeImageTag      bool `yaml:"imageTag"`
}

type backoffState struct {
	currentDelay time.Duration
	maxDelay     time.Duration
	minDelay     time.Duration
}

func newBackoffState() *backoffState {
	return &backoffState{
		currentDelay: 10 * time.Second,
		maxDelay:     60 * time.Second,
		minDelay:     10 * time.Second,
	}
}

func (b *backoffState) nextDelay() time.Duration {
	delay := b.currentDelay
	b.currentDelay *= 2

	if b.currentDelay > b.maxDelay {
		b.currentDelay = b.maxDelay
	}

	return delay
}

func (b *backoffState) reset() {
	b.currentDelay = b.minDelay
}

func buildPrometheusLabels(container *types.Container, labelConfig LabelConfig, containerType string) prometheus.Labels {
	labels := prometheus.Labels{}

	// Always include container type
	labels["type"] = containerType

	if labelConfig.IncludeContainerName && len(container.Names) > 0 {
		// Remove leading '/' from container name
		name := strings.TrimPrefix(container.Names[0], "/")
		labels["container_name"] = name
	}

	if labelConfig.IncludeContainerID {
		labels["container_id"] = container.ID[:12] // Short ID
	}

	if labelConfig.IncludeImageName {
		parts := strings.SplitN(container.Image, ":", 2)
		labels["image_name"] = parts[0]
	}

	if labelConfig.IncludeImageTag {
		parts := strings.SplitN(container.Image, ":", 2)
		if len(parts) > 1 {
			labels["image_tag"] = parts[1]
		} else {
			labels["image_tag"] = "latest"
		}
	}

	// Ensure we have at least container name label to match metric definitions
	// This must match the fallback logic in newMetrics()
	if !labelConfig.IncludeContainerName && len(container.Names) > 0 {
		name := strings.TrimPrefix(container.Names[0], "/")
		labels["container_name"] = name
	}

	return labels
}
