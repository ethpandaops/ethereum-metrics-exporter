package disk

import "github.com/prometheus/client_golang/prometheus"

type Metrics interface {
	ObserveDiskUsage(disk DiskUsed)
}

type metrics struct {
	diskUsage *prometheus.GaugeVec
}

func NewMetrics(namespace string) Metrics {
	constLabels := make(prometheus.Labels)

	m := &metrics{
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
