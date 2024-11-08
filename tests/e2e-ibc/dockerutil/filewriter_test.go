package dockerutil_test

import (
	"context"
	"testing"

	volumetypes "github.com/docker/docker/api/types/volume"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/internal/dockerutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestFileWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping due to short mode")
	}

	t.Parallel()

	cli, network := interchaintest.DockerSetup(t)

	ctx := context.Background()
	v, err := cli.VolumeCreate(ctx, volumetypes.CreateOptions{
		Labels: map[string]string{dockerutil.CleanupLabel: t.Name()},
	})
	require.NoError(t, err)

	img := dockerutil.NewImage(
		zaptest.NewLogger(t),
		cli,
		network,
		t.Name(),
		"busybox", "stable",
	)

	fw := dockerutil.NewFileWriter(zaptest.NewLogger(t), cli, t.Name())

	t.Run("top-level file", func(t *testing.T) {
		require.NoError(t, fw.WriteFile(context.Background(), v.Name, "hello.txt", []byte("hello world")))
		res := img.Run(
			ctx,
			[]string{"sh", "-c", "cat /mnt/test/hello.txt"},
			dockerutil.ContainerOptions{
				Binds: []string{v.Name + ":/mnt/test"},
				User:  dockerutil.GetRootUserString(),
			},
		)
		require.NoError(t, res.Err)

		require.Equal(t, string(res.Stdout), "hello world")
	})

	t.Run("create nested file", func(t *testing.T) {
		require.NoError(t, fw.WriteFile(context.Background(), v.Name, "a/b/c/d.txt", []byte(":D")))
		res := img.Run(
			ctx,
			[]string{"sh", "-c", "cat /mnt/test/a/b/c/d.txt"},
			dockerutil.ContainerOptions{
				Binds: []string{v.Name + ":/mnt/test"},
				User:  dockerutil.GetRootUserString(),
			},
		)
		require.NoError(t, err)

		require.Equal(t, string(res.Stdout), ":D")
	})
}
