package dockerutil

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// VolumeOwnerOptions contain the configuration for the SetVolumeOwner function.
type VolumeOwnerOptions struct {
	Log *zap.Logger

	Client *client.Client

	VolumeName string
	ImageRef   string
	TestName   string
	UidGid     string
}

// SetVolumeOwner configures the owner of a volume to match the default user in the supplied image reference.
func SetVolumeOwner(ctx context.Context, opts VolumeOwnerOptions) error {
	owner := opts.UidGid
	if owner == "" {
		owner = GetRootUserString()
	}

	// Start a one-off container to chmod and chown the volume.

	containerName := fmt.Sprintf("interchaintest-volumeowner-%d-%s", time.Now().UnixNano(), RandLowerCaseLetterString(5))

	if err := ensureBusybox(ctx, opts.Client); err != nil {
		return err
	}

	const mountPath = "/mnt/dockervolume"
	cc, err := opts.Client.ContainerCreate(
		ctx,
		&container.Config{
			Image: busyboxRef, // Using busybox image which has chown and chmod.

			Entrypoint: []string{"sh", "-c"},
			Cmd: []string{
				`chown "$2" "$1" && chmod 0700 "$1"`,
				"_", // Meaningless arg0 for sh -c with positional args.
				mountPath,
				owner,
			},

			// Root user so we have permissions to set ownership and mode.
			User: GetRootUserString(),

			Labels: map[string]string{CleanupLabel: opts.TestName},
		},
		&container.HostConfig{
			Binds:      []string{opts.VolumeName + ":" + mountPath},
			AutoRemove: true,
		},
		nil, // No networking necessary.
		nil,
		containerName,
	)
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	autoRemoved := false
	defer func() {
		if autoRemoved {
			// No need to attempt removing the container if we successfully started and waited for it to complete.
			return
		}

		if err := opts.Client.ContainerRemove(ctx, cc.ID, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			opts.Log.Warn("Failed to remove volume-owner container", zap.String("container_id", cc.ID), zap.Error(err))
		}
	}()

	if err := opts.Client.ContainerStart(ctx, cc.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("starting volume-owner container: %w", err)
	}

	waitCh, errCh := opts.Client.ContainerWait(ctx, cc.ID, container.WaitConditionNotRunning)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	case res := <-waitCh:
		autoRemoved = true

		if res.Error != nil {
			return fmt.Errorf("waiting for volume-owner container: %s", res.Error.Message)
		}

		if res.StatusCode != 0 {
			return fmt.Errorf("configuring volume exited %d", res.StatusCode)
		}
	}

	return nil
}
