package jobs

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	eth2client "github.com/attestantio/go-eth2-client"
)

type Forks struct {
	MetricExporter
	Epochs prometheus.GaugeVec
	client eth2client.Service
	log    logrus.FieldLogger
}

const (
	NameFork = "fork"
)

func NewForksJob(client eth2client.Service, log logrus.FieldLogger, namespace string, constLabels map[string]string) Forks {
	constLabels["module"] = NameFork
	namespace = namespace + "_fork"
	return Forks{
		client: client,
		log:    log,
		Epochs: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "epoch",
				Help:        "The epoch for the fork.",
				ConstLabels: constLabels,
			},
			[]string{
				"fork",
			},
		),
	}
}

func (f *Forks) Name() string {
	return NameFork
}

func (f *Forks) Start(ctx context.Context) {
	f.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 600):
			f.tick(ctx)
		}
	}
}

func (f *Forks) tick(ctx context.Context) {
	if err := f.ForkEpochs(ctx); err != nil {
		f.log.WithError(err).Error("Failed to fetch fork epochs")
	}
}

func (f *Forks) ForkEpochs(ctx context.Context) error {
	// Extract the forks out of the spec.
	provider, isProvider := f.client.(eth2client.SpecProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.SpecProvider")
	}

	spec, err := provider.Spec(ctx)
	if err != nil {
		return err
	}

	for k, v := range spec {
		if strings.Contains(k, "_FORK_EPOCH") {
			f.ObserveForkEpoch(strings.Replace(k, "_FORK_EPOCH", "", -1), cast.ToUint64(v))
		}
	}

	return nil
}

func (f *Forks) ObserveForkEpoch(name string, epoch uint64) {
	f.Epochs.WithLabelValues(name).Set(float64(epoch))
}
