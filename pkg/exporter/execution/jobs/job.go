package jobs

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]

	return ok
}

// ExporterCanRun returns true if the job can run with the enabled modules.
func ExporterCanRun(enabledModules, requiredModules []string) bool {
	for _, module := range requiredModules {
		if !contains(enabledModules, module) {
			return false
		}
	}

	return true
}
