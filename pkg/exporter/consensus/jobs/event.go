package jobs

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/consensus/beacon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Event reports event counts.
type Event struct {
	log                logrus.FieldLogger
	Count              prometheus.CounterVec
	TimeSinceLastEvent prometheus.Gauge

	beacon beacon.Node

	LastEventTime time.Time
}

const (
	NameEvent = "event"
)

// NewEvent creates a new Event instance.
func NewEventJob(client eth2client.Service, bc beacon.Node, log logrus.FieldLogger, namespace string, constLabels map[string]string) Event {
	constLabels["module"] = NameEvent
	namespace += "_event"

	return Event{
		log:    log,
		beacon: bc,
		Count: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "count",
				Help:        "The count of beacon events.",
				ConstLabels: constLabels,
			},
			[]string{
				"name",
			},
		),
		TimeSinceLastEvent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "time_since_last_subscription_event_ms",
				Help:        "The amount of time since the last subscription event (in milliseconds).",
				ConstLabels: constLabels,
			},
		),
		LastEventTime: time.Now(),
	}
}

func (e *Event) Name() string {
	return NameEvent
}

func (e *Event) Start(ctx context.Context) error {
	if _, err := e.beacon.OnEvent(ctx, e.HandleEvent); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 1):
			e.tick(ctx)
		}
	}
}

//nolint:unparam // ctx will probably be used in the future
func (e *Event) tick(ctx context.Context) {
	e.TimeSinceLastEvent.Set(float64(time.Since(e.LastEventTime).Milliseconds()))
}

func (e *Event) HandleEvent(ctx context.Context, event *v1.Event) error {
	e.Count.WithLabelValues(event.Topic).Inc()
	e.LastEventTime = time.Now()
	e.TimeSinceLastEvent.Set(0)

	return nil
}
