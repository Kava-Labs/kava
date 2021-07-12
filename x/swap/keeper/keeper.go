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

// GetSwapFee returns the swap fee set in the module parameters
func (k Keeper) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return k.GetParams(ctx).SwapFee
}

// GetPool retrieves a pool record from the store
func (k Keeper) GetPool(ctx sdk.Context, poolID string) (types.PoolRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)

	bz := store.Get(types.PoolKey(poolID))
	if bz == nil {
		return types.PoolRecord{}, false
	}

	var record types.PoolRecord
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &record)

	return record, true
}

// SetPool saves a pool record to the store
func (k Keeper) SetPool(ctx sdk.Context, record types.PoolRecord) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(record)
	store.Set(types.PoolKey(record.PoolID), bz)
}

// DeletePool deletes a pool record from the store
func (k Keeper) DeletePool(ctx sdk.Context, poolID string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	store.Delete(types.PoolKey(poolID))
}

// GetDepositorShares gets a share record from the store
func (k Keeper) GetDepositorShares(ctx sdk.Context, depositor sdk.AccAddress, poolID string) (types.ShareRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	bz := store.Get(types.DepositorPoolSharesKey(depositor, poolID))
	if bz == nil {
		return types.ShareRecord{}, false
	}
	var record types.ShareRecord
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &record)
	return record, true
}

// SetDepositorShares saves a share record to the store
func (k Keeper) SetDepositorShares(ctx sdk.Context, record types.ShareRecord) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(record)
	store.Set(types.DepositorPoolSharesKey(record.Depositor, record.PoolID), bz)
}

// DeleteDepositorShares deletes a share record from the store
func (k Keeper) DeleteDepositorShares(ctx sdk.Context, depositor sdk.AccAddress, poolID string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	store.Delete(types.DepositorPoolSharesKey(depositor, poolID))
}
