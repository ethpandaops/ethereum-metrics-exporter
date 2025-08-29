package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
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

// volumeUsageCache holds cached volume usage data with timestamps
type volumeUsageCache struct {
	mu           sync.RWMutex
	cache        map[string]cachedVolumeUsage
	cacheTimeout time.Duration
}

type cachedVolumeUsage struct {
	usage           *VolumeUsage
	timestamp       time.Time
	calculationTime time.Duration // Time it took to calculate
	fileCount       int           // Number of files found
	dynamicInterval time.Duration // Calculated optimal cache interval
}

// Global cache instance with 5-minute default timeout
var globalVolumeCache = &volumeUsageCache{
	cache:        make(map[string]cachedVolumeUsage),
	cacheTimeout: 5 * time.Minute, // Default 5-minute cache
}

// getVolumeUsage calculates directory-specific usage stats for a volume with caching
func getVolumeUsage(path string) (*VolumeUsage, error) {
	return globalVolumeCache.getOrCalculate(path)
}

// getOrCalculate retrieves cached volume usage or calculates new usage if cache expired
func (c *volumeUsageCache) getOrCalculate(path string) (*VolumeUsage, error) {
	// Check cache first
	c.mu.RLock()

	if cached, exists := c.cache[path]; exists {
		// Use dynamic interval if available, otherwise fall back to global timeout
		cacheLimit := cached.dynamicInterval
		if cacheLimit == 0 {
			cacheLimit = c.cacheTimeout
		}

		if time.Since(cached.timestamp) < cacheLimit {
			c.mu.RUnlock()
			logrus.WithFields(logrus.Fields{
				"path":             path,
				"cached_age":       time.Since(cached.timestamp),
				"cache_limit":      cacheLimit,
				"dynamic_interval": cached.dynamicInterval > 0,
				"file_count":       cached.fileCount,
				"calc_time":        cached.calculationTime,
			}).Debug("Using cached volume usage")

			return cached.usage, nil
		}
	}

	c.mu.RUnlock()

	// Cache miss or expired, calculate new usage
	logrus.WithField("path", path).Debug("Calculating fresh volume usage (cache miss or expired)")

	startTime := time.Now()
	usage, fileCount, err := c.calculateVolumeUsageWithStats(path)
	calculationTime := time.Since(startTime)

	if err != nil {
		return nil, err
	}

	// Calculate dynamic interval based on directory complexity
	dynamicInterval := c.calculateOptimalInterval(usage.UsedBytes, fileCount, calculationTime)

	// Update cache
	c.mu.Lock()
	c.cache[path] = cachedVolumeUsage{
		usage:           usage,
		timestamp:       time.Now(),
		calculationTime: calculationTime,
		fileCount:       fileCount,
		dynamicInterval: dynamicInterval,
	}
	c.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"path":             path,
		"used_bytes":       usage.UsedBytes,
		"file_count":       fileCount,
		"calc_time":        calculationTime,
		"dynamic_interval": dynamicInterval,
	}).Info("Cached fresh volume usage calculation with dynamic interval")

	return usage, nil
}

// setCacheTimeout updates the cache timeout for volume usage calculations
func (c *volumeUsageCache) setCacheTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cacheTimeout = timeout
}

// calculateOptimalInterval determines the best cache interval based on directory characteristics
func (c *volumeUsageCache) calculateOptimalInterval(sizeBytes uint64, fileCount int, calculationTime time.Duration) time.Duration {
	// Base intervals based on calculation time
	baseInterval := calculationTime * 10 // Cache for 10x the calculation time

	// Adjust based on directory size
	switch {
	case sizeBytes > 100*1024*1024*1024: // > 100GB
		baseInterval = max(baseInterval, 30*time.Minute)
	case sizeBytes > 10*1024*1024*1024: // > 10GB
		baseInterval = max(baseInterval, 15*time.Minute)
	case sizeBytes > 1024*1024*1024: // > 1GB
		baseInterval = max(baseInterval, 5*time.Minute)
	case sizeBytes > 100*1024*1024: // > 100MB
		baseInterval = max(baseInterval, 2*time.Minute)
	default: // < 100MB
		baseInterval = max(baseInterval, 1*time.Minute)
	}

	// Adjust based on file count (more files = more expensive to walk)
	switch {
	case fileCount > 100000: // > 100k files
		baseInterval = max(baseInterval, 20*time.Minute)
	case fileCount > 10000: // > 10k files
		baseInterval = max(baseInterval, 10*time.Minute)
	case fileCount > 1000: // > 1k files
		baseInterval = max(baseInterval, 3*time.Minute)
	}

	// Cap the maximum interval
	if baseInterval > time.Hour {
		baseInterval = time.Hour
	}

	// Ensure minimum interval
	if baseInterval < time.Minute {
		baseInterval = time.Minute
	}

	return baseInterval
}

// calculateVolumeUsageWithStats performs filesystem calculation and returns file count
func (c *volumeUsageCache) calculateVolumeUsageWithStats(path string) (*VolumeUsage, int, error) {
	// Get filesystem stats for available/free space
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return nil, 0, fmt.Errorf("failed to get filesystem stats for %s: %w", path, err)
	}

	blockSize := uint64(stat.Bsize)
	totalBytes := stat.Blocks * blockSize
	freeBytes := stat.Bfree * blockSize
	availableBytes := stat.Bavail * blockSize

	// Calculate directory-specific usage and count files
	usedBytes, fileCount, err := calculateDirectoryUsageWithCount(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to calculate directory usage for %s: %w", path, err)
	}

	return &VolumeUsage{
		TotalBytes:     totalBytes,     // Filesystem total (for capacity info)
		UsedBytes:      usedBytes,      // Directory-specific usage
		AvailableBytes: availableBytes, // Filesystem available
		FreeBytes:      freeBytes,      // Filesystem free
	}, fileCount, nil
}

// calculateVolumeUsage performs the actual filesystem calculation
func (c *volumeUsageCache) calculateVolumeUsage(path string) (*VolumeUsage, error) {
	// Get filesystem stats for available/free space
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats for %s: %w", path, err)
	}

	blockSize := uint64(stat.Bsize)
	totalBytes := stat.Blocks * blockSize
	freeBytes := stat.Bfree * blockSize
	availableBytes := stat.Bavail * blockSize

	// Calculate directory-specific usage by walking the directory
	usedBytes, err := calculateDirectoryUsage(path)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate directory usage for %s: %w", path, err)
	}

	return &VolumeUsage{
		TotalBytes:     totalBytes,     // Filesystem total (for capacity info)
		UsedBytes:      usedBytes,      // Directory-specific usage
		AvailableBytes: availableBytes, // Filesystem available
		FreeBytes:      freeBytes,      // Filesystem free
	}, nil
}

// calculateDirectoryUsage calculates the total size of a directory and its contents
func calculateDirectoryUsage(path string) (uint64, error) {
	var totalSize uint64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access rather than failing completely
			return nil
		}

		if !info.IsDir() {
			totalSize += uint64(info.Size())
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return totalSize, nil
}

// calculateDirectoryUsageWithCount calculates directory size and counts files
func calculateDirectoryUsageWithCount(path string) (totalSize uint64, fileCount int, err error) {
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Skip files we can't access rather than failing completely
			return nil
		}

		if !info.IsDir() {
			totalSize += uint64(info.Size())
			fileCount++
		}

		return nil
	})

	return totalSize, fileCount, err
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
