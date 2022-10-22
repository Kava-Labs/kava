package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// MigrateStore performs in-place migrations from kava 10 to kava 11.
// The migration includes:
//
// - Update stability committee with new permissions to x/evm, x/evmutil, x/savings & x/earn.
func MigrateStore(ctx sdk.Context, k keeper.Keeper) error {
	stabilityCommittee, found := k.GetCommittee(ctx, 1)
	if !found {
		return sdkerrors.Wrap(types.ErrUnknownCommittee, "stability committee not found")
	}

	permissions := stabilityCommittee.GetPermissions()
	for i := 0; i < len(permissions); i++ {
		permission := permissions[i]
		switch targetPermission := permission.(type) {
		case *types.ParamsChangePermission:
			{
				newPermissions := migrateStabilityCommitteeParamsChangePermission(ctx, k, targetPermission)
				permissions[i] = newPermissions
			}
		}
	}
	stabilityCommittee.SetPermissions(permissions)
	k.SetCommittee(ctx, stabilityCommittee)
	return nil
}

func migrateStabilityCommitteeParamsChangePermission(ctx sdk.Context, k keeper.Keeper, p *types.ParamsChangePermission) *types.ParamsChangePermission {
	// allow all changes to x/evm eip712 allowed msgs
	evmChange := types.AllowedParamsChange{
		Subspace: evmtypes.ModuleName,
		Key:      string(evmtypes.ParamStoreKeyEIP712AllowedMsgs),
	}

	// allow all changes to x/evmutil enabled conversion pairs
	evmutilChange := types.AllowedParamsChange{
		Subspace: evmutiltypes.ModuleName,
		Key:      string(evmutiltypes.KeyEnabledConversionPairs),
	}

	// allow all changes to x/savings supported denoms
	savingsChange := types.AllowedParamsChange{
		Subspace: savingstypes.ModuleName,
		Key:      string(savingstypes.KeySupportedDenoms),
	}

	// allow all changes to x/earn AllowedVaults
	earnChange := types.AllowedParamsChange{
		Subspace: earntypes.ModuleName,
		Key:      string(earntypes.KeyAllowedVaults),
	}

	// add all new param changes
	p.AllowedParamsChanges = append(
		p.AllowedParamsChanges,
		evmChange, evmutilChange, savingsChange, earnChange,
	)
	return p
}
