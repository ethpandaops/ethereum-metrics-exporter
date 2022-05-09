package jobs

import (
	"github.com/prometheus/client_golang/prometheus"
)

type ForkMetrics struct {
	Forks prometheus.GaugeVec
}

func NewForkMetrics(namespace string, constLabels map[string]string) ForkMetrics {
	namespace = namespace + "_fork"
	return ForkMetrics{
		Forks: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "epoch",
				Help:        "The fork epoch for the version.",
				ConstLabels: constLabels,
			},
			[]string{
				"fork",
			},
		),
	}
}

func (f *ForkMetrics) ObserveFork(name string, epoch uint64) {
	f.Forks.WithLabelValues(name).Set(float64(epoch))
}
