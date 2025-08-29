package docker

import (
	"context"
	"time"

	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/filesystem"
)

// ContainerInfo holds container configuration with metadata.
type ContainerInfo struct {
	Name       string
	Type       string
	Volumes    []VolumeConfig
	Filesystem FilesystemConfig
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
	fsMonitor  filesystem.Monitor
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

	// Initialize metrics and filesystem monitor
	metrics := newMetrics(namespace, labels)
	collector := newCollector(dockerClient, log)
	backoff := newBackoffState()

	// Create filesystem monitor with intelligent caching
	fsConfig := &filesystem.MonitorConfig{
		CacheConfig: filesystem.DefaultCacheConfig(),
	}
	fsMonitor := filesystem.NewMonitor(fsConfig, log)

	return &containerMetrics{
		log:        log,
		client:     dockerClient,
		containers: containers,
		interval:   interval,
		labels:     labels,
		metrics:    metrics,
		collector:  collector,
		backoff:    backoff,
		fsMonitor:  fsMonitor,
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

	// Collect filesystem metrics if enabled
	if containerInfo.Filesystem.Enabled {
		c.log.WithField("container", containerInfo.Name).Debug("Collecting container filesystem metrics")

		if err := c.collectFilesystemMetrics(ctx, container.ID, labels); err != nil {
			c.log.WithError(err).WithField("container", containerInfo.Name).Warn("Failed to collect filesystem metrics")
		}
	}

	// Collect volume metrics (auto-discover if no volumes configured)
	c.log.WithFields(logrus.Fields{
		"container":      containerInfo.Name,
		"volume_configs": len(containerInfo.Volumes),
		"auto_discovery": len(containerInfo.Volumes) == 0,
	}).Debug("Initiating volume metrics collection")

	// Calculate filesystem interval from container config (default to 5 minutes if not set)
	filesystemInterval := 5 * time.Minute
	if containerInfo.Filesystem.Interval > 0 {
		filesystemInterval = time.Duration(containerInfo.Filesystem.Interval) * time.Second
	}

	if err := c.collectVolumeMetrics(ctx, container.ID, containerInfo.Volumes, filesystemInterval, labels); err != nil {
		c.log.WithError(err).WithField("container", containerInfo.Name).Warn("Failed to collect volume metrics")
	}

	return nil
}

// collectFilesystemMetrics collects filesystem usage metrics for a container
func (c *containerMetrics) collectFilesystemMetrics(ctx context.Context, containerID string, labels prometheus.Labels) error {
	usage, err := c.collector.getContainerFilesystemUsage(ctx, containerID)
	if err != nil {
		return err
	}

	c.metrics.updateFilesystemMetrics(usage, labels)

	return nil
}

// collectVolumeMetrics collects volume usage metrics for configured volumes
func (c *containerMetrics) collectVolumeMetrics(ctx context.Context, containerID string, volumeConfigs []VolumeConfig, filesystemInterval time.Duration, labels prometheus.Labels) error {
	volumeUsages, volumes, err := c.collector.getContainerVolumeUsage(ctx, containerID, volumeConfigs, c.fsMonitor)
	if err != nil {
		return err
	}

	c.metrics.updateVolumeMetrics(volumeUsages, volumes, labels)
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
