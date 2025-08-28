package docker

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	// Existing metrics
	cpuUsagePercent   *prometheus.GaugeVec
	memoryUsageBytes  *prometheus.GaugeVec
	memoryLimitBytes  *prometheus.GaugeVec
	blockIOReadBytes  *prometheus.GaugeVec
	blockIOWriteBytes *prometheus.GaugeVec
	processCount      *prometheus.GaugeVec

	// New CPU metrics
	cpuKernelTime       *prometheus.GaugeVec
	cpuUserTime         *prometheus.GaugeVec
	cpuOnlineCount      *prometheus.GaugeVec
	cpuThrottlePeriods  *prometheus.GaugeVec
	cpuThrottledPeriods *prometheus.GaugeVec
	cpuThrottledTime    *prometheus.GaugeVec

	// New memory metrics
	memoryActiveAnon      *prometheus.GaugeVec
	memoryInactiveAnon    *prometheus.GaugeVec
	memoryActiveFile      *prometheus.GaugeVec
	memoryInactiveFile    *prometheus.GaugeVec
	memoryPageFaults      *prometheus.GaugeVec
	memoryMajorPageFaults *prometheus.GaugeVec
	memoryFileDirty       *prometheus.GaugeVec

	// Network metrics
	networkRxBytes   *prometheus.GaugeVec
	networkTxBytes   *prometheus.GaugeVec
	networkRxPackets *prometheus.GaugeVec
	networkTxPackets *prometheus.GaugeVec
	networkRxErrors  *prometheus.GaugeVec
	networkTxErrors  *prometheus.GaugeVec
	networkRxDropped *prometheus.GaugeVec
	networkTxDropped *prometheus.GaugeVec

	// Detailed Block I/O metrics
	blockIOReadBytesPerDevice  *prometheus.GaugeVec
	blockIOWriteBytesPerDevice *prometheus.GaugeVec

	// Process metrics
	pidsLimit *prometheus.GaugeVec
}

func newMetrics(namespace string, labelConfig LabelConfig) *metrics {
	// Build label names based on configuration
	// Always include "type" label
	labelNames := []string{"type"}

	if labelConfig.IncludeContainerName {
		labelNames = append(labelNames, "container_name")
	}

	if labelConfig.IncludeContainerID {
		labelNames = append(labelNames, "container_id")
	}

	if labelConfig.IncludeImageName {
		labelNames = append(labelNames, "image_name")
	}

	if labelConfig.IncludeImageTag {
		labelNames = append(labelNames, "image_tag")
	}

	// Ensure we have container_name label if not included
	if !labelConfig.IncludeContainerName {
		labelNames = append(labelNames, "container_name")
	}

	m := &metrics{
		cpuUsagePercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cpu_usage_percent",
				Help:      "CPU usage percentage (0-100) for Docker container",
			},
			labelNames,
		),
		memoryUsageBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_usage_bytes",
				Help:      "Current memory usage in bytes for Docker container",
			},
			labelNames,
		),
		memoryLimitBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_limit_bytes",
				Help:      "Memory limit in bytes for Docker container",
			},
			labelNames,
		),
		blockIOReadBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "block_io_read_bytes_total",
				Help:      "Total bytes read from block devices by Docker container",
			},
			labelNames,
		),
		blockIOWriteBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "block_io_write_bytes_total",
				Help:      "Total bytes written to block devices by Docker container",
			},
			labelNames,
		),
		processCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "process_count",
				Help:      "Number of processes running in Docker container",
			},
			labelNames,
		),
	}

	// New CPU metrics
	m.cpuKernelTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cpu_kernel_time_nanoseconds",
			Help:      "CPU time spent in kernel mode in nanoseconds for Docker container",
		},
		labelNames,
	)
	m.cpuUserTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cpu_user_time_nanoseconds",
			Help:      "CPU time spent in user mode in nanoseconds for Docker container",
		},
		labelNames,
	)
	m.cpuOnlineCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cpu_online_count",
			Help:      "Number of online CPUs available to Docker container",
		},
		labelNames,
	)
	m.cpuThrottlePeriods = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cpu_throttle_periods_total",
			Help:      "Total number of CPU throttling periods for Docker container",
		},
		labelNames,
	)
	m.cpuThrottledPeriods = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cpu_throttled_periods_total",
			Help:      "Total number of throttled CPU periods for Docker container",
		},
		labelNames,
	)
	m.cpuThrottledTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cpu_throttled_time_nanoseconds",
			Help:      "Total CPU throttled time in nanoseconds for Docker container",
		},
		labelNames,
	)

	// New memory metrics
	m.memoryActiveAnon = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_active_anon_bytes",
			Help:      "Active anonymous memory in bytes for Docker container",
		},
		labelNames,
	)
	m.memoryInactiveAnon = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_inactive_anon_bytes",
			Help:      "Inactive anonymous memory in bytes for Docker container",
		},
		labelNames,
	)
	m.memoryActiveFile = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_active_file_bytes",
			Help:      "Active file cache memory in bytes for Docker container",
		},
		labelNames,
	)
	m.memoryInactiveFile = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_inactive_file_bytes",
			Help:      "Inactive file cache memory in bytes for Docker container",
		},
		labelNames,
	)
	m.memoryPageFaults = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_page_faults_total",
			Help:      "Total page faults for Docker container",
		},
		labelNames,
	)
	m.memoryMajorPageFaults = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_major_page_faults_total",
			Help:      "Total major page faults for Docker container",
		},
		labelNames,
	)
	m.memoryFileDirty = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_file_dirty_bytes",
			Help:      "Dirty file pages in bytes for Docker container",
		},
		labelNames,
	)

	// Network metrics
	m.networkRxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_rx_bytes_total",
			Help:      "Total network bytes received by Docker container",
		},
		append(labelNames, "interface"),
	)
	m.networkTxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_tx_bytes_total",
			Help:      "Total network bytes transmitted by Docker container",
		},
		append(labelNames, "interface"),
	)
	m.networkRxPackets = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_rx_packets_total",
			Help:      "Total network packets received by Docker container",
		},
		append(labelNames, "interface"),
	)
	m.networkTxPackets = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_tx_packets_total",
			Help:      "Total network packets transmitted by Docker container",
		},
		append(labelNames, "interface"),
	)
	m.networkRxErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_rx_errors_total",
			Help:      "Total network receive errors for Docker container",
		},
		append(labelNames, "interface"),
	)
	m.networkTxErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_tx_errors_total",
			Help:      "Total network transmit errors for Docker container",
		},
		append(labelNames, "interface"),
	)
	m.networkRxDropped = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_rx_dropped_total",
			Help:      "Total network packets dropped on receive for Docker container",
		},
		append(labelNames, "interface"),
	)
	m.networkTxDropped = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "network_tx_dropped_total",
			Help:      "Total network packets dropped on transmit for Docker container",
		},
		append(labelNames, "interface"),
	)

	// Detailed Block I/O metrics
	m.blockIOReadBytesPerDevice = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_io_read_bytes_per_device",
			Help:      "Bytes read from specific block devices by Docker container",
		},
		append(labelNames, "device_major", "device_minor"),
	)
	m.blockIOWriteBytesPerDevice = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_io_write_bytes_per_device",
			Help:      "Bytes written to specific block devices by Docker container",
		},
		append(labelNames, "device_major", "device_minor"),
	)

	// Process metrics
	m.pidsLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "pids_limit",
			Help:      "Maximum number of processes allowed in Docker container",
		},
		labelNames,
	)

	// Register all metrics
	prometheus.MustRegister(m.cpuUsagePercent)
	prometheus.MustRegister(m.memoryUsageBytes)
	prometheus.MustRegister(m.memoryLimitBytes)
	prometheus.MustRegister(m.blockIOReadBytes)
	prometheus.MustRegister(m.blockIOWriteBytes)
	prometheus.MustRegister(m.processCount)

	// Register new metrics
	prometheus.MustRegister(m.cpuKernelTime)
	prometheus.MustRegister(m.cpuUserTime)
	prometheus.MustRegister(m.cpuOnlineCount)
	prometheus.MustRegister(m.cpuThrottlePeriods)
	prometheus.MustRegister(m.cpuThrottledPeriods)
	prometheus.MustRegister(m.cpuThrottledTime)
	prometheus.MustRegister(m.memoryActiveAnon)
	prometheus.MustRegister(m.memoryInactiveAnon)
	prometheus.MustRegister(m.memoryActiveFile)
	prometheus.MustRegister(m.memoryInactiveFile)
	prometheus.MustRegister(m.memoryPageFaults)
	prometheus.MustRegister(m.memoryMajorPageFaults)
	prometheus.MustRegister(m.memoryFileDirty)
	prometheus.MustRegister(m.networkRxBytes)
	prometheus.MustRegister(m.networkTxBytes)
	prometheus.MustRegister(m.networkRxPackets)
	prometheus.MustRegister(m.networkTxPackets)
	prometheus.MustRegister(m.networkRxErrors)
	prometheus.MustRegister(m.networkTxErrors)
	prometheus.MustRegister(m.networkRxDropped)
	prometheus.MustRegister(m.networkTxDropped)
	prometheus.MustRegister(m.blockIOReadBytesPerDevice)
	prometheus.MustRegister(m.blockIOWriteBytesPerDevice)
	prometheus.MustRegister(m.pidsLimit)

	return m
}

func (m *metrics) updateContainerMetrics(containerName string, stats *types.StatsJSON, labels prometheus.Labels) {
	// Calculate CPU usage percentage
	cpuPercent := calculateCPUPercent(stats)
	m.cpuUsagePercent.With(labels).Set(cpuPercent)

	// New CPU metrics
	m.cpuKernelTime.With(labels).Set(float64(stats.CPUStats.CPUUsage.UsageInKernelmode))
	m.cpuUserTime.With(labels).Set(float64(stats.CPUStats.CPUUsage.UsageInUsermode))
	m.cpuOnlineCount.With(labels).Set(float64(stats.CPUStats.OnlineCPUs))
	m.cpuThrottlePeriods.With(labels).Set(float64(stats.CPUStats.ThrottlingData.Periods))
	m.cpuThrottledPeriods.With(labels).Set(float64(stats.CPUStats.ThrottlingData.ThrottledPeriods))
	m.cpuThrottledTime.With(labels).Set(float64(stats.CPUStats.ThrottlingData.ThrottledTime))

	// Memory metrics
	m.memoryUsageBytes.With(labels).Set(float64(stats.MemoryStats.Usage))
	m.memoryLimitBytes.With(labels).Set(float64(stats.MemoryStats.Limit))

	// New detailed memory metrics
	if stats.MemoryStats.Stats != nil {
		if val, ok := stats.MemoryStats.Stats["active_anon"]; ok {
			m.memoryActiveAnon.With(labels).Set(float64(val))
		}

		if val, ok := stats.MemoryStats.Stats["inactive_anon"]; ok {
			m.memoryInactiveAnon.With(labels).Set(float64(val))
		}

		if val, ok := stats.MemoryStats.Stats["active_file"]; ok {
			m.memoryActiveFile.With(labels).Set(float64(val))
		}

		if val, ok := stats.MemoryStats.Stats["inactive_file"]; ok {
			m.memoryInactiveFile.With(labels).Set(float64(val))
		}

		if val, ok := stats.MemoryStats.Stats["pgfault"]; ok {
			m.memoryPageFaults.With(labels).Set(float64(val))
		}

		if val, ok := stats.MemoryStats.Stats["pgmajfault"]; ok {
			m.memoryMajorPageFaults.With(labels).Set(float64(val))
		}

		if val, ok := stats.MemoryStats.Stats["file_dirty"]; ok {
			m.memoryFileDirty.With(labels).Set(float64(val))
		}
	}

	// Network metrics
	for ifaceName, netStats := range stats.Networks {
		// Create extended labels with interface name
		networkLabels := make(prometheus.Labels, len(labels)+1)
		for k, v := range labels {
			networkLabels[k] = v
		}

		networkLabels["interface"] = ifaceName

		m.networkRxBytes.With(networkLabels).Set(float64(netStats.RxBytes))
		m.networkTxBytes.With(networkLabels).Set(float64(netStats.TxBytes))
		m.networkRxPackets.With(networkLabels).Set(float64(netStats.RxPackets))
		m.networkTxPackets.With(networkLabels).Set(float64(netStats.TxPackets))
		m.networkRxErrors.With(networkLabels).Set(float64(netStats.RxErrors))
		m.networkTxErrors.With(networkLabels).Set(float64(netStats.TxErrors))
		m.networkRxDropped.With(networkLabels).Set(float64(netStats.RxDropped))
		m.networkTxDropped.With(networkLabels).Set(float64(netStats.TxDropped))
	}

	// Block I/O metrics - calculate total bytes read/written and per-device stats
	var readBytes, writeBytes uint64

	for _, bioEntry := range stats.BlkioStats.IoServiceBytesRecursive {
		// Create extended labels with device information
		deviceLabels := make(prometheus.Labels, len(labels)+2)
		for k, v := range labels {
			deviceLabels[k] = v
		}

		deviceLabels["device_major"] = fmt.Sprintf("%d", bioEntry.Major)
		deviceLabels["device_minor"] = fmt.Sprintf("%d", bioEntry.Minor)

		switch bioEntry.Op {
		case "read":
			readBytes += bioEntry.Value
			m.blockIOReadBytesPerDevice.With(deviceLabels).Set(float64(bioEntry.Value))
		case "write":
			writeBytes += bioEntry.Value
			m.blockIOWriteBytesPerDevice.With(deviceLabels).Set(float64(bioEntry.Value))
		}
	}

	// Set block I/O metrics as gauges with total values
	m.blockIOReadBytes.With(labels).Set(float64(readBytes))
	m.blockIOWriteBytes.With(labels).Set(float64(writeBytes))

	// Process count and limit
	m.processCount.With(labels).Set(float64(stats.PidsStats.Current))
	m.pidsLimit.With(labels).Set(float64(stats.PidsStats.Limit))
}

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	// Calculate CPU usage percentage
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		// Use OnlineCPUs if PercpuUsage is not available (which is common)
		cpuCount := float64(stats.CPUStats.OnlineCPUs)
		if cpuCount == 0 {
			// Fallback to PercpuUsage length if OnlineCPUs is not set
			cpuCount = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
			if cpuCount == 0 {
				// If neither is available, assume 1 CPU
				cpuCount = 1
			}
		}

		return (cpuDelta / systemDelta) * cpuCount * 100.0
	}

	return 0.0
}
