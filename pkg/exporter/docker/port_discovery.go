package docker

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

const (
	tcpProtocol    = "tcp"
	udpProtocol    = "udp"
	tcpListenState = "0A"
)

type PortInfo struct {
	Port     int
	Protocol string
	Type     string // "published", "exposed", "listening"
}

type PortDiscovery interface {
	DiscoverContainerPorts(ctx context.Context, containerID string) ([]PortInfo, error)
}

type portDiscovery struct {
	dockerClient client.APIClient
	logger       logrus.FieldLogger
}

var _ PortDiscovery = (*portDiscovery)(nil)

func NewPortDiscovery(dockerClient client.APIClient, logger logrus.FieldLogger) PortDiscovery {
	return &portDiscovery{
		dockerClient: dockerClient,
		logger:       logger.WithField("component", "port_discovery"),
	}
}

func (pd *portDiscovery) DiscoverContainerPorts(ctx context.Context, containerID string) ([]PortInfo, error) {
	container, err := pd.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	var ports []PortInfo

	portMap := make(map[string]bool, 50)

	publishedPorts := pd.getPublishedPorts(container)
	for _, port := range publishedPorts {
		key := fmt.Sprintf("%d:%s", port.Port, port.Protocol)
		if !portMap[key] {
			ports = append(ports, port)
			portMap[key] = true
		}
	}

	exposedPorts := pd.getExposedPorts(container)
	for _, port := range exposedPorts {
		key := fmt.Sprintf("%d:%s", port.Port, port.Protocol)
		if !portMap[key] {
			ports = append(ports, port)
			portMap[key] = true
		}
	}

	listeningPorts, err := pd.getListeningPorts(ctx, containerID)
	if err != nil {
		pd.logger.WithError(err).WithField("container_id", containerID).
			Warn("Failed to discover listening ports, using published/exposed ports only")
	} else {
		for _, port := range listeningPorts {
			key := fmt.Sprintf("%d:%s", port.Port, port.Protocol)
			if !portMap[key] {
				ports = append(ports, port)
				portMap[key] = true
			}
		}
	}

	// Log detailed port discovery results
	portDetails := make([]map[string]interface{}, 0, len(ports))
	for _, port := range ports {
		portDetails = append(portDetails, map[string]interface{}{
			"port":     port.Port,
			"protocol": port.Protocol,
			"type":     port.Type,
		})
	}

	pd.logger.WithFields(logrus.Fields{
		"container_id": containerID,
		"port_count":   len(ports),
		"ports":        portDetails,
	}).Info("Discovered container ports")

	return ports, nil
}

func (pd *portDiscovery) getPublishedPorts(container types.ContainerJSON) []PortInfo {
	ports := make([]PortInfo, 0, len(container.NetworkSettings.Ports))

	for portSpec, bindings := range container.NetworkSettings.Ports {
		if len(bindings) == 0 {
			continue
		}

		port, protocol, err := parsePortSpec(string(portSpec))
		if err != nil {
			pd.logger.WithError(err).WithField("port_spec", portSpec).
				Warn("Failed to parse published port specification")

			continue
		}

		ports = append(ports, PortInfo{
			Port:     port,
			Protocol: protocol,
			Type:     "published",
		})
	}

	return ports
}

func (pd *portDiscovery) getExposedPorts(container types.ContainerJSON) []PortInfo {
	ports := make([]PortInfo, 0, len(container.Config.ExposedPorts))

	for portSpec := range container.Config.ExposedPorts {
		port, protocol, err := parsePortSpec(string(portSpec))
		if err != nil {
			pd.logger.WithError(err).WithField("port_spec", portSpec).
				Warn("Failed to parse exposed port specification")

			continue
		}

		ports = append(ports, PortInfo{
			Port:     port,
			Protocol: protocol,
			Type:     "exposed",
		})
	}

	return ports
}

func (pd *portDiscovery) getListeningPorts(ctx context.Context, containerID string) ([]PortInfo, error) {
	nsHandle, err := netns.GetFromDocker(containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network namespace for container %s: %w", containerID, err)
	}
	defer nsHandle.Close()

	return pd.getListeningPortsInNamespace(nsHandle)
}

func (pd *portDiscovery) getListeningPortsInNamespace(nsHandle netns.NsHandle) ([]PortInfo, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	originalNS, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get original network namespace: %w", err)
	}
	defer originalNS.Close()

	if err := netns.Set(nsHandle); err != nil {
		return nil, fmt.Errorf("failed to set network namespace: %w", err)
	}

	defer func() {
		if err := netns.Set(originalNS); err != nil {
			pd.logger.WithError(err).Error("Failed to restore original network namespace")
		}
	}()

	var ports []PortInfo

	tcpPorts, err := pd.parseNetstatFile("/proc/net/tcp", "tcp")
	if err != nil {
		pd.logger.WithError(err).Warn("Failed to parse TCP listening ports")
	} else {
		ports = append(ports, tcpPorts...)
	}

	tcp6Ports, err := pd.parseNetstatFile("/proc/net/tcp6", "tcp")
	if err != nil {
		pd.logger.WithError(err).Warn("Failed to parse TCP6 listening ports")
	} else {
		ports = append(ports, tcp6Ports...)
	}

	udpPorts, err := pd.parseNetstatFile("/proc/net/udp", "udp")
	if err != nil {
		pd.logger.WithError(err).Warn("Failed to parse UDP listening ports")
	} else {
		ports = append(ports, udpPorts...)
	}

	udp6Ports, err := pd.parseNetstatFile("/proc/net/udp6", "udp")
	if err != nil {
		pd.logger.WithError(err).Warn("Failed to parse UDP6 listening ports")
	} else {
		ports = append(ports, udp6Ports...)
	}

	return ports, nil
}

func (pd *portDiscovery) parseNetstatFile(filename, protocol string) ([]PortInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer file.Close()

	var ports []PortInfo

	scanner := bufio.NewScanner(file)

	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 4 {
			continue
		}

		localAddr := fields[1]
		state := fields[3]

		if protocol == tcpProtocol && state != tcpListenState {
			continue
		}

		port, err := pd.extractPortFromAddress(localAddr)
		if err != nil {
			continue
		}

		if port == 0 {
			continue
		}

		ports = append(ports, PortInfo{
			Port:     port,
			Protocol: protocol,
			Type:     "listening",
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}

	return ports, nil
}

func (pd *portDiscovery) extractPortFromAddress(addr string) (int, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid address format: %s", addr)
	}

	portHex := parts[1]

	port, err := strconv.ParseInt(portHex, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse port %s: %w", portHex, err)
	}

	return int(port), nil
}

func parsePortSpec(portSpec string) (port int, protocol string, err error) {
	parts := strings.Split(portSpec, "/")
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid port specification: %s", portSpec)
	}

	portStr := parts[0]
	protocol = strings.ToLower(parts[1])

	port, err = strconv.Atoi(portStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid port number %s: %w", portStr, err)
	}

	if port <= 0 || port > 65535 {
		return 0, "", fmt.Errorf("port number out of range: %d", port)
	}

	if protocol != "tcp" && protocol != "udp" {
		return 0, "", fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return port, protocol, nil
}
