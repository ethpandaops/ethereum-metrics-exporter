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
	log       logrus.FieldLogger
	diskUsage *prometheus.GaugeVec
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
	}

	prometheus.MustRegister(m.diskUsage)

	return m
}

func (m *metrics) ObserveDiskUsage(usage Usage) {
	m.diskUsage.WithLabelValues(usage.Directory).Set(float64(usage.UsageBytes))
}
