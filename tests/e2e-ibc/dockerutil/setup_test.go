package dockerutil_test

import (
	"context"
	"fmt"
	"testing"

	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/errdefs"
	"github.com/strangelove-ventures/interchaintest/v8/internal/dockerutil"
	"github.com/strangelove-ventures/interchaintest/v8/internal/mocktesting"
	"github.com/stretchr/testify/require"
)

func TestDockerSetup_KeepVolumes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping due to short mode")
	}

	cli, _ := dockerutil.DockerSetup(t)

	origKeep := dockerutil.KeepVolumesOnFailure
	defer func() {
		dockerutil.KeepVolumesOnFailure = origKeep
	}()

	ctx := context.Background()

	for _, tc := range []struct {
		keep       bool
		passed     bool
		volumeKept bool
	}{
		{keep: false, passed: false, volumeKept: false},
		{keep: true, passed: false, volumeKept: true},
		{keep: false, passed: true, volumeKept: false},
		{keep: true, passed: true, volumeKept: false},
	} {
		tc := tc
		state := "failed"
		if tc.passed {
			state = "passed"
		}

		testName := fmt.Sprintf("keep=%t, test %s", tc.keep, state)
		t.Run(testName, func(t *testing.T) {
			dockerutil.KeepVolumesOnFailure = tc.keep
			mt := mocktesting.NewT(t.Name())

			var volumeName string
			mt.Simulate(func() {
				cli, _ := dockerutil.DockerSetup(mt)

				v, err := cli.VolumeCreate(ctx, volumetypes.CreateOptions{
					Labels: map[string]string{dockerutil.CleanupLabel: mt.Name()},
				})
				require.NoError(t, err)

				volumeName = v.Name

				if !tc.passed {
					mt.Fail()
				}
			})

			require.Equal(t, !tc.passed, mt.Failed())

			_, err := cli.VolumeInspect(ctx, volumeName)
			if !tc.volumeKept {
				require.Truef(t, errdefs.IsNotFound(err), "expected not found error, got %v", err)
				return
			}

			require.NoError(t, err)
			if err := cli.VolumeRemove(ctx, volumeName, true); err != nil {
				t.Logf("failed to remove volume %s: %v", volumeName, err)
			}
		})
	}
}
