package keeper

import (
	"github.com/kava-labs/kava/x/swap/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetDepositorShares(ctx sdk.Context, depositor sdk.AccAddress, poolName string) (sdk.Int, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	bz := store.Get(types.DepositorPoolSharesKey(depositor, poolName))
	if bz == nil {
		return sdk.ZeroInt(), false
	}
	var shares sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &shares)
	return shares, true
}

func (k Keeper) SetDepositorShares(ctx sdk.Context, depositor sdk.AccAddress, poolName string, shares sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(shares)
	store.Set(types.DepositorPoolSharesKey(depositor, poolName), bz)
}

func (k Keeper) DeleteDepositorShares(ctx sdk.Context, depositor sdk.AccAddress, poolName string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	store.Delete(types.DepositorPoolSharesKey(depositor, poolName))
}
