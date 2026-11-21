package docker

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestResolveCollectionInterval(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  time.Duration
	}{
		{name: "zero falls back to default", input: 0, want: defaultCollectionInterval},
		{name: "negative falls back to default", input: -5 * time.Second, want: defaultCollectionInterval},
		{name: "positive value is preserved", input: 30 * time.Second, want: 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveCollectionInterval(tt.input)
			if got != tt.want {
				t.Fatalf("resolveCollectionInterval(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestStartAsync_ResolvedInterval_DoesNotPanic confirms that a containerMetrics built with an
// already-resolved interval (as NewContainerMetrics now guarantees) starts its ticker cleanly.
func TestStartAsync_ResolvedInterval_DoesNotPanic(t *testing.T) {
	log := logrus.New()

	cm := &containerMetrics{
		log:        log,
		containers: nil,
		interval:   resolveCollectionInterval(0),
		metrics:    newMetrics("docker_test_startasync", LabelConfig{}),
		collector:  newCollector(nil, log),
		backoff:    newBackoffState(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("StartAsync panicked with a resolved interval: %v", r)
		}
	}()

	cm.StartAsync(ctx)
}
