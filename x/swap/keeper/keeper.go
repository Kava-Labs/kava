package keeper

import (
	"github.com/kava-labs/kava/x/swap/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Keeper keeper for the swap module
type Keeper struct {
	key           sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	accountKeeper types.AccountKeeper
	supplyKeeper  types.SupplyKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc *codec.Codec,
	key sdk.StoreKey,
	paramstore subspace.Subspace,
	accountKeeper types.AccountKeeper,
	supplyKeeper types.SupplyKeeper,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore,
		accountKeeper: accountKeeper,
		supplyKeeper:  supplyKeeper,
	}
}

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

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
