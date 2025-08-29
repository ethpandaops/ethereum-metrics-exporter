package disk

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Metrics defines the interface for reporting disk usage metrics.
type Metrics interface {
	// ObserveDiskUsage reports the disk usage for the directory.
	ObserveDiskUsage(usage Usage)
}

type metrics struct {
	log           logrus.FieldLogger
	diskUsage     *prometheus.GaugeVec
	diskSize      *prometheus.GaugeVec
	diskAvailable *prometheus.GaugeVec
	diskFree      *prometheus.GaugeVec
}

// NewMetrics returns a new Metrics instance.
func NewMetrics(log logrus.FieldLogger, namespace string) Metrics {
	constLabels := make(prometheus.Labels)

	m := &metrics{
		log: log,
		diskUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "usage_bytes",
				Help:        "How large the directory is (in bytes).",
				ConstLabels: constLabels,
			},
			[]string{
				"directory",
			},
		),
		diskSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "size_bytes",
				Help:        "Total filesystem capacity (in bytes).",
				ConstLabels: constLabels,
			},
			[]string{
				"directory",
			},
		),
		diskAvailable: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "available_bytes",
				Help:        "Available space on filesystem (in bytes).",
				ConstLabels: constLabels,
			},
			[]string{
				"directory",
			},
		),
		diskFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "free_bytes",
				Help:        "Free space on filesystem (in bytes).",
				ConstLabels: constLabels,
			},
			[]string{
				"directory",
			},
		),
	}

	prometheus.MustRegister(
		m.diskUsage,
		m.diskSize,
		m.diskAvailable,
		m.diskFree,
	)

	return m
}

func (m *metrics) ObserveDiskUsage(usage Usage) {
	m.diskUsage.WithLabelValues(usage.Directory).Set(float64(usage.UsageBytes))
	m.diskSize.WithLabelValues(usage.Directory).Set(float64(usage.FilesystemTotal))
	m.diskAvailable.WithLabelValues(usage.Directory).Set(float64(usage.FilesystemAvailable))
	m.diskFree.WithLabelValues(usage.Directory).Set(float64(usage.FilesystemFree))
}
