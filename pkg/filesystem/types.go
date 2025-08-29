package filesystem

import (
	"context"
	"time"
)

// DirectoryStats represents comprehensive statistics about a directory
type DirectoryStats struct {
	Path            string        // Absolute path to directory
	TotalBytes      uint64        // Total directory size in bytes
	FileCount       int           // Number of files in directory tree
	CalculationTime time.Duration // Time spent calculating these stats
	Timestamp       time.Time     // When these stats were calculated

	// Filesystem-level statistics
	FilesystemTotal     uint64 // Total filesystem capacity
	FilesystemAvailable uint64 // Available space on filesystem
	FilesystemFree      uint64 // Free space on filesystem
}

// CacheConfig defines caching behavior for filesystem monitoring
type CacheConfig struct {
	DefaultTimeout      time.Duration    // Default cache timeout
	MinimumTimeout      time.Duration    // Minimum allowed cache timeout
	MaximumTimeout      time.Duration    // Maximum allowed cache timeout
	SizeThresholds      []SizeThreshold  // Cache intervals based on directory size
	FileCountThresholds []CountThreshold // Cache intervals based on file count
}

// SizeThreshold defines cache interval based on directory size
type SizeThreshold struct {
	Bytes    uint64        // Size threshold in bytes
	Interval time.Duration // Cache interval for directories >= this size
}

// CountThreshold defines cache interval based on file count
type CountThreshold struct {
	Count    int           // File count threshold
	Interval time.Duration // Cache interval for directories >= this count
}

// MonitorConfig configures filesystem monitoring behavior
type MonitorConfig struct {
	Paths       []string      // Paths to monitor
	CacheConfig CacheConfig   // Caching configuration
	Interval    time.Duration // Collection interval (0 = no automatic collection)
}

// Monitor provides filesystem monitoring capabilities with intelligent caching
type Monitor interface {
	// GetStats returns current stats for a specific path
	GetStats(path string) (*DirectoryStats, error)

	// GetAllStats returns stats for all monitored paths
	GetAllStats() map[string]*DirectoryStats

	// Start begins automatic collection if interval > 0
	Start(ctx context.Context) error

	// Stop gracefully shuts down the monitor
	Stop() error
}

// DefaultCacheConfig returns a sensible default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		DefaultTimeout: 5 * time.Minute,
		MinimumTimeout: 1 * time.Minute,
		MaximumTimeout: time.Hour,
		SizeThresholds: []SizeThreshold{
			{Bytes: 100 * 1024 * 1024 * 1024, Interval: 30 * time.Minute}, // >100GB: 30min
			{Bytes: 10 * 1024 * 1024 * 1024, Interval: 15 * time.Minute},  // >10GB: 15min
			{Bytes: 1024 * 1024 * 1024, Interval: 5 * time.Minute},        // >1GB: 5min
			{Bytes: 100 * 1024 * 1024, Interval: 2 * time.Minute},         // >100MB: 2min
		},
		FileCountThresholds: []CountThreshold{
			{Count: 100000, Interval: 20 * time.Minute}, // >100k files: 20min
			{Count: 10000, Interval: 10 * time.Minute},  // >10k files: 10min
			{Count: 1000, Interval: 3 * time.Minute},    // >1k files: 3min
		},
	}
}
