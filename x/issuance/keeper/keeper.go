package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/issuance/types"
)

// Keeper keeper for the issuance module
type Keeper struct {
	key           storetypes.StoreKey
	cdc           codec.Codec
	paramSubspace paramtypes.Subspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper returns a new keeper
func NewKeeper(cdc codec.Codec, key storetypes.StoreKey, paramstore paramtypes.Subspace, ak types.AccountKeeper, bk types.BankKeeper) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore,
		accountKeeper: ak,
		bankKeeper:    bk,
	}
}

// GetAssetSupply gets an asset's current supply from the store.
func (k Keeper) GetAssetSupply(ctx sdk.Context, denom string) (types.AssetSupply, bool) {
	var assetSupply types.AssetSupply
	store := prefix.NewStore(ctx.KVStore(k.key), types.AssetSupplyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.AssetSupply{}, false
	}
	k.cdc.MustUnmarshal(bz, &assetSupply)
	return assetSupply, true
}

// SetAssetSupply updates an asset's supply
func (k Keeper) SetAssetSupply(ctx sdk.Context, supply types.AssetSupply, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AssetSupplyPrefix)
	store.Set([]byte(denom), k.cdc.MustMarshal(&supply))
}

// IterateAssetSupplies provides an iterator over all stored AssetSupplies.
func (k Keeper) IterateAssetSupplies(ctx sdk.Context, cb func(supply types.AssetSupply) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.AssetSupplyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var supply types.AssetSupply
		k.cdc.MustUnmarshal(iterator.Value(), &supply)
		if cb(supply) {
			break
		}
	}
}

// GetAllAssetSupplies returns all asset supplies from the store
func (k Keeper) GetAllAssetSupplies(ctx sdk.Context) (supplies []types.AssetSupply) {
	k.IterateAssetSupplies(ctx, func(supply types.AssetSupply) bool {
		supplies = append(supplies, supply)
		return false
	})
	return
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	b := store.Get(types.PreviousBlockTimeKey)
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	b, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set(types.PreviousBlockTimeKey, b)
}
