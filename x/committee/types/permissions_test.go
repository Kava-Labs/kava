package types_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/kava-labs/kava/x/committee/types"
	"github.com/stretchr/testify/require"
)

func TestPackPermissions_Success(t *testing.T) {
	_, err := types.PackPermissions([]types.Permission{&types.GodPermission{}})
	require.NoError(t, err)
}

func TestPackPermissions_Failure(t *testing.T) {
	_, err := types.PackPermissions([]types.Permission{nil})
	require.Error(t, err)
}

func TestUnPackPermissions_Success(t *testing.T) {
	packedPermissions, err := types.PackPermissions([]types.Permission{&types.GodPermission{}})
	require.NoError(t, err)
	unpackedPermissions, err := types.UnpackPermissions(packedPermissions)
	require.NoError(t, err)
	require.Len(t, unpackedPermissions, 1)
	_, ok := unpackedPermissions[0].(*types.GodPermission)
	require.True(t, ok)
}

func TestUnPackPermissions_Failure(t *testing.T) {
	vote, err := codectypes.NewAnyWithValue(&types.Vote{ProposalID: 1})
	require.NoError(t, err)
	_, err = types.UnpackPermissions([]*codectypes.Any{vote})
	require.Error(t, err)
}
