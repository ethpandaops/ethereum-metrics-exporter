package docker

import (
	"context"
	"time"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

// ContainerInfo holds container configuration with metadata.
type ContainerInfo struct {
	Name string
	Type string
}

type ContainerMetrics interface {
	StartAsync(ctx context.Context)
}

type containerMetrics struct {
	log        logrus.FieldLogger
	client     *client.Client
	containers []ContainerInfo
	interval   time.Duration
	labels     LabelConfig
	metrics    *metrics
	collector  *containerCollector
	backoff    *backoffState
}

func NewContainerMetrics(ctx context.Context, log logrus.FieldLogger, namespace string, containers []ContainerInfo, endpoint string, interval time.Duration, labels LabelConfig) (ContainerMetrics, error) {
	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.WithHost(endpoint), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	// Test connection to Docker daemon
	_, err = dockerClient.Ping(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	metrics := newMetrics(namespace, labels)
	collector := newCollector(dockerClient, log)
	backoff := newBackoffState()

	return &containerMetrics{
		log:        log,
		client:     dockerClient,
		containers: containers,
		interval:   interval,
		labels:     labels,
		metrics:    metrics,
		collector:  collector,
		backoff:    backoff,
	}, nil
}

func (c *containerMetrics) StartAsync(ctx context.Context) {
	// Initial collection
	c.collectMetrics(ctx)

	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				c.log.Info("Docker metrics collection stopped")
				return
			case <-ticker.C:
				c.collectMetrics(ctx)
			}
		}
	}()
}

func (c *containerMetrics) collectMetrics(ctx context.Context) {
	for _, containerInfo := range c.containers {
		if err := c.collectContainerMetrics(ctx, containerInfo); err != nil {
			delay := c.handleDockerError(err)
			c.log.WithError(err).WithField("container", containerInfo.Name).WithField("retry_delay", delay).Error("Failed to collect container metrics")
		} else {
			c.backoff.reset() // Reset backoff on successful collection
		}
	}
}

func (c *containerMetrics) collectContainerMetrics(ctx context.Context, containerInfo ContainerInfo) error {
	// Find container
	container, err := c.collector.findContainer(ctx, containerInfo.Name)
	if err != nil {
		return err
	}

	if container == nil {
		c.log.WithField("container", containerInfo.Name).Debug("Container not found")
		return nil // Container doesn't exist, skip silently
	}

	if !c.collector.isContainerRunning(container) {
		c.log.WithField("container", containerInfo.Name).Debug("Container is not running")
		return nil // Container is not running, skip silently
	}

	// Get container stats
	stats, err := c.collector.getContainerStats(ctx, container.ID)
	if err != nil {
		return err
	}

	// Build labels based on configuration, including the container type
	labels := buildPrometheusLabels(container, c.labels, containerInfo.Type)

	// Update metrics
	containerName := containerInfo.Name
	if len(container.Names) > 0 {
		containerName = container.Names[0]
	}

	c.metrics.updateContainerMetrics(containerName, stats, labels)

	return nil
}

func (c *containerMetrics) handleDockerError(err error) time.Duration {
	delay := c.backoff.nextDelay()

	// Schedule retry after delay
	go func() {
		time.Sleep(delay)
	}()

	return delay
}
