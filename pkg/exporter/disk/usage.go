package disk

// Usage contains usage information for a single directory.
type Usage struct {
	Directory  string
	UsageBytes int64
	Type       string // Type of directory: "el_db", "cl_db", or "general"
}
