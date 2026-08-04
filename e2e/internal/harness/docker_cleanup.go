package harness

import (
	"context"
	"errors"
	"fmt"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	volumetypes "github.com/docker/docker/api/types/volume"
	dockerclient "github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8/dockerutil"
)

const (
	dockerCleanupTimeout   = 45 * time.Second
	dockerOperationTimeout = 10 * time.Second
)

type dockerResourceClient interface {
	ContainerList(context.Context, dockertypes.ContainerListOptions) ([]dockertypes.Container, error)
	ContainerStop(context.Context, string, container.StopOptions) error
	ContainerRemove(context.Context, string, dockertypes.ContainerRemoveOptions) error
	VolumeList(context.Context, volumetypes.ListOptions) (volumetypes.ListResponse, error)
	VolumeRemove(context.Context, string, bool) error
	NetworkList(context.Context, dockertypes.NetworkListOptions) ([]dockertypes.NetworkResource, error)
	NetworkRemove(context.Context, string) error
}

func cleanupDockerResources(client *dockerclient.Client, runID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dockerCleanupTimeout)
	defer cancel()

	cleanupErr := cleanupDockerResourcesWithContext(ctx, client, runID)
	closeErr := client.Close()
	return errors.Join(cleanupErr, closeErr)
}

func cleanupDockerResourcesWithContext(ctx context.Context, client dockerResourceClient, runID string) error {
	label := filters.NewArgs(filters.Arg("label", dockerutil.CleanupLabel+"="+runID))
	var cleanupErrors []error

	containers, err := boundedDockerResult(ctx, func(operationCtx context.Context) ([]dockertypes.Container, error) {
		return client.ContainerList(operationCtx, dockertypes.ContainerListOptions{All: true, Filters: label})
	})
	if err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("list containers: %w", err))
	} else {
		for _, item := range containers {
			stopSeconds := 5
			if err := boundedDockerOperation(ctx, func(operationCtx context.Context) error {
				return client.ContainerStop(operationCtx, item.ID, container.StopOptions{Timeout: &stopSeconds})
			}); dockerutil.IsLoggableStopError(err) {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("stop container %s: %w", item.ID, err))
			}
			if err := boundedDockerOperation(ctx, func(operationCtx context.Context) error {
				return client.ContainerRemove(operationCtx, item.ID, dockertypes.ContainerRemoveOptions{Force: true})
			}); err != nil {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("remove container %s: %w", item.ID, err))
			}
		}
	}

	volumes, err := boundedDockerResult(ctx, func(operationCtx context.Context) (volumetypes.ListResponse, error) {
		return client.VolumeList(operationCtx, volumetypes.ListOptions{Filters: label})
	})
	if err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("list volumes: %w", err))
	} else {
		for _, volume := range volumes.Volumes {
			if err := boundedDockerOperation(ctx, func(operationCtx context.Context) error {
				return client.VolumeRemove(operationCtx, volume.Name, true)
			}); err != nil {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("remove volume %s: %w", volume.Name, err))
			}
		}
	}

	networks, err := boundedDockerResult(ctx, func(operationCtx context.Context) ([]dockertypes.NetworkResource, error) {
		return client.NetworkList(operationCtx, dockertypes.NetworkListOptions{Filters: label})
	})
	if err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("list networks: %w", err))
	} else {
		for _, network := range networks {
			if err := boundedDockerOperation(ctx, func(operationCtx context.Context) error {
				return client.NetworkRemove(operationCtx, network.ID)
			}); err != nil {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("remove network %s: %w", network.ID, err))
			}
		}
	}

	remainingContainers, err := boundedDockerResult(ctx, func(operationCtx context.Context) ([]dockertypes.Container, error) {
		return client.ContainerList(operationCtx, dockertypes.ContainerListOptions{All: true, Filters: label})
	})
	if err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("verify containers: %w", err))
	} else if len(remainingContainers) != 0 {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("verify containers: %d labeled resources remain", len(remainingContainers)))
	}
	remainingVolumes, err := boundedDockerResult(ctx, func(operationCtx context.Context) (volumetypes.ListResponse, error) {
		return client.VolumeList(operationCtx, volumetypes.ListOptions{Filters: label})
	})
	if err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("verify volumes: %w", err))
	} else if len(remainingVolumes.Volumes) != 0 {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("verify volumes: %d labeled resources remain", len(remainingVolumes.Volumes)))
	}
	remainingNetworks, err := boundedDockerResult(ctx, func(operationCtx context.Context) ([]dockertypes.NetworkResource, error) {
		return client.NetworkList(operationCtx, dockertypes.NetworkListOptions{Filters: label})
	})
	if err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("verify networks: %w", err))
	} else if len(remainingNetworks) != 0 {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("verify networks: %d labeled resources remain", len(remainingNetworks)))
	}

	if ctx.Err() != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("cleanup deadline: %w", ctx.Err()))
	}
	return errors.Join(cleanupErrors...)
}

func boundedDockerResult[T any](
	ctx context.Context,
	operation func(context.Context) (T, error),
) (T, error) {
	operationCtx, cancel := context.WithTimeout(ctx, dockerOperationTimeout)
	defer cancel()
	result, err := operation(operationCtx)
	if err != nil {
		return result, err
	}
	if err := operationCtx.Err(); err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

func boundedDockerOperation(ctx context.Context, operation func(context.Context) error) error {
	operationCtx, cancel := context.WithTimeout(ctx, dockerOperationTimeout)
	defer cancel()
	if err := operation(operationCtx); err != nil {
		return err
	}
	return operationCtx.Err()
}
