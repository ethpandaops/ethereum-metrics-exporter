package disk

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// UsageMetrics reports disk usage metrics
type UsageMetrics interface {
	// StartAsync starts the disk usage metrics collection.
	StartAsync(ctx context.Context)
	// GetUsage returns the usage of the directories.
	GetUsage(ctx context.Context, directories []string) ([]Usage, error)
}

type diskUsage struct {
	log         logrus.FieldLogger
	metrics     Metrics
	directories []string
	interval    time.Duration
}

// NewUsage returns a new DiskUsage instance.
func NewUsage(ctx context.Context, log logrus.FieldLogger, namespace string, directories []string, interval time.Duration) (UsageMetrics, error) {
	return &diskUsage{
		log:         log,
		metrics:     NewMetrics(log, namespace),
		directories: directories,
		interval:    interval,
	}, nil
}

func (d *diskUsage) StartAsync(ctx context.Context) {
	d.log.WithField("directories", d.directories).Info("Starting disk usage metrics...")

	_, err := d.GetUsage(ctx, d.directories)
	if err != nil {
		d.log.WithError(err).Error("Failed to get disk usage")
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(d.interval):
				if _, err := d.GetUsage(ctx, d.directories); err != nil {
					d.log.WithError(err).Error("Failed to get disk usage")
				}
			}
		}
	}()
}

func (d *diskUsage) GetUsage(ctx context.Context, directories []string) ([]Usage, error) {
	//nolint:prealloc // we dont know how much success we'll have
	var diskUsed []Usage

	for _, directory := range directories {
		_, err := os.Lstat(directory)
		if err != nil {
			d.log.WithField("directory", directory).Warn("Directory does not exist")

			continue
		}

		used, err := getDiskUsed(directory)
		if err != nil {
			d.log.WithField("directory", directory).WithError(err).Error("Failed to get usage")

			continue
		}

		diskUsed = append(diskUsed, Usage{
			Directory:  directory,
			UsageBytes: used,
		})
	}

	for _, disk := range diskUsed {
		d.metrics.ObserveDiskUsage(disk)
	}

	return diskUsed, nil
}

func getDiskUsed(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return err
	})

	return size, err
}
