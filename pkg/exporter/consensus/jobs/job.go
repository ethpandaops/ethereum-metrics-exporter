package jobs

type MetricExporter interface {
	// Name returns the name of the job.
	Name() string
	// Start starts the job.
	Start()
}
