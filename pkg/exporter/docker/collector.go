package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

const (
	volumeTypeBind    = "bind"
	volumeTypeTmpfs   = "tmpfs"
	volumeTypeNamed   = "named"
	dockerVolumesPath = "/var/lib/docker/volumes"
)

// VolumeConfig represents volume monitoring configuration
type VolumeConfig struct {
	Name    string `yaml:"name"`    // Volume name or identifier
	Path    string `yaml:"path"`    // Mount path in container
	Type    string `yaml:"type"`    // Volume type: "named", "bind", "tmpfs"
	Monitor bool   `yaml:"monitor"` // Whether to collect usage metrics
}

// FilesystemConfig defines filesystem monitoring settings
type FilesystemConfig struct {
	Enabled  bool `yaml:"enabled"`  // Enable container filesystem monitoring
	Interval int  `yaml:"interval"` // Collection interval in seconds (0 = same as container stats)
}

type containerCollector struct {
	client *client.Client
	log    logrus.FieldLogger
}

func newCollector(dockerClient *client.Client, log logrus.FieldLogger) *containerCollector {
	return &containerCollector{
		client: dockerClient,
		log:    log,
	}
}

func (c *containerCollector) getContainerStats(ctx context.Context, containerID string) (*types.StatsJSON, error) {
	stats, err := c.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var v types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *containerCollector) findContainer(ctx context.Context, nameOrID string) (*types.Container, error) {
	containers, err := c.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for i := range containers {
		ctr := &containers[i]
		// Check by ID (both full and short)
		if strings.HasPrefix(ctr.ID, nameOrID) {
			return ctr, nil
		}

		// Check by name
		for _, name := range ctr.Names {
			// Remove leading '/' from container name
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == nameOrID {
				return ctr, nil
			}
		}
	}

	return nil, nil // Container not found
}

func (c *containerCollector) isContainerRunning(ctr *types.Container) bool {
	return ctr.State == "running"
}

// getContainerFilesystemUsage retrieves container filesystem usage via Docker inspect
func (c *containerCollector) getContainerFilesystemUsage(ctx context.Context, containerID string) (*FilesystemUsage, error) {
	// Use ContainerInspectWithRaw with size calculation to get filesystem usage
	containerJSON, _, err := c.client.ContainerInspectWithRaw(ctx, containerID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	totalBytes := uint64(0)
	if containerJSON.SizeRootFs != nil {
		totalBytes = uint64(*containerJSON.SizeRootFs)
	}

	writableBytes := uint64(0)
	if containerJSON.SizeRw != nil {
		writableBytes = uint64(*containerJSON.SizeRw)
	}

	readOnlyBytes := uint64(0)
	if totalBytes > writableBytes {
		readOnlyBytes = totalBytes - writableBytes
	}

	return &FilesystemUsage{
		TotalBytes:    totalBytes,
		WritableBytes: writableBytes,
		ReadOnlyBytes: readOnlyBytes,
	}, nil
}

// getContainerVolumeUsage collects usage stats for all configured volumes
func (c *containerCollector) getContainerVolumeUsage(ctx context.Context, containerID string, volumeConfigs []VolumeConfig, filesystemInterval time.Duration) ([]VolumeUsage, []VolumeInfo, error) {
	// Use ContainerInspectWithRaw but don't need size calculation for volumes (saves overhead)
	containerJSON, _, err := c.client.ContainerInspectWithRaw(ctx, containerID, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to inspect container for volumes %s: %w", containerID, err)
	}

	volumes := parseContainerVolumes(containerJSON, volumeConfigs)
	volumeUsages := make([]VolumeUsage, 0, len(volumes))
	monitoredVolumes := make([]VolumeInfo, 0, len(volumes))

	// Configure cache timeout based on filesystem interval (cache for 80% of interval)
	cacheTimeout := time.Duration(float64(filesystemInterval) * 0.8)
	if cacheTimeout < time.Minute {
		cacheTimeout = time.Minute // Minimum 1-minute cache
	}

	globalVolumeCache.setCacheTimeout(cacheTimeout)

	c.log.WithFields(logrus.Fields{
		"container_id":    containerID[:12],
		"total_volumes":   len(volumes),
		"config_provided": len(volumeConfigs) > 0,
		"cache_timeout":   cacheTimeout,
	}).Info("Starting volume usage collection")

	for _, volume := range volumes {
		if !volume.Monitor {
			c.log.WithFields(logrus.Fields{
				"volume_name": volume.Name,
				"reason":      "monitoring disabled",
			}).Debug("Skipping volume")

			continue
		}

		c.log.WithFields(logrus.Fields{
			"volume_name": volume.Name,
			"volume_type": volume.Type,
			"source_path": volume.Source,
			"mount_path":  volume.Target,
		}).Debug("Processing volume for usage metrics")

		var usagePath string

		switch volume.Type {
		case volumeTypeBind:
			usagePath = volume.Source
		case volumeTypeNamed:
			// Docker volumes are typically at /var/lib/docker/volumes/<name>/_data
			standardPath := filepath.Join(dockerVolumesPath, volume.Name, "_data")
			usagePath = resolveVolumeUsagePath(volume.Name, standardPath)
		case volumeTypeTmpfs:
			// For tmpfs, monitor the mount point inside container's mount namespace
			// This requires special handling - skip for now as it's memory-based
			continue
		default:
			c.log.WithField("volume_type", volume.Type).Warn("Unknown volume type, skipping")
			continue
		}

		c.log.WithFields(logrus.Fields{
			"volume_name": volume.Name,
			"usage_path":  usagePath,
		}).Debug("Attempting to collect filesystem usage stats")

		usage, err := getVolumeUsage(usagePath)
		if err != nil {
			c.log.WithError(err).WithFields(logrus.Fields{
				"volume_name": volume.Name,
				"usage_path":  usagePath,
			}).Warn("Failed to get volume usage (volume may not be accessible from host)")

			continue
		}

		c.log.WithFields(logrus.Fields{
			"volume_name":     volume.Name,
			"total_bytes":     usage.TotalBytes,
			"used_bytes":      usage.UsedBytes,
			"available_bytes": usage.AvailableBytes,
		}).Info("Successfully collected volume usage metrics")

		volumeUsages = append(volumeUsages, *usage)
		monitoredVolumes = append(monitoredVolumes, volume)
	}

	c.log.WithFields(logrus.Fields{
		"container_id":       containerID[:12],
		"total_volumes":      len(volumes),
		"monitored_volumes":  len(monitoredVolumes),
		"successful_metrics": len(volumeUsages),
	}).Info("Completed volume usage collection")

	return volumeUsages, monitoredVolumes, nil
}

// isContainerInspectRequired checks if we need inspect API call for this container
func (c *containerCollector) isContainerInspectRequired(volumeConfigs []VolumeConfig) bool {
	return len(volumeConfigs) > 0 // Need inspect if monitoring volumes
}
