package filesystem

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// cacheEntry holds cached directory stats with dynamic interval calculation
type cacheEntry struct {
	stats           *DirectoryStats
	dynamicInterval time.Duration
}

// adaptiveCache provides intelligent caching based on directory characteristics
type adaptiveCache struct {
	mu     sync.RWMutex
	cache  map[string]*cacheEntry
	config CacheConfig
	log    logrus.FieldLogger
}

// newAdaptiveCache creates a new adaptive cache with the given configuration
func newAdaptiveCache(config CacheConfig, log logrus.FieldLogger) *adaptiveCache {
	return &adaptiveCache{
		cache:  make(map[string]*cacheEntry),
		config: config,
		log:    log.WithField("component", "cache"),
	}
}

// get retrieves cached stats for a path if they haven't expired
func (c *adaptiveCache) get(path string) (*DirectoryStats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[path]
	if !exists {
		c.log.WithField("path", path).Debug("Cache miss - no entry found")
		return nil, false
	}

	// Use dynamic interval if available, otherwise fall back to default timeout
	cacheLimit := entry.dynamicInterval
	if cacheLimit == 0 {
		cacheLimit = c.config.DefaultTimeout
	}

	age := time.Since(entry.stats.Timestamp)
	if age >= cacheLimit {
		c.log.WithFields(logrus.Fields{
			"path":             path,
			"age":              age,
			"cache_limit":      cacheLimit,
			"dynamic_interval": entry.dynamicInterval > 0,
		}).Debug("Cache miss - entry expired")

		return nil, false
	}

	c.log.WithFields(logrus.Fields{
		"path":             path,
		"age":              age,
		"cache_limit":      cacheLimit,
		"dynamic_interval": entry.dynamicInterval > 0,
		"file_count":       entry.stats.FileCount,
		"calc_time":        entry.stats.CalculationTime,
	}).Debug("Cache hit - using cached stats")

	return entry.stats, true
}

// set stores stats in the cache with calculated optimal interval
func (c *adaptiveCache) set(path string, stats *DirectoryStats) {
	dynamicInterval := c.calculateOptimalInterval(stats)

	c.mu.Lock()
	c.cache[path] = &cacheEntry{
		stats:           stats,
		dynamicInterval: dynamicInterval,
	}
	c.mu.Unlock()

	c.log.WithFields(logrus.Fields{
		"path":             path,
		"total_bytes":      stats.TotalBytes,
		"file_count":       stats.FileCount,
		"calc_time":        stats.CalculationTime,
		"dynamic_interval": dynamicInterval,
	}).Info("Cached directory stats with dynamic interval")
}

// calculateOptimalInterval determines the best cache interval based on directory characteristics
func (c *adaptiveCache) calculateOptimalInterval(stats *DirectoryStats) time.Duration {
	// Base interval = calculation_time * 10
	baseInterval := stats.CalculationTime * 10

	// Apply size thresholds from config
	for _, threshold := range c.config.SizeThresholds {
		if stats.TotalBytes >= threshold.Bytes {
			if threshold.Interval > baseInterval {
				baseInterval = threshold.Interval
			}
		}
	}

	// Apply file count thresholds from config
	for _, threshold := range c.config.FileCountThresholds {
		if stats.FileCount >= threshold.Count {
			if threshold.Interval > baseInterval {
				baseInterval = threshold.Interval
			}
		}
	}

	// Clamp between min and max timeout
	if baseInterval < c.config.MinimumTimeout {
		baseInterval = c.config.MinimumTimeout
	}

	if baseInterval > c.config.MaximumTimeout {
		baseInterval = c.config.MaximumTimeout
	}

	return baseInterval
}

// isExpired checks if a cache entry has expired (for cleanup)
func (c *adaptiveCache) isExpired(entry *cacheEntry) bool {
	cacheLimit := entry.dynamicInterval
	if cacheLimit == 0 {
		cacheLimit = c.config.DefaultTimeout
	}

	return time.Since(entry.stats.Timestamp) >= cacheLimit
}

// clear removes all entries from the cache
func (c *adaptiveCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	c.log.Debug("Cache cleared")
}
