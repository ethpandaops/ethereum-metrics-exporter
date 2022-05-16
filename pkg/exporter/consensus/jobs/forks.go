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

// Forks reports the state of any forks (previous, active or upcoming).
type Forks struct {
	MetricExporter
	Epochs              prometheus.GaugeVec
	Activated           prometheus.GaugeVec
	Current             prometheus.GaugeVec
	client              eth2client.Service
	log                 logrus.FieldLogger
	previousCurrentFork string
}

const (
	NameFork = "fork"
)

// NewForksJob returns a new Forks instance.
func NewForksJob(client eth2client.Service, log logrus.FieldLogger, namespace string, constLabels map[string]string) Forks {
	constLabels["module"] = NameFork
	namespace = namespace + "_fork"
	return Forks{
		client:              client,
		log:                 log,
		previousCurrentFork: "",
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
	if err := f.GetCurrent(ctx); err != nil {
		f.log.WithError(err).Error("Failed to fetch current fork")
	}
}

func (f *Forks) ForkEpochs(ctx context.Context) error {
	spec, err := f.getSpec(ctx)
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

func (f *Forks) GetCurrent(ctx context.Context) error {
	// Get the current head slot.
	provider, isProvider := f.client.(eth2client.BeaconBlockHeadersProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.BeaconBlockHeadersProvider")
	}

	headSlot, err := provider.BeaconBlockHeader(ctx, "head")
	if err != nil {
		return err
	}

	spec, err := f.getSpec(ctx)
	if err != nil {
		return err
	}

	slotsPerEpoch := 32
	if v, ok := spec["SLOTS_PER_EPOCH"]; ok {
		slotsPerEpoch = cast.ToInt(v)
	}

	current := ""
	currentSlot := 0
	for k, v := range spec {
		if strings.Contains(k, "_FORK_EPOCH") {
			forkName := strings.Replace(k, "_FORK_EPOCH", "", -1)
			if int(headSlot.Header.Message.Slot)/slotsPerEpoch > cast.ToInt(v) {
				f.Activated.WithLabelValues(forkName).Set(1)
			} else {
				f.Activated.WithLabelValues(forkName).Set(0)
			}

			if currentSlot < cast.ToInt(v) {
				current = forkName
				currentSlot = cast.ToInt(v)
			}
		}
	}

	if current != f.previousCurrentFork {
		f.Current.WithLabelValues(current).Set(1)

		if f.previousCurrentFork != "" {
			f.Current.WithLabelValues(f.previousCurrentFork).Set(0)
		}

		f.previousCurrentFork = current
	}

	return nil
}

func (f *Forks) getSpec(ctx context.Context) (map[string]interface{}, error) {
	provider, isProvider := f.client.(eth2client.SpecProvider)
	if !isProvider {
		return nil, errors.New("client does not implement eth2client.SpecProvider")
	}

	return provider.Spec(ctx)
}

func (f *Forks) ObserveForkEpoch(name string, epoch uint64) {
	f.Epochs.WithLabelValues(name).Set(float64(epoch))
}
