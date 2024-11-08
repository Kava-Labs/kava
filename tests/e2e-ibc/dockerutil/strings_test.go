package dockerutil

import (
	"math/rand"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func TestGetHostPort(t *testing.T) {
	for _, tt := range []struct {
		Container types.ContainerJSON
		PortID    string
		Want      string
	}{
		{
			types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							nat.Port("test"): []nat.PortBinding{
								{HostIP: "1.2.3.4", HostPort: "8080"},
								{HostIP: "0.0.0.0", HostPort: "9999"},
							},
						},
					},
				},
			}, "test", "1.2.3.4:8080",
		},
		{
			types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							nat.Port("test"): []nat.PortBinding{
								{HostIP: "0.0.0.0", HostPort: "3000"},
							},
						},
					},
				},
			}, "test", "0.0.0.0:3000",
		},

		{types.ContainerJSON{}, "", ""},
		{types.ContainerJSON{NetworkSettings: &types.NetworkSettings{}}, "does-not-matter", ""},
	} {
		require.Equal(t, tt.Want, GetHostPort(tt.Container, tt.PortID), tt)
	}
}

func TestRandLowerCaseLetterString(t *testing.T) {
	require.Empty(t, RandLowerCaseLetterString(0))

	rand.Seed(1)
	require.Equal(t, "xvlbzgbaicmr", RandLowerCaseLetterString(12))

	rand.Seed(1)
	require.Equal(t, "xvlbzgbaicmrajwwhthctcuaxhxkqf", RandLowerCaseLetterString(30))
}

func TestCondenseHostName(t *testing.T) {
	for _, tt := range []struct {
		HostName, Want string
	}{
		{"", ""},
		{"test", "test"},
		{"some-really-very-incredibly-long-hostname-that-is-greater-than-64-characters", "some-really-very-incredibly-lo_._-is-greater-than-64-characters"},
	} {
		require.Equal(t, tt.Want, CondenseHostName(tt.HostName), tt)
	}
}

func TestSanitizeContainerName(t *testing.T) {
	for _, tt := range []struct {
		Name, Want string
	}{
		{"hello-there", "hello-there"},
		{"hello@there", "hello_there"},
		{"hello@/there", "hello__there"},
		// edge cases
		{"?", "_"},
		{"", ""},
	} {
		require.Equal(t, tt.Want, SanitizeContainerName(tt.Name), tt)
	}
}
