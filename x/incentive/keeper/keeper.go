package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/incentive/types"
)

// Keeper keeper for the incentive module
type Keeper struct {
	accountKeeper types.AccountKeeper
	cdc           *codec.Codec
	cdpKeeper     types.CdpKeeper
	key           sdk.StoreKey
	paramSubspace subspace.Subspace
	supplyKeeper  types.SupplyKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, sk types.SupplyKeeper,
	cdpk types.CdpKeeper, ak types.AccountKeeper,
) Keeper {

	return Keeper{
		accountKeeper: ak,
		cdc:           cdc,
		cdpKeeper:     cdpk,
		key:           key,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
		supplyKeeper:  sk,
	}
}

// GetUSDXMintingClaim returns the claim in the store corresponding the the input address collateral type and id and a boolean for if the claim was found
func (k Keeper) GetUSDXMintingClaim(ctx sdk.Context, addr sdk.AccAddress) (types.USDXMintingClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.USDXMintingClaim{}, false
	}
	var c types.USDXMintingClaim
	k.cdc.MustUnmarshalBinaryBare(bz, &c)
	return c, true
}

// SetUSDXMintingClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetUSDXMintingClaim(ctx sdk.Context, c types.USDXMintingClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(c)
	store.Set(c.Owner, bz)

}

// DeleteUSDXMintingClaim deletes the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) DeleteUSDXMintingClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	store.Delete(owner)
}

// IterateUSDXMintingClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateUSDXMintingClaims(ctx sdk.Context, cb func(c types.USDXMintingClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.USDXMintingClaim
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllUSDXMintingClaims returns all Claim objects in the store
func (k Keeper) GetAllUSDXMintingClaims(ctx sdk.Context) types.USDXMintingClaims {
	cs := types.USDXMintingClaims{}
	k.IterateUSDXMintingClaims(ctx, func(c types.USDXMintingClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetPreviousUSDXMintingAccrualTime returns the last time a collateral type accrued USDX minting rewards
func (k Keeper) GetPreviousUSDXMintingAccrualTime(ctx sdk.Context, ctype string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousUSDXMintingRewardAccrualTimeKeyPrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &blockTime)
	return blockTime, true
}

// SetPreviousUSDXMintingAccrualTime sets the last time a collateral type accrued USDX minting rewards
func (k Keeper) SetPreviousUSDXMintingAccrualTime(ctx sdk.Context, ctype string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousUSDXMintingRewardAccrualTimeKeyPrefix)
	store.Set([]byte(ctype), k.cdc.MustMarshalBinaryBare(blockTime))
}

// IterateUSDXMintingAccrualTimes iterates over all previous USDX minting accrual times and preforms a callback function
func (k Keeper) IterateUSDXMintingAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousUSDXMintingRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var accrualTime time.Time
		var collateralType string
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &collateralType)
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &accrualTime)
		if cb(collateralType, accrualTime) {
			break
		}
	}
}

// GetUSDXMintingRewardFactor returns the current reward factor for an individual collateral type
func (k Keeper) GetUSDXMintingRewardFactor(ctx sdk.Context, ctype string) (factor sdk.Dec, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingRewardFactorKeyPrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return sdk.ZeroDec(), false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &factor)
	return factor, true
}

// SetUSDXMintingRewardFactor sets the current reward factor for an individual collateral type
func (k Keeper) SetUSDXMintingRewardFactor(ctx sdk.Context, ctype string, factor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingRewardFactorKeyPrefix)
	store.Set([]byte(ctype), k.cdc.MustMarshalBinaryBare(factor))
}

// GetHardLiquidityProviderClaim returns the claim in the store corresponding the the input address collateral type and id and a boolean for if the claim was found
func (k Keeper) GetHardLiquidityProviderClaim(ctx sdk.Context, addr sdk.AccAddress) (types.HardLiquidityProviderClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.HardLiquidityProviderClaim{}, false
	}
	var c types.HardLiquidityProviderClaim
	k.cdc.MustUnmarshalBinaryBare(bz, &c)
	return c, true
}

// SetHardLiquidityProviderClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetHardLiquidityProviderClaim(ctx sdk.Context, c types.HardLiquidityProviderClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(c)
	store.Set(c.Owner, bz)

}

// DeleteHardLiquidityProviderClaim deletes the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) DeleteHardLiquidityProviderClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	store.Delete(owner)
}

// IterateHardLiquidityProviderClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateHardLiquidityProviderClaims(ctx sdk.Context, cb func(c types.HardLiquidityProviderClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.HardLiquidityProviderClaim
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllHardLiquidityProviderClaims returns all Claim objects in the store
func (k Keeper) GetAllHardLiquidityProviderClaims(ctx sdk.Context) types.HardLiquidityProviderClaims {
	cs := types.HardLiquidityProviderClaims{}
	k.IterateHardLiquidityProviderClaims(ctx, func(c types.HardLiquidityProviderClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetHardSupplyRewardFactor returns the current reward factor for an individual collateral type
func (k Keeper) GetHardSupplyRewardFactor(ctx sdk.Context, ctype string) (factor sdk.Dec, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardFactorKeyPrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return sdk.ZeroDec(), false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &factor)
	return factor, true
}

// SetHardSupplyRewardFactor sets the current reward factor for an individual collateral type
func (k Keeper) SetHardSupplyRewardFactor(ctx sdk.Context, ctype string, factor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardFactorKeyPrefix)
	store.Set([]byte(ctype), k.cdc.MustMarshalBinaryBare(factor))
}

// GetHardBorrowRewardFactor returns the current reward factor for an individual collateral type
func (k Keeper) GetHardBorrowRewardFactor(ctx sdk.Context, ctype string) (factor sdk.Dec, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardFactorKeyPrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return sdk.ZeroDec(), false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &factor)
	return factor, true
}

// SetHardBorrowRewardFactor sets the current reward factor for an individual collateral type
func (k Keeper) SetHardBorrowRewardFactor(ctx sdk.Context, ctype string, factor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardFactorKeyPrefix)
	store.Set([]byte(ctype), k.cdc.MustMarshalBinaryBare(factor))
}

// GetHardDelegatorRewardFactor returns the current reward factor for an individual collateral type
func (k Keeper) GetHardDelegatorRewardFactor(ctx sdk.Context, ctype string) (factor sdk.Dec, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardDelegatorRewardFactorKeyPrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return sdk.ZeroDec(), false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &factor)
	return factor, true
}

// SetHardDelegatorRewardFactor sets the current reward factor for an individual collateral type
func (k Keeper) SetHardDelegatorRewardFactor(ctx sdk.Context, ctype string, factor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardDelegatorRewardFactorKeyPrefix)
	store.Set([]byte(ctype), k.cdc.MustMarshalBinaryBare(factor))
}
