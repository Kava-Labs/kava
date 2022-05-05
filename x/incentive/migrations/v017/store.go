package v017

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, paramstore types.ParamSubspace) error {
	migrateParamsStore(ctx, paramstore)

	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore types.ParamSubspace) {
	paramstore.Set(ctx, types.KeySavingsRewardPeriods, types.DefaultMultiRewardPeriods)
}
