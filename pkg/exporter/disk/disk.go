package disk

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type DiskUsage interface {
	GetUsage(ctx context.Context, directories []string) ([]DiskUsed, error)
}

type diskUsage struct {
	log     logrus.FieldLogger
	metrics Metrics
}

func NewDiskUsage(ctx context.Context, log logrus.FieldLogger, metrics Metrics) (DiskUsage, error) {
	return &diskUsage{
		log:     log,
		metrics: metrics,
	}, nil
}

func (d *diskUsage) GetUsage(ctx context.Context, directories []string) ([]DiskUsed, error) {
	var diskUsed []DiskUsed
	for _, directory := range directories {
		info, err := os.Lstat(directory)
		if err != nil {
			d.log.WithField("directory", directory).Warn("Directory does not exist")
			continue
		}

		used, err := getDiskUsed(directory, info)
		if err != nil {
			d.log.WithField("directory", directory).WithError(err).Error("Failed to get usage")
			continue
		}

		diskUsed = append(diskUsed, DiskUsed{
			Directory:  directory,
			UsageBytes: used,
		})
	}

	for _, disk := range diskUsed {
		d.log.WithField("directory", disk.Directory).WithField("usage", disk.UsageBytes).Info("Disk usage")
		d.metrics.ObserveDiskUsage(disk)
	}

	return diskUsed, nil
}

func getDiskUsed(currentPath string, info os.FileInfo) (int64, error) {
	size := info.Size()
	if !info.IsDir() {
		return size, nil
	}

	dir, err := os.Open(currentPath)

	if err != nil {
		return size, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return size, err
	}

	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}
		s, _ := getDiskUsed(currentPath+"/"+file.Name(), file)
		size += s
	}

	return size, nil
}
