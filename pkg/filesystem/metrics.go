package filesystem

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// metricsCollector collects and exposes Prometheus metrics for filesystem operations
type metricsCollector struct {
	namespace     string
	directorySize *prometheus.GaugeVec
	fileCount     *prometheus.GaugeVec
	scanDuration  *prometheus.HistogramVec
	cacheHits     *prometheus.CounterVec
	cacheMisses   *prometheus.CounterVec
	log           logrus.FieldLogger
}

// newMetricsCollector creates a new metrics collector with the given namespace
func newMetricsCollector(namespace string, log logrus.FieldLogger) *metricsCollector {
	return &metricsCollector{
		namespace: namespace,
		directorySize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "directory_size_bytes",
				Help:      "Size of directory in bytes",
			},
			[]string{"path"},
		),
		fileCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "directory_file_count",
				Help:      "Number of files in directory",
			},
			[]string{"path"},
		),
		scanDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "scan_duration_seconds",
				Help:      "Time spent scanning directory",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~32s
			},
			[]string{"path"},
		),
		cacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Number of cache hits",
			},
			[]string{"path"},
		),
		cacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Number of cache misses",
			},
			[]string{"path"},
		),
		log: log.WithField("component", "metrics"),
	}
}

// recordDirectoryStats records directory statistics metrics
func (m *metricsCollector) recordDirectoryStats(stats *DirectoryStats) {
	m.directorySize.WithLabelValues(stats.Path).Set(float64(stats.TotalBytes))
	m.fileCount.WithLabelValues(stats.Path).Set(float64(stats.FileCount))
	m.scanDuration.WithLabelValues(stats.Path).Observe(stats.CalculationTime.Seconds())

	m.log.WithFields(logrus.Fields{
		"path":        stats.Path,
		"total_bytes": stats.TotalBytes,
		"file_count":  stats.FileCount,
		"calc_time":   stats.CalculationTime,
	}).Debug("Recorded directory stats metrics")
}

// recordCacheHit records a cache hit for the given path
func (m *metricsCollector) recordCacheHit(path string) {
	m.cacheHits.WithLabelValues(path).Inc()
	m.log.WithField("path", path).Debug("Recorded cache hit")
}

// recordCacheMiss records a cache miss for the given path
func (m *metricsCollector) recordCacheMiss(path string) {
	m.cacheMisses.WithLabelValues(path).Inc()
	m.log.WithField("path", path).Debug("Recorded cache miss")
}

// register registers all metrics with Prometheus
func (m *metricsCollector) register() {
	prometheus.MustRegister(
		m.directorySize,
		m.fileCount,
		m.scanDuration,
		m.cacheHits,
		m.cacheMisses,
	)

	m.log.WithField("namespace", m.namespace).Info("Registered filesystem metrics with Prometheus")
}
