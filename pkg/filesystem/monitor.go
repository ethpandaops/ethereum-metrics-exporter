package filesystem

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// monitor implements the Monitor interface with intelligent caching and orchestration
type monitor struct {
	config   MonitorConfig
	cache    *adaptiveCache
	analyzer *directoryAnalyzer
	metrics  *metricsCollector

	mu    sync.RWMutex
	stats map[string]*DirectoryStats

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	log logrus.FieldLogger
}

// NewMonitor creates a new filesystem monitor with the given configuration
func NewMonitor(config *MonitorConfig, log logrus.FieldLogger) Monitor {
	return &monitor{
		config:   *config,
		cache:    newAdaptiveCache(config.CacheConfig, log),
		analyzer: newDirectoryAnalyzer(log),
		metrics:  newMetricsCollector("filesystem", log),
		stats:    make(map[string]*DirectoryStats),
		log:      log.WithField("component", "filesystem-monitor"),
	}
}

// GetStats returns current stats for a specific path
func (m *monitor) GetStats(path string) (*DirectoryStats, error) {
	// Check cache first
	if stats, hit := m.cache.get(path); hit {
		m.metrics.recordCacheHit(path)
		return stats, nil
	}

	// Cache miss - analyze the path
	m.metrics.recordCacheMiss(path)

	stats, err := m.analyzer.analyze(path)
	if err != nil {
		return nil, err
	}

	// Update cache and metrics
	m.cache.set(path, stats)
	m.metrics.recordDirectoryStats(stats)

	// Update local stats map
	m.mu.Lock()
	m.stats[path] = stats
	m.mu.Unlock()

	return stats, nil
}

// GetAllStats returns stats for all monitored paths
func (m *monitor) GetAllStats() map[string]*DirectoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]*DirectoryStats, len(m.stats))
	for path, stats := range m.stats {
		result[path] = stats
	}

	return result
}

// Start begins automatic collection if interval > 0
func (m *monitor) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	// Register metrics
	m.metrics.register()

	// If no interval configured, just return (manual collection only)
	if m.config.Interval <= 0 {
		m.log.Info("Filesystem monitor started in manual mode (no automatic collection)")
		return nil
	}

	// Start collection goroutine
	m.wg.Add(1)

	go func() {
		defer m.wg.Done()

		m.collectLoop()
	}()

	m.log.WithFields(logrus.Fields{
		"interval":     m.config.Interval,
		"path_count":   len(m.config.Paths),
		"cache_config": m.config.CacheConfig.DefaultTimeout,
	}).Info("Filesystem monitor started with automatic collection")

	return nil
}

// Stop gracefully shuts down the monitor
func (m *monitor) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}

	m.wg.Wait()
	m.log.Info("Filesystem monitor stopped")

	return nil
}

// collectLoop runs the periodic collection process
func (m *monitor) collectLoop() {
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	// Initial collection
	m.collectAllPaths()

	for {
		select {
		case <-m.ctx.Done():
			m.log.Debug("Collection loop stopped due to context cancellation")
			return
		case <-ticker.C:
			m.collectAllPaths()
		}
	}
}

// collectAllPaths collects stats for all configured paths
func (m *monitor) collectAllPaths() {
	for _, path := range m.config.Paths {
		if err := m.collectPath(path); err != nil {
			m.log.WithError(err).WithField("path", path).Warn("Failed to collect path stats")
		}
	}
}

// collectPath collects stats for a single path
func (m *monitor) collectPath(path string) error {
	_, err := m.GetStats(path)
	return err
}
