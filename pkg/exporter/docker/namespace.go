package docker

import (
	"context"
	"fmt"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

type NamespaceOperation func() error

type NamespaceManager interface {
	ExecuteInContainerNamespace(ctx context.Context, containerID string, operation NamespaceOperation) error
	GetContainerNetworkNamespace(ctx context.Context, containerID string) (netns.NsHandle, error)
}

type namespaceManager struct {
	dockerClient client.APIClient
	logger       logrus.FieldLogger
}

var _ NamespaceManager = (*namespaceManager)(nil)

func NewNamespaceManager(dockerClient client.APIClient, logger logrus.FieldLogger) NamespaceManager {
	return &namespaceManager{
		dockerClient: dockerClient,
		logger:       logger.WithField("component", "namespace_manager"),
	}
}

func (nm *namespaceManager) ExecuteInContainerNamespace(ctx context.Context, containerID string, operation NamespaceOperation) error {
	nsHandle, err := nm.GetContainerNetworkNamespace(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container network namespace: %w", err)
	}
	defer nsHandle.Close()

	return nm.withNamespaceLock(nsHandle, operation)
}

func (nm *namespaceManager) GetContainerNetworkNamespace(ctx context.Context, containerID string) (netns.NsHandle, error) {
	container, err := nm.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return netns.NsHandle(0), fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	if container.State == nil || !container.State.Running {
		return netns.NsHandle(0), fmt.Errorf("container %s is not running", containerID)
	}

	if container.State.Pid == 0 {
		return netns.NsHandle(0), fmt.Errorf("container %s has no PID", containerID)
	}

	nsHandle, err := netns.GetFromPid(container.State.Pid)
	if err != nil {
		return netns.NsHandle(0), fmt.Errorf("failed to get network namespace from PID %d: %w", container.State.Pid, err)
	}

	nm.logger.WithFields(logrus.Fields{
		"container_id": containerID,
		"pid":          container.State.Pid,
	}).Debug("Retrieved container network namespace")

	return nsHandle, nil
}

func (nm *namespaceManager) withNamespaceLock(nsHandle netns.NsHandle, operation NamespaceOperation) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	originalNS, err := netns.Get()
	if err != nil {
		return fmt.Errorf("failed to get current network namespace: %w", err)
	}
	defer originalNS.Close()

	defer func() {
		if err := netns.Set(originalNS); err != nil {
			nm.logger.WithError(err).Error("Failed to restore original network namespace")
		}
	}()

	if err := netns.Set(nsHandle); err != nil {
		return fmt.Errorf("failed to switch to container network namespace: %w", err)
	}

	nm.logger.Debug("Executing operation in container network namespace")

	if err := operation(); err != nil {
		return fmt.Errorf("operation failed in container namespace: %w", err)
	}

	nm.logger.Debug("Operation completed successfully in container network namespace")

	return nil
}
