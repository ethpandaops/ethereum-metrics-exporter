package docker

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type containerCollector struct {
	client *client.Client
	log    logrus.FieldLogger
}

func newCollector(dockerClient *client.Client, log logrus.FieldLogger) *containerCollector {
	return &containerCollector{
		client: dockerClient,
		log:    log,
	}
}

func (c *containerCollector) getContainerStats(ctx context.Context, containerID string) (*types.StatsJSON, error) {
	stats, err := c.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var v types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *containerCollector) findContainer(ctx context.Context, nameOrID string) (*types.Container, error) {
	containers, err := c.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for i := range containers {
		ctr := &containers[i]
		// Check by ID (both full and short)
		if strings.HasPrefix(ctr.ID, nameOrID) {
			return ctr, nil
		}

		// Check by name
		for _, name := range ctr.Names {
			// Remove leading '/' from container name
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == nameOrID {
				return ctr, nil
			}
		}
	}

	return nil, nil // Container not found
}

func (c *containerCollector) isContainerRunning(ctr *types.Container) bool {
	return ctr.State == "running"
}
