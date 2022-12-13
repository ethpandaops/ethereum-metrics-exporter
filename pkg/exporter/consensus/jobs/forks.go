package jobs

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/beacon"
	"github.com/sirupsen/logrus"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// Forks reports the state of any forks (previous, active or upcoming).
type Forks struct {
	Epochs    prometheus.GaugeVec
	Activated prometheus.GaugeVec
	Current   prometheus.GaugeVec
	beacon    beacon.Node
	log       logrus.FieldLogger
}

const (
	NameFork = "fork"
)

// NewForksJob returns a new Forks instance.
func NewForksJob(beac beacon.Node, log logrus.FieldLogger, namespace string, constLabels map[string]string) Forks {
	constLabels["module"] = NameFork

	namespace += "_fork"

	return Forks{
		beacon: beac,
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
		Activated: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "activated",
				Help:        "The activation status of the fork (1 for activated).",
				ConstLabels: constLabels,
			},
			[]string{
				"fork",
			},
		),
		Current: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "current",
				Help:        "The current fork.",
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

func (f *Forks) Start(ctx context.Context) error {
	f.beacon.OnBlock(ctx, func(ctx context.Context, event *v1.BlockEvent) error {
		return f.calculateCurrent(ctx, event.Slot)
	})

	return nil
}

func (f *Forks) calculateCurrent(ctx context.Context, slot phase0.Slot) error {
	spec, err := f.beacon.GetSpec(ctx)
	if err != nil {
		return err
	}

	slotsPerEpoch := spec.SlotsPerEpoch

	f.Activated.Reset()
	f.Epochs.Reset()

	for _, fork := range spec.ForkEpochs {
		f.Epochs.WithLabelValues(fork.Name).Set(float64(fork.Epoch))

		if fork.Active(slot, slotsPerEpoch) {
			f.Activated.WithLabelValues(fork.Name).Set(1)
		} else {
			f.Activated.WithLabelValues(fork.Name).Set(0)
		}
	}

	current, err := spec.ForkEpochs.CurrentFork(slot, slotsPerEpoch)
	if err != nil {
		f.log.WithError(err).Error("Failed to set current fork")
	} else {
		f.Current.Reset()

		f.Current.WithLabelValues(current.Name).Set(1)
	}

	return nil
}
