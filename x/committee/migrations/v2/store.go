package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

// MigrateStore performs in-place migrations from kava 10 to kava 11.
// The migration includes:
//
// - Update stability committee with new permissions to x/evm & x/evmutil.
func MigrateStore(ctx sdk.Context, k keeper.Keeper) error {
	stabilityCommittee, found := k.GetCommittee(ctx, 1)
	if !found {
		return sdkerrors.Wrap(types.ErrUnknownCommittee, "stability committee not found")
	}

	permissions := stabilityCommittee.GetPermissions()
	for i := 0; i < len(permissions); i++ {
		permission := permissions[i]
		paramsChangePermission, ok := permission.(types.ParamsChangePermission)
		if ok {
			newPermissions := migrateParamsChangePermission(ctx, k, paramsChangePermission)
			permissions[i] = newPermissions
		}
	}

	k.SetCommittee(ctx, stabilityCommittee)

	return nil
}

func migrateParamsChangePermission(ctx sdk.Context, k keeper.Keeper, p types.ParamsChangePermission) types.ParamsChangePermission {
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

	p.AllowedParamsChanges = append(p.AllowedParamsChanges, evmChange, evmutilChange)
	return p
}
