package jobs

type MetricExporter interface {
	// RequiredModules returns the list of modules that are required to run the job.
	RequiredModules() []string
	// Name returns the name of the job.
	Name() string
	// Start starts the job.
	Start()
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func ExporterCanRun(enabledModules []string, requiredModules []string) bool {
	for _, module := range requiredModules {
		if !contains(enabledModules, module) {
			return false
		}
	}

	return true
}
