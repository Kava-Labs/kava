package dockerutil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testDockerImage = "busybox"
	testDockerTag   = "latest"
)

func TestNewImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Parallel()

	cl, networkID := DockerSetup(t)

	for _, tt := range []struct {
		Client     *client.Client
		NetworkID  string
		Repository string
		TestName   string
	}{
		{nil, networkID, "repo", t.Name()},
		{cl, "", "repo", t.Name()},
		{cl, networkID, "", t.Name()},
		{cl, networkID, "repo", ""},
	} {
		require.Panics(t, func() {
			NewImage(zap.NewNop(), tt.Client, tt.NetworkID, tt.TestName, tt.Repository, "")
		}, tt)
	}
}

func TestImage_Run(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	client, networkID := DockerSetup(t)
	image := NewImage(zap.NewNop(), client, networkID, t.Name(), testDockerImage, testDockerTag)

	t.Run("happy path", func(t *testing.T) {
		res := image.Run(ctx, []string{"echo", "-n", "hello"}, ContainerOptions{})
		stdout, stderr, err := res.Stdout, res.Stderr, res.Err

		require.NoError(t, err)
		require.Equal(t, "hello", string(stdout))
		require.Empty(t, string(stderr))
	})

	t.Run("binds", func(t *testing.T) {
		const scriptBody = `#!/bin/sh
echo -n hi from stderr >> /dev/stderr
`
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "test.sh"), []byte(scriptBody), 0777)
		require.NoError(t, err)

		opts := ContainerOptions{
			Binds: []string{tmpDir + ":/test"},
		}

		res := image.Run(ctx, []string{"/test/test.sh"}, opts)
		stdout, stderr, err := res.Stdout, res.Stderr, res.Err

		require.NoError(t, err)
		require.Empty(t, string(stdout))
		require.Equal(t, "hi from stderr", string(stderr))
	})

	t.Run("env vars", func(t *testing.T) {
		opts := ContainerOptions{Env: []string{"MY_ENV_VAR=foo"}}
		res := image.Run(ctx, []string{"printenv", "MY_ENV_VAR"}, opts)
		stdout, stderr, err := res.Stdout, res.Stderr, res.Err

		require.NoError(t, err)
		require.Equal(t, "foo", strings.TrimSpace(string(stdout)))
		require.Empty(t, string(stderr))
	})

	t.Run("context cancelled", func(t *testing.T) {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		res := image.Run(cctx, []string{"sleep", "100"}, ContainerOptions{})
		err := res.Err

		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("errors", func(t *testing.T) {
		for _, tt := range []struct {
			Args    []string
			WantErr string
		}{
			{[]string{"program-does-not-exist"}, "executable file not found"},
			{[]string{"sleep", "not-valid-arg"}, "sleep: invalid"},
		} {
			res := image.Run(ctx, tt.Args, ContainerOptions{})
			err := res.Err

			require.Error(t, err, tt)
			require.Contains(t, err.Error(), tt.WantErr, tt)
		}
	})

	t.Run("missing required args", func(t *testing.T) {
		require.PanicsWithError(t, "cmd cannot be empty", func() {
			_ = image.Run(ctx, nil, ContainerOptions{})
		})
	})
}

func TestContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	cl, networkID := DockerSetup(t)
	image := NewImage(zap.NewNop(), cl, networkID, t.Name(), testDockerImage, testDockerTag)

	t.Run("wait", func(t *testing.T) {
		c, err := image.Start(ctx, []string{"echo", "-n", "started"}, ContainerOptions{})

		require.NoError(t, err)
		require.NotEmpty(t, c.Name)
		require.NotEmpty(t, c.Hostname)

		res := c.Wait(ctx, 0)
		stdout, stderr, err := res.Stdout, res.Stderr, res.Err

		require.NoError(t, err)
		require.Equal(t, "started", string(stdout))
		require.Empty(t, stderr)

		containers, err := image.client.ContainerList(ctx, types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("name", c.Name)),
		})
		require.NoError(t, err)
		require.Empty(t, containers, "container was not removed")

		require.NoError(t, c.Stop(5*time.Second))
	})

	t.Run("stop long running container", func(t *testing.T) {
		c, err := image.Start(ctx, []string{"sleep", "100"}, ContainerOptions{})
		require.NoError(t, err)
		require.NoError(t, c.Stop(10*time.Second))
		require.NoError(t, c.Stop(10*time.Second)) // assert idempotent

		containers, err := image.client.ContainerList(ctx, types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("name", c.Name)),
		})
		require.NoError(t, err)
		require.Empty(t, containers, "container was not removed")
	})

	t.Run("start error", func(t *testing.T) {
		c, err := image.Start(ctx, []string{"sleep", "not valid arg"}, ContainerOptions{})
		require.NoError(t, err)

		res := c.Wait(ctx, 0)
		require.Error(t, res.Err)
	})

	t.Run("missing command", func(t *testing.T) {
		require.Panics(t, func() {
			_, _ = image.Start(ctx, nil, ContainerOptions{})
		})
	})
}
