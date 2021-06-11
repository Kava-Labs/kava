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
