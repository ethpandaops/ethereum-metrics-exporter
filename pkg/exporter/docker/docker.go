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
	Name          string              `yaml:"name"`
	Type          string              `yaml:"type"`
	Volumes       []VolumeConfig      `yaml:"volumes"`
	Filesystem    FilesystemConfig    `yaml:"filesystem"`
	PortBandwidth PortBandwidthConfig `yaml:"port_bandwidth"`
}

type ContainerMetrics interface {
	StartAsync(ctx context.Context)
}

type containerMetrics struct {
	log           logrus.FieldLogger
	client        *client.Client
	containers    []ContainerInfo
	interval      time.Duration
	labels        LabelConfig
	metrics       *metrics
	collector     *containerCollector
	backoff       *backoffState
	fsMonitor     filesystem.Monitor
	portBandwidth PortBandwidthConfig
}

func NewContainerMetrics(ctx context.Context, log logrus.FieldLogger, namespace string, containers []ContainerInfo, endpoint string, interval time.Duration, labels LabelConfig) (ContainerMetrics, error) {
	log.WithFields(logrus.Fields{
		"container_count": len(containers),
		"endpoint":        endpoint,
		"interval":        interval,
	}).Info("Initializing Docker container metrics")

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

	// Determine port bandwidth configuration from containers
	var portBandwidthConfig *PortBandwidthConfig

	enabledContainers := 0

	for i := range containers {
		log.WithFields(logrus.Fields{
			"container_name":         containers[i].Name,
			"port_bandwidth_enabled": containers[i].PortBandwidth.Enabled,
			"monitor_all_ports":      containers[i].PortBandwidth.MonitorAllPorts,
			"protocols":              containers[i].PortBandwidth.Protocols,
		}).Info("Checking container port bandwidth configuration")

		if containers[i].PortBandwidth.Enabled {
			enabledContainers++

			if portBandwidthConfig == nil {
				portBandwidthConfig = &containers[i].PortBandwidth
			}
		}
	}

	log.WithFields(logrus.Fields{
		"total_containers":           len(containers),
		"port_monitoring_enabled":    enabledContainers,
		"port_monitoring_configured": portBandwidthConfig != nil,
	}).Info("Port bandwidth monitoring configuration summary")

	// Initialize metrics and filesystem monitor
	metrics := newMetrics(namespace, labels)
	collector := newCollector(dockerClient, log, portBandwidthConfig)
	backoff := newBackoffState()

	// Create filesystem monitor with intelligent caching
	fsConfig := &filesystem.MonitorConfig{
		CacheConfig: filesystem.DefaultCacheConfig(),
	}
	fsMonitor := filesystem.NewMonitor(fsConfig, log)

	containerMetrics := &containerMetrics{
		log:        log,
		client:     dockerClient,
		containers: containers,
		interval:   interval,
		labels:     labels,
		metrics:    metrics,
		collector:  collector,
		backoff:    backoff,
		fsMonitor:  fsMonitor,
	}

	if portBandwidthConfig != nil {
		containerMetrics.portBandwidth = *portBandwidthConfig
	}

	return containerMetrics, nil
}

func (c *containerMetrics) StartAsync(ctx context.Context) {
	// Start port monitoring if enabled
	if err := c.collector.startPortMonitoring(ctx); err != nil {
		c.log.WithError(err).Error("Failed to start port monitoring")
	}

	// Initial collection
	c.collectMetrics(ctx)

	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		defer func() {
			if err := c.collector.stopPortMonitoring(); err != nil {
				c.log.WithError(err).Error("Failed to stop port monitoring during shutdown")
			}
		}()

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

	// Add container to port monitoring and collect port bandwidth metrics if enabled
	if containerInfo.PortBandwidth.Enabled {
		c.log.WithFields(logrus.Fields{
			"container":    containerInfo.Name,
			"container_id": container.ID,
		}).Debug("Initiating port bandwidth monitoring for container")

		// Ensure container is being monitored
		if err := c.collector.addContainerToPortMonitoring(ctx, container.ID); err != nil {
			c.log.WithError(err).WithFields(logrus.Fields{
				"container":    containerInfo.Name,
				"container_id": container.ID,
			}).Warn("Failed to add container to port monitoring")
		} else {
			// Collect port bandwidth metrics
			portStats, err := c.collector.collectPortBandwidthMetrics()
			if err != nil {
				c.log.WithError(err).WithField("container", containerInfo.Name).Warn("Failed to collect port bandwidth metrics")
			} else if containerPortStats, exists := portStats[container.ID]; exists {
				c.log.WithFields(logrus.Fields{
					"container":    containerInfo.Name,
					"metric_count": len(containerPortStats),
				}).Debug("Updated port bandwidth metrics for container")
				c.metrics.updatePortBandwidthMetrics(containerName, container.ID, containerPortStats)
			}
		}
	} else {
		c.log.WithFields(logrus.Fields{
			"container": containerInfo.Name,
			"reason":    "port_bandwidth_disabled",
		}).Debug("Port bandwidth monitoring disabled for container")
	}

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
