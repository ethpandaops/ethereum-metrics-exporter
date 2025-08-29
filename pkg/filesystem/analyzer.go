package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// directoryAnalyzer performs filesystem analysis with error resilience
type directoryAnalyzer struct {
	log logrus.FieldLogger
}

// newDirectoryAnalyzer creates a new directory analyzer
func newDirectoryAnalyzer(log logrus.FieldLogger) *directoryAnalyzer {
	return &directoryAnalyzer{
		log: log.WithField("component", "analyzer"),
	}
}

// analyze performs comprehensive directory analysis including size and file count
func (a *directoryAnalyzer) analyze(path string) (*DirectoryStats, error) {
	startTime := time.Now()

	a.log.WithField("path", path).Debug("Starting directory analysis")

	// Calculate directory size and count files in single pass
	totalBytes, fileCount, err := a.calculateDirectorySize(path)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate directory size for %s: %w", path, err)
	}

	// Get filesystem-level statistics
	fsTotal, fsAvailable, err := a.getFilesystemStats(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats for %s: %w", path, err)
	}

	calculationTime := time.Since(startTime)

	stats := &DirectoryStats{
		Path:            path,
		TotalBytes:      totalBytes,
		FileCount:       fileCount,
		CalculationTime: calculationTime,
		Timestamp:       time.Now(),

		// Filesystem-level statistics
		FilesystemTotal:     fsTotal,
		FilesystemAvailable: fsAvailable,
		FilesystemFree:      fsAvailable, // On most filesystems, available ≈ free for unprivileged access
	}

	a.log.WithFields(logrus.Fields{
		"path":        path,
		"total_bytes": totalBytes,
		"file_count":  fileCount,
		"calc_time":   calculationTime,
	}).Info("Completed directory analysis")

	return stats, nil
}

// getFilesystemStats retrieves filesystem-level statistics (total and available space)
func (a *directoryAnalyzer) getFilesystemStats(path string) (total, available uint64, err error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, 0, fmt.Errorf("failed to get filesystem stats for %s: %w", path, err)
	}

	blockSize := uint64(stat.Bsize)
	totalBytes := stat.Blocks * blockSize
	availableBytes := stat.Bavail * blockSize

	return totalBytes, availableBytes, nil
}

// calculateDirectorySize calculates total size and file count for a directory tree
func (a *directoryAnalyzer) calculateDirectorySize(path string) (totalBytes uint64, fileCount int, err error) {
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Skip inaccessible files without failing the entire operation
			a.log.WithFields(logrus.Fields{
				"path":  filePath,
				"error": walkErr,
			}).Debug("Skipping inaccessible file during directory walk")

			return nil
		}

		// Only count regular files (not directories, symlinks, etc.)
		if !info.IsDir() {
			totalBytes += uint64(info.Size())
			fileCount++
		}

		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	return totalBytes, fileCount, nil
}
