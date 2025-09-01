package disk

// Usage contains usage information for a single directory.
type Usage struct {
	Directory string

	// Directory-specific usage
	UsageBytes int64

	// Filesystem-level statistics
	FilesystemTotal     int64 // Total filesystem capacity
	FilesystemAvailable int64 // Available space on filesystem
	FilesystemFree      int64 // Free space on filesystem
}
