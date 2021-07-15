package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// Keeper keeper for the incentive module
type Keeper struct {
	accountKeeper types.AccountKeeper
	cdc           *codec.Codec
	key           sdk.StoreKey
	paramSubspace types.ParamSubspace
	supplyKeeper  types.SupplyKeeper
	cdpKeeper     types.CdpKeeper
	hardKeeper    types.HardKeeper
	stakingKeeper types.StakingKeeper
	swapKeeper    types.SwapKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramstore types.ParamSubspace, sk types.SupplyKeeper,
	cdpk types.CdpKeeper, hk types.HardKeeper, ak types.AccountKeeper, stk types.StakingKeeper,
	swpk types.SwapKeeper,
) Keeper {

	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		accountKeeper: ak,
		cdc:           cdc,
		key:           key,
		paramSubspace: paramstore,
		supplyKeeper:  sk,
		cdpKeeper:     cdpk,
		hardKeeper:    hk,
		stakingKeeper: stk,
		swapKeeper:    swpk,
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
		denom := string(iterator.Key())
		var accrualTime time.Time
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &accrualTime)
		if cb(denom, accrualTime) {
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

// IterateUSDXMintingRewardFactors iterates over all USDX Minting reward factor objects in the store and preforms a callback function
func (k Keeper) IterateUSDXMintingRewardFactors(ctx sdk.Context, cb func(denom string, factor sdk.Dec) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingRewardFactorKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var factor sdk.Dec
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &factor)
		if cb(string(iterator.Key()), factor) {
			break
		}
	}
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

// GetDelegatorClaim returns the claim in the store corresponding the the input address collateral type and id and a boolean for if the claim was found
func (k Keeper) GetDelegatorClaim(ctx sdk.Context, addr sdk.AccAddress) (types.DelegatorClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.DelegatorClaim{}, false
	}
	var c types.DelegatorClaim
	k.cdc.MustUnmarshalBinaryBare(bz, &c)
	return c, true
}

// SetDelegatorClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetDelegatorClaim(ctx sdk.Context, c types.DelegatorClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(c)
	store.Set(c.Owner, bz)
}

// DeleteDelegatorClaim deletes the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) DeleteDelegatorClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	store.Delete(owner)
}

// IterateDelegatorClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateDelegatorClaims(ctx sdk.Context, cb func(c types.DelegatorClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.DelegatorClaim
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllDelegatorClaims returns all DelegatorClaim objects in the store
func (k Keeper) GetAllDelegatorClaims(ctx sdk.Context) types.DelegatorClaims {
	cs := types.DelegatorClaims{}
	k.IterateDelegatorClaims(ctx, func(c types.DelegatorClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetSwapClaim returns the claim in the store corresponding the the input address.
func (k Keeper) GetSwapClaim(ctx sdk.Context, addr sdk.AccAddress) (types.SwapClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.SwapClaim{}, false
	}
	var c types.SwapClaim
	k.cdc.MustUnmarshalBinaryBare(bz, &c)
	return c, true
}

// SetSwapClaim sets the claim in the store corresponding to the input address.
func (k Keeper) SetSwapClaim(ctx sdk.Context, c types.SwapClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(c)
	store.Set(c.Owner, bz)
}

// DeleteSwapClaim deletes the claim in the store corresponding to the input address.
func (k Keeper) DeleteSwapClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	store.Delete(owner)
}

// IterateSwapClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateSwapClaims(ctx sdk.Context, cb func(c types.SwapClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.SwapClaim
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllSwapClaims returns all Claim objects in the store
func (k Keeper) GetAllSwapClaims(ctx sdk.Context) types.SwapClaims {
	cs := types.SwapClaims{}
	k.IterateSwapClaims(ctx, func(c types.SwapClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// SetHardSupplyRewardIndexes sets the current reward indexes for an individual denom
func (k Keeper) SetHardSupplyRewardIndexes(ctx sdk.Context, denom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(indexes)
	store.Set([]byte(denom), bz)
}

// GetHardSupplyRewardIndexes gets the current reward indexes for an individual denom
func (k Keeper) GetHardSupplyRewardIndexes(ctx sdk.Context, denom string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardIndexesKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var rewardIndexes types.RewardIndexes
	k.cdc.MustUnmarshalBinaryBare(bz, &rewardIndexes)
	return rewardIndexes, true
}

// IterateHardSupplyRewardIndexes iterates over all Hard supply reward index objects in the store and preforms a callback function
func (k Keeper) IterateHardSupplyRewardIndexes(ctx sdk.Context, cb func(denom string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var indexes types.RewardIndexes
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &indexes)
		if cb(string(iterator.Key()), indexes) {
			break
		}
	}
}

func (k Keeper) IterateHardSupplyRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardSupplyRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key())
		var accrualTime time.Time
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &accrualTime)
		if cb(denom, accrualTime) {
			break
		}
	}
}

// SetHardBorrowRewardIndexes sets the current reward indexes for an individual denom
func (k Keeper) SetHardBorrowRewardIndexes(ctx sdk.Context, denom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(indexes)
	store.Set([]byte(denom), bz)
}

// GetHardBorrowRewardIndexes gets the current reward indexes for an individual denom
func (k Keeper) GetHardBorrowRewardIndexes(ctx sdk.Context, denom string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardIndexesKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var rewardIndexes types.RewardIndexes
	k.cdc.MustUnmarshalBinaryBare(bz, &rewardIndexes)
	return rewardIndexes, true
}

// IterateHardBorrowRewardIndexes iterates over all Hard borrow reward index objects in the store and preforms a callback function
func (k Keeper) IterateHardBorrowRewardIndexes(ctx sdk.Context, cb func(denom string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var indexes types.RewardIndexes
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &indexes)
		if cb(string(iterator.Key()), indexes) {
			break
		}
	}
}

func (k Keeper) IterateHardBorrowRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardBorrowRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key())
		var accrualTime time.Time
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &accrualTime)
		if cb(denom, accrualTime) {
			break
		}
	}
}

// GetDelegatorRewardIndexes gets the current reward indexes for an individual denom
func (k Keeper) GetDelegatorRewardIndexes(ctx sdk.Context, denom string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorRewardIndexesKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var rewardIndexes types.RewardIndexes
	k.cdc.MustUnmarshalBinaryBare(bz, &rewardIndexes)
	return rewardIndexes, true
}

// SetDelegatorRewardIndexes sets the current reward indexes for an individual denom
func (k Keeper) SetDelegatorRewardIndexes(ctx sdk.Context, denom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(indexes)
	store.Set([]byte(denom), bz)
}

// IterateDelegatorRewardIndexes iterates over all delegator reward index objects in the store and preforms a callback function
func (k Keeper) IterateDelegatorRewardIndexes(ctx sdk.Context, cb func(denom string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var indexes types.RewardIndexes
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &indexes)
		if cb(string(iterator.Key()), indexes) {
			break
		}
	}
}

func (k Keeper) IterateDelegatorRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDelegatorRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key())
		var accrualTime time.Time
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &accrualTime)
		if cb(denom, accrualTime) {
			break
		}
	}
}

// GetPreviousHardSupplyRewardAccrualTime returns the last time a denom accrued Hard protocol supply-side rewards
func (k Keeper) GetPreviousHardSupplyRewardAccrualTime(ctx sdk.Context, denom string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardSupplyRewardAccrualTimeKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &blockTime)
	return blockTime, true
}

// SetPreviousHardSupplyRewardAccrualTime sets the last time a denom accrued Hard protocol supply-side rewards
func (k Keeper) SetPreviousHardSupplyRewardAccrualTime(ctx sdk.Context, denom string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardSupplyRewardAccrualTimeKeyPrefix)
	store.Set([]byte(denom), k.cdc.MustMarshalBinaryBare(blockTime))
}

// GetPreviousHardBorrowRewardAccrualTime returns the last time a denom accrued Hard protocol borrow-side rewards
func (k Keeper) GetPreviousHardBorrowRewardAccrualTime(ctx sdk.Context, denom string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardBorrowRewardAccrualTimeKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &blockTime)
	return blockTime, true
}

// SetPreviousHardBorrowRewardAccrualTime sets the last time a denom accrued Hard protocol borrow-side rewards
func (k Keeper) SetPreviousHardBorrowRewardAccrualTime(ctx sdk.Context, denom string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardBorrowRewardAccrualTimeKeyPrefix)
	store.Set([]byte(denom), k.cdc.MustMarshalBinaryBare(blockTime))
}

// GetPreviousDelegatorRewardAccrualTime returns the last time a denom accrued protocol delegator rewards
func (k Keeper) GetPreviousDelegatorRewardAccrualTime(ctx sdk.Context, denom string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDelegatorRewardAccrualTimeKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &blockTime)
	return blockTime, true
}

// SetPreviousDelegatorRewardAccrualTime sets the last time a denom accrued protocol delegator rewards
func (k Keeper) SetPreviousDelegatorRewardAccrualTime(ctx sdk.Context, denom string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDelegatorRewardAccrualTimeKeyPrefix)
	store.Set([]byte(denom), k.cdc.MustMarshalBinaryBare(blockTime))
}

// SetSwapRewardIndexes stores the global reward indexes that track total rewards to a swap pool.
func (k Keeper) SetSwapRewardIndexes(ctx sdk.Context, poolID string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(indexes)
	store.Set([]byte(poolID), bz)
}

// GetSwapRewardIndexes fetches the global reward indexes that track total rewards to a swap pool.
func (k Keeper) GetSwapRewardIndexes(ctx sdk.Context, poolID string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapRewardIndexesKeyPrefix)
	bz := store.Get([]byte(poolID))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var rewardIndexes types.RewardIndexes
	k.cdc.MustUnmarshalBinaryBare(bz, &rewardIndexes)
	return rewardIndexes, true
}

// IterateSwapRewardIndexes iterates over all swap reward index objects in the store and preforms a callback function
func (k Keeper) IterateSwapRewardIndexes(ctx sdk.Context, cb func(poolID string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var indexes types.RewardIndexes
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &indexes)
		if cb(string(iterator.Key()), indexes) {
			break
		}
	}
}

// GetSwapRewardAccrualTime fetches the last time rewards were accrued for a swap pool.
func (k Keeper) GetSwapRewardAccrualTime(ctx sdk.Context, poolID string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSwapRewardAccrualTimeKeyPrefix)
	bz := store.Get([]byte(poolID))
	if bz == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &blockTime)
	return blockTime, true
}

// SetSwapRewardAccrualTime stores the last time rewards were accrued for a swap pool.
func (k Keeper) SetSwapRewardAccrualTime(ctx sdk.Context, poolID string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSwapRewardAccrualTimeKeyPrefix)
	store.Set([]byte(poolID), k.cdc.MustMarshalBinaryBare(blockTime))
}

func (k Keeper) IterateSwapRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSwapRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		poolID := string(iterator.Key())
		var accrualTime time.Time
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &accrualTime)
		if cb(poolID, accrualTime) {
			break
		}
	}
}
