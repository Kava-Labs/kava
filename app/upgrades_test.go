package app_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestUpdateStabilityCommitteePermissions(t *testing.T) {
	tapp := app.NewTestApp()
	tapp.InitializeFromGenesisStates()
	genesisTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := tapp.NewContext(false, tmproto.Header{Height: 1, Time: genesisTime})

	ck := tapp.GetCommitteeKeeper()

	testCommittee, err := committeetypes.NewMemberCommittee(
		1,
		"Kava Stability Committee",
		[]sdk.AccAddress{
			sdk.AccAddress([]byte("addr1")),
		},
		[]committeetypes.Permission{
			&committeetypes.ParamsChangePermission{
				AllowedParamsChanges: committeetypes.AllowedParamsChanges{
					{
						Subspace:                   cdptypes.ModuleName,
						Key:                        string(cdptypes.KeyGlobalDebtLimit),
						SingleSubparamAllowedAttrs: nil,
						MultiSubparamsRequirements: nil,
					},
					{
						Subspace:                   cdptypes.ModuleName,
						Key:                        string(cdptypes.KeySurplusLot),
						SingleSubparamAllowedAttrs: nil,
						MultiSubparamsRequirements: nil,
					},
					{
						Subspace:                   evmtypes.ModuleName,
						Key:                        "EIP712AllowedMsgs",
						SingleSubparamAllowedAttrs: nil,
						MultiSubparamsRequirements: nil,
					},
					{
						Subspace:                   evmutiltypes.ModuleName,
						Key:                        string(evmutiltypes.KeyEnabledConversionPairs),
						SingleSubparamAllowedAttrs: nil,
						MultiSubparamsRequirements: nil,
					},
				},
			},
		},
		sdk.NewDecWithPrec(5, 1),
		604800*time.Second,
		committeetypes.TALLY_OPTION_FIRST_PAST_THE_POST,
	)
	require.NoError(t, err)

	ck.SetCommittee(ctx, testCommittee)

	app.AddNewPermissionsToStabilityCommittee(
		ctx,
		ck,
		1,
	)

	committee, found := ck.GetCommittee(ctx, 1)
	require.True(t, found)

	permissions := committee.GetPermissions()
	require.Len(t, permissions, 4, "should have 4 total after adding 3 in update")

	allowedParams := permissions[0].(*committeetypes.ParamsChangePermission).AllowedParamsChanges
	assert.Len(t, allowedParams, 4, "should be unchanged")

	// Test removing permissions
	app.RemoveEVMCommitteePermissions(
		ctx,
		ck,
		1,
	)

	// Refetch committee after removing permissions
	committee, found = ck.GetCommittee(ctx, 1)
	require.True(t, found)

	permissions = committee.GetPermissions()

	allowedParams = permissions[0].(*committeetypes.ParamsChangePermission).AllowedParamsChanges
	assert.Len(t, allowedParams, 3, "should have 3 allowed params after removing x/evm")
	require.Equal(
		t,
		committeetypes.AllowedParamsChanges{
			{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyGlobalDebtLimit),
				SingleSubparamAllowedAttrs: nil,
				MultiSubparamsRequirements: nil,
			},
			{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeySurplusLot),
				SingleSubparamAllowedAttrs: nil,
				MultiSubparamsRequirements: nil,
			},
			{
				Subspace:                   evmutiltypes.ModuleName,
				Key:                        string(evmutiltypes.KeyEnabledConversionPairs),
				SingleSubparamAllowedAttrs: nil,
				MultiSubparamsRequirements: nil,
			},
		},
		allowedParams,
		"x/evm should be removed from allowed params",
	)
}
