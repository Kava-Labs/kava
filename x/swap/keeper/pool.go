package keeper

import (
	"github.com/kava-labs/kava/x/swap/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetPool(ctx sdk.Context, poolName string) (types.Pool, bool) {
	var pool types.Pool
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	bz := store.Get(types.PoolKey(poolName))
	if bz == nil {
		return pool, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &pool)
	return pool, true
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pool)
	store.Set(types.PoolKey(pool.Name()), bz)
}

func (k Keeper) DeletePool(ctx sdk.Context, poolName string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	store.Delete(types.PoolKey(poolName))
}

func (k Keeper) InitializePool(ctx sdk.Context, depositor sdk.AccAddress, amountA, amountB sdk.Coin) error {
	pool, err := types.NewPool(amountA, amountB)
	if err != nil {
		return err
	}
	k.SetPool(ctx, pool)
	k.SetDepositorShares(ctx, depositor, pool.Name(), pool.TotalShares)
	return nil
}
