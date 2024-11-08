package dockerutil

import (
	"context"
	"fmt"
	"net"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

type ContainerLifecycle struct {
	log               *zap.Logger
	client            *dockerclient.Client
	containerName     string
	id                string
	preStartListeners Listeners
}

func NewContainerLifecycle(log *zap.Logger, client *dockerclient.Client, containerName string) *ContainerLifecycle {
	return &ContainerLifecycle{
		log:           log,
		client:        client,
		containerName: containerName,
	}
}

func (c *ContainerLifecycle) CreateContainer(
	ctx context.Context,
	testName string,
	networkID string,
	image ibc.DockerImage,
	ports nat.PortMap,
	volumeBinds []string,
	mounts []mount.Mount,
	hostName string,
	cmd []string,
	env []string,
) error {
	imageRef := image.Ref()
	c.log.Info(
		"Will run command",
		zap.String("image", imageRef),
		zap.String("container", c.containerName),
		zap.String("command", strings.Join(cmd, " ")),
	)

	pS := nat.PortSet{}
	for k := range ports {
		pS[k] = struct{}{}
	}

	pb, listeners, err := GeneratePortBindings(ports)
	if err != nil {
		return fmt.Errorf("failed to generate port bindings: %w", err)
	}

	c.preStartListeners = listeners

	cc, err := c.client.ContainerCreate(
		ctx,
		&container.Config{
			Image: imageRef,

			Entrypoint: []string{},
			Cmd:        cmd,
			Env:        env,

			Hostname: hostName,

			Labels: map[string]string{CleanupLabel: testName},

			ExposedPorts: pS,
		},
		&container.HostConfig{
			Binds:           volumeBinds,
			PortBindings:    pb,
			PublishAllPorts: true,
			AutoRemove:      false,
			DNS:             []string{},
			Mounts:          mounts,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkID: {},
			},
		},
		nil,
		c.containerName,
	)
	if err != nil {
		listeners.CloseAll()
		c.preStartListeners = []net.Listener{}
		return err
	}
	c.id = cc.ID
	return nil
}

func (c *ContainerLifecycle) StartContainer(ctx context.Context) error {
	// lock port allocation for the time between freeing the ports from the
	// temporary listeners to the consumption of the ports by the container
	mu.RLock()
	defer mu.RUnlock()

	c.preStartListeners.CloseAll()
	c.preStartListeners = []net.Listener{}

	if err := StartContainer(ctx, c.client, c.id); err != nil {
		return err
	}

	c.log.Info("Container started", zap.String("container", c.containerName))

	return nil
}

func (c *ContainerLifecycle) PauseContainer(ctx context.Context) error {
	return c.client.ContainerPause(ctx, c.id)
}

func (c *ContainerLifecycle) UnpauseContainer(ctx context.Context) error {
	return c.client.ContainerUnpause(ctx, c.id)
}

func (c *ContainerLifecycle) StopContainer(ctx context.Context) error {
	var timeout container.StopOptions
	timeoutSec := 30
	timeout.Timeout = &timeoutSec

	return c.client.ContainerStop(ctx, c.id, timeout)
}

func (c *ContainerLifecycle) RemoveContainer(ctx context.Context) error {
	err := c.client.ContainerRemove(ctx, c.id, dockertypes.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("remove container %s: %w", c.containerName, err)
	}
	return nil
}

func (c *ContainerLifecycle) ContainerID() string {
	return c.id
}

func (c *ContainerLifecycle) GetHostPorts(ctx context.Context, portIDs ...string) ([]string, error) {
	cjson, err := c.client.ContainerInspect(ctx, c.id)
	if err != nil {
		return nil, err
	}
	ports := make([]string, len(portIDs))
	for i, p := range portIDs {
		ports[i] = GetHostPort(cjson, p)
	}
	return ports, nil
}

// Running will inspect the container and check its state to determine if it is currently running.
// If the container is running nil will be returned, otherwise an error is returned.
func (c *ContainerLifecycle) Running(ctx context.Context) error {
	cjson, err := c.client.ContainerInspect(ctx, c.id)
	if err != nil {
		return err
	}
	if cjson.State.Running {
		return nil
	}
	return fmt.Errorf("container with name %s and id %s is not running", c.containerName, c.id)
}
