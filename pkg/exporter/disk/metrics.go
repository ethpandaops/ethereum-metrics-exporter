package disk

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Metrics interface {
	ObserveDiskUsage(disk DiskUsed)
}

type metrics struct {
	log       logrus.FieldLogger
	diskUsage *prometheus.GaugeVec
}

func NewMetrics(log logrus.FieldLogger, namespace string) Metrics {
	constLabels := make(prometheus.Labels)

	m := &metrics{
		log: log,
		diskUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "disk_usage",
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

func (m *metrics) ObserveDiskUsage(disk DiskUsed) {
	m.diskUsage.WithLabelValues(disk.Directory).Set(float64(disk.UsageBytes))
}
