package disk

// DiskUsed contains usage information for a single directory.
type DiskUsed struct {
	Directory  string
	UsageBytes int64
}
