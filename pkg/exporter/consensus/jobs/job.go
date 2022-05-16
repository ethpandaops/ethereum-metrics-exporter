package jobs

// MetricExporter defines a consensus metric job.
type MetricExporter interface {
	// Name returns the name of the job.
	Name() string
	// Start starts the job.
	Start()
}
