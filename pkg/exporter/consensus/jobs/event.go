package jobs

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/consensus/api"
	"github.com/sirupsen/logrus"
)

// Event reports event counts.
type Event struct {
	log                logrus.FieldLogger
	Count              prometheus.CounterVec
	TimeSinceLastEvent prometheus.Gauge

	LastEventTime time.Time
}

const (
	NameEvent = "event"
)

// NewEvent creates a new Event instance.
func NewEventJob(client eth2client.Service, ap api.ConsensusClient, log logrus.FieldLogger, namespace string, constLabels map[string]string) Event {
	constLabels["module"] = NameEvent
	namespace += "_event"

	return Event{
		log: log,
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

func (b *Event) Name() string {
	return NameEvent
}

func (b *Event) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 1):
			b.tick(ctx)
		}
	}
}

func (b *Event) tick(ctx context.Context) {
	b.TimeSinceLastEvent.Set(float64(time.Since(b.LastEventTime).Milliseconds()))
}

func (b *Event) HandleEvent(ctx context.Context, event *v1.Event) {
	b.Count.WithLabelValues(event.Topic).Inc()
	b.LastEventTime = time.Now()
	b.TimeSinceLastEvent.Set(0)
}
