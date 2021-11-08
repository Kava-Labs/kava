package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
)

func TestPackPermissions_Success(t *testing.T) {
	_, err := PackPermissions([]Permission{&GodPermission{}})
	require.NoError(t, err)
}

func TestPackPermissions_Failure(t *testing.T) {
	_, err := PackPermissions([]Permission{nil})
	require.Error(t, err)
}

func TestUnPackPermissions_Success(t *testing.T) {
	packedPermissions, err := PackPermissions([]Permission{&GodPermission{}})
	require.NoError(t, err)
	unpackedPermissions, err := UnpackPermissions(packedPermissions)
	require.NoError(t, err)
	require.Len(t, unpackedPermissions, 1)
	_, ok := unpackedPermissions[0].(*GodPermission)
	require.True(t, ok)
}

func TestUnPackPermissions_Failure(t *testing.T) {
	vote, err := types.NewAnyWithValue(&Vote{ProposalID: 1})
	require.NoError(t, err)
	_, err = UnpackPermissions([]*types.Any{vote})
	require.Error(t, err)
}
