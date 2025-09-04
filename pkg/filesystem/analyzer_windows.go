//go:build windows

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceExW")
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
		FilesystemFree:      fsAvailable, // On Windows, available ≈ free for unprivileged access
	}

	a.log.WithFields(logrus.Fields{
		"path":        path,
		"total_bytes": totalBytes,
		"file_count":  fileCount,
		"calc_time":   calculationTime,
	}).Info("Completed directory analysis")

	return stats, nil
}

// getFilesystemStats retrieves filesystem-level statistics (total and available space) on Windows
func (a *directoryAnalyzer) getFilesystemStats(path string) (total, available uint64, err error) {
	// Convert path to UTF-16
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert path to UTF-16: %w", err)
	}

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	// Call GetDiskFreeSpaceExW
	r1, _, e1 := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if r1 == 0 {
		return 0, 0, fmt.Errorf("GetDiskFreeSpaceExW failed: %v", e1)
	}

	return totalNumberOfBytes, freeBytesAvailable, nil
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
