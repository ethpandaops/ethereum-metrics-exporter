package jobs

import (
	"context"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Event reports event counts.
type Event struct {
	log   logrus.FieldLogger
	Count prometheus.CounterVec
}

const (
	NameEvent = "event"
)

// NewEvent creates a new Event instance.
func NewEventJob(client eth2client.Service, log logrus.FieldLogger, namespace string, constLabels map[string]string) Event {
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
	}
}

func (b *Event) Name() string {
	return NameEvent
}

func (b *Event) Start(ctx context.Context) {}

func (b *Event) HandleEvent(ctx context.Context, event *v1.Event) {
	b.Count.WithLabelValues(event.Topic).Inc()
}
