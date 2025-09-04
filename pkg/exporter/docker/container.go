package docker

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
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

// FilesystemUsage represents container filesystem usage stats
type FilesystemUsage struct {
	TotalBytes    uint64 // Total container filesystem size (SizeRootFs)
	WritableBytes uint64 // Writable layer size (SizeRw)
	ReadOnlyBytes uint64 // Base image layers size (TotalBytes - WritableBytes)
}

// VolumeUsage represents volume filesystem usage stats
type VolumeUsage struct {
	TotalBytes     uint64 // Total volume capacity
	UsedBytes      uint64 // Used space
	AvailableBytes uint64 // Available space
	FreeBytes      uint64 // Free space (may differ from available)
}

// VolumeInfo represents volume configuration and metadata
type VolumeInfo struct {
	Name    string // Volume name or bind mount source
	Type    string // "named", "bind", "tmpfs"
	Source  string // Host path or volume name
	Target  string // Container mount path
	Monitor bool   // Whether to collect usage metrics
}

// parseContainerVolumes extracts volume information from container inspect data
func parseContainerVolumes(containerJSON types.ContainerJSON, volumeConfigs []VolumeConfig) []VolumeInfo {
	volumes := make([]VolumeInfo, 0, len(containerJSON.Mounts))

	// If no volumes configured, auto-discover all volumes
	if len(volumeConfigs) == 0 {
		logrus.WithFields(logrus.Fields{
			"container_id": containerJSON.ID[:12],
			"mount_count":  len(containerJSON.Mounts),
		}).Debug("Auto-discovering all container volumes")

		for _, mount := range containerJSON.Mounts {
			volumeType := volumeTypeNamed
			name := mount.Name

			if name == "" {
				// For bind mounts, use the source path as name
				name = mount.Source
			}

			switch mount.Type {
			case "bind":
				volumeType = volumeTypeBind
			case "tmpfs":
				volumeType = volumeTypeTmpfs
			}

			logrus.WithFields(logrus.Fields{
				"volume_name":  name,
				"volume_type":  volumeType,
				"source_path":  mount.Source,
				"mount_path":   mount.Destination,
				"container_id": containerJSON.ID[:12],
			}).Info("Auto-discovered container volume")

			volumes = append(volumes, VolumeInfo{
				Name:    name,
				Type:    volumeType,
				Source:  mount.Source,
				Target:  mount.Destination,
				Monitor: true, // Monitor all auto-discovered volumes by default
			})
		}

		logrus.WithFields(logrus.Fields{
			"container_id":     containerJSON.ID[:12],
			"discovered_count": len(volumes),
		}).Info("Completed volume auto-discovery")

		return volumes
	}

	// Process configured volumes with filtering rules
	var wildcardConfig *VolumeConfig

	specificConfigs := make(map[string]VolumeConfig)

	// Separate wildcard config from specific configs
	for _, cfg := range volumeConfigs {
		if cfg.Name == "*" {
			wildcardConfig = &cfg
		} else {
			if cfg.Name != "" {
				specificConfigs[cfg.Name] = cfg
			}

			if cfg.Path != "" {
				specificConfigs[cfg.Path] = cfg
			}
		}
	}

	for _, mount := range containerJSON.Mounts {
		volumeType := volumeTypeNamed
		name := mount.Name

		if name == "" {
			// For bind mounts, use the source path as name
			name = mount.Source
		}

		switch mount.Type {
		case "bind":
			volumeType = volumeTypeBind
		case "tmpfs":
			volumeType = volumeTypeTmpfs
		}

		// Check if this volume has a specific configuration
		var volConfig VolumeConfig

		found := false

		// Check by mount name
		if cfg, ok := specificConfigs[mount.Name]; ok && mount.Name != "" {
			volConfig = cfg
			found = true
		}
		// Check by destination path
		if !found {
			if cfg, ok := specificConfigs[mount.Destination]; ok {
				volConfig = cfg
				found = true
			}
		}

		// If no specific config found, use wildcard config if available
		if !found && wildcardConfig != nil {
			volConfig = *wildcardConfig
			found = true
		}

		// If no config matches, skip this volume
		if !found {
			continue
		}

		// Use the actual mount name/source, not the config name
		// (config name "*" shouldn't be used as actual volume name)
		actualName := name
		if volConfig.Name != "" && volConfig.Name != "*" {
			actualName = volConfig.Name
		}

		volumes = append(volumes, VolumeInfo{
			Name:    actualName,
			Type:    volumeType,
			Source:  mount.Source,
			Target:  mount.Destination,
			Monitor: volConfig.Monitor,
		})
	}

	return volumes
}

// resolveVolumeUsagePath attempts to find the correct filesystem path for volume usage stats
// On macOS with OrbStack, Docker volumes are accessible at $HOME/OrbStack/docker/volumes/
func resolveVolumeUsagePath(volumeName, standardPath string) string {
	// First try the standard Docker path
	if _, err := os.Stat(standardPath); err == nil {
		return standardPath
	}

	// On macOS, try OrbStack path if standard path fails
	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			// For OrbStack, remove /_data suffix and use OrbStack path
			orbStackPath := filepath.Join(homeDir, "OrbStack", "docker", "volumes", volumeName)
			if _, err := os.Stat(orbStackPath); err == nil {
				logrus.WithFields(logrus.Fields{
					"volume_name":   volumeName,
					"standard_path": standardPath,
					"orbstack_path": orbStackPath,
				}).Debug("Using OrbStack volume path instead of standard Docker path")

				return orbStackPath
			}
		}
	}

	// Return standard path as fallback (will likely fail but allows for proper error logging)
	return standardPath
}
