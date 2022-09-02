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
	cdc           codec.Codec
	key           sdk.StoreKey
	paramSubspace types.ParamSubspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	cdpKeeper     types.CdpKeeper
	hardKeeper    types.HardKeeper
	stakingKeeper types.StakingKeeper
	swapKeeper    types.SwapKeeper
	savingsKeeper types.SavingsKeeper
	liquidKeeper  types.LiquidKeeper
	earnKeeper    types.EarnKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc codec.Codec, key sdk.StoreKey, paramstore types.ParamSubspace, bk types.BankKeeper,
	cdpk types.CdpKeeper, hk types.HardKeeper, ak types.AccountKeeper, stk types.StakingKeeper,
	swpk types.SwapKeeper, svk types.SavingsKeeper, lqk types.LiquidKeeper, ek types.EarnKeeper,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		accountKeeper: ak,
		cdc:           cdc,
		key:           key,
		paramSubspace: paramstore,
		bankKeeper:    bk,
		cdpKeeper:     cdpk,
		hardKeeper:    hk,
		stakingKeeper: stk,
		swapKeeper:    swpk,
		savingsKeeper: svk,
		liquidKeeper:  lqk,
		earnKeeper:    ek,
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
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetUSDXMintingClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetUSDXMintingClaim(ctx sdk.Context, c types.USDXMintingClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
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
		k.cdc.MustUnmarshal(iterator.Value(), &c)
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
	b := store.Get([]byte(ctype))
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetPreviousUSDXMintingAccrualTime sets the last time a collateral type accrued USDX minting rewards
func (k Keeper) SetPreviousUSDXMintingAccrualTime(ctx sdk.Context, ctype string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousUSDXMintingRewardAccrualTimeKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(ctype), bz)
}

// IterateUSDXMintingAccrualTimes iterates over all previous USDX minting accrual times and preforms a callback function
func (k Keeper) IterateUSDXMintingAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousUSDXMintingRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var accrualTime time.Time
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
		denom := string(iterator.Key())
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
	if err := factor.Unmarshal(bz); err != nil {
		panic(err)
	}
	return factor, true
}

// SetUSDXMintingRewardFactor sets the current reward factor for an individual collateral type
func (k Keeper) SetUSDXMintingRewardFactor(ctx sdk.Context, ctype string, factor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingRewardFactorKeyPrefix)
	bz, err := factor.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(ctype), bz)
}

// IterateUSDXMintingRewardFactors iterates over all USDX Minting reward factor objects in the store and preforms a callback function
func (k Keeper) IterateUSDXMintingRewardFactors(ctx sdk.Context, cb func(denom string, factor sdk.Dec) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingRewardFactorKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var factor sdk.Dec
		if err := factor.Unmarshal(iterator.Value()); err != nil {
			panic(err)
		}
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
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetHardLiquidityProviderClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetHardLiquidityProviderClaim(ctx sdk.Context, c types.HardLiquidityProviderClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
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
		k.cdc.MustUnmarshal(iterator.Value(), &c)
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
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetDelegatorClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetDelegatorClaim(ctx sdk.Context, c types.DelegatorClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
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
		k.cdc.MustUnmarshal(iterator.Value(), &c)
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
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetSwapClaim sets the claim in the store corresponding to the input address.
func (k Keeper) SetSwapClaim(ctx sdk.Context, c types.SwapClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
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
		k.cdc.MustUnmarshal(iterator.Value(), &c)
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

// GetSavingsClaim returns the claim in the store corresponding the the input address.
func (k Keeper) GetSavingsClaim(ctx sdk.Context, addr sdk.AccAddress) (types.SavingsClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.SavingsClaim{}, false
	}
	var c types.SavingsClaim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetSavingsClaim sets the claim in the store corresponding to the input address.
func (k Keeper) SetSavingsClaim(ctx sdk.Context, c types.SavingsClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)
}

// DeleteSavingsClaim deletes the claim in the store corresponding to the input address.
func (k Keeper) DeleteSavingsClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	store.Delete(owner)
}

// IterateSavingsClaims iterates over all savings claim objects in the store and preforms a callback function
func (k Keeper) IterateSavingsClaims(ctx sdk.Context, cb func(c types.SavingsClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.SavingsClaim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllSavingsClaims returns all savings claim objects in the store
func (k Keeper) GetAllSavingsClaims(ctx sdk.Context) types.SavingsClaims {
	cs := types.SavingsClaims{}
	k.IterateSavingsClaims(ctx, func(c types.SavingsClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetEarnClaim returns the claim in the store corresponding the the input address.
func (k Keeper) GetEarnClaim(ctx sdk.Context, addr sdk.AccAddress) (types.EarnClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.EarnClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.EarnClaim{}, false
	}
	var c types.EarnClaim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetEarnClaim sets the claim in the store corresponding to the input address.
func (k Keeper) SetEarnClaim(ctx sdk.Context, c types.EarnClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.EarnClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)
}

// DeleteEarnClaim deletes the claim in the store corresponding to the input address.
func (k Keeper) DeleteEarnClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.EarnClaimKeyPrefix)
	store.Delete(owner)
}

// IterateEarnClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateEarnClaims(ctx sdk.Context, cb func(c types.EarnClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.EarnClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.EarnClaim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllEarnClaims returns all Claim objects in the store
func (k Keeper) GetAllEarnClaims(ctx sdk.Context) types.EarnClaims {
	cs := types.EarnClaims{}
	k.IterateEarnClaims(ctx, func(c types.EarnClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// SetHardSupplyRewardIndexes sets the current reward indexes for an individual denom
func (k Keeper) SetHardSupplyRewardIndexes(ctx sdk.Context, denom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set([]byte(denom), bz)
}

// GetHardSupplyRewardIndexes gets the current reward indexes for an individual denom
func (k Keeper) GetHardSupplyRewardIndexes(ctx sdk.Context, denom string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardIndexesKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)

	return proto.RewardIndexes, true
}

// IterateHardSupplyRewardIndexes iterates over all Hard supply reward index objects in the store and preforms a callback function
func (k Keeper) IterateHardSupplyRewardIndexes(ctx sdk.Context, cb func(denom string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardSupplyRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		k.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()), proto.RewardIndexes) {
			break
		}
	}
}

func (k Keeper) IterateHardSupplyRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardSupplyRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var accrualTime time.Time
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
		denom := string(iterator.Key())
		if cb(denom, accrualTime) {
			break
		}
	}
}

// SetHardBorrowRewardIndexes sets the current reward indexes for an individual denom
func (k Keeper) SetHardBorrowRewardIndexes(ctx sdk.Context, denom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set([]byte(denom), bz)
}

// GetHardBorrowRewardIndexes gets the current reward indexes for an individual denom
func (k Keeper) GetHardBorrowRewardIndexes(ctx sdk.Context, denom string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardIndexesKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)

	return proto.RewardIndexes, true
}

// IterateHardBorrowRewardIndexes iterates over all Hard borrow reward index objects in the store and preforms a callback function
func (k Keeper) IterateHardBorrowRewardIndexes(ctx sdk.Context, cb func(denom string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardBorrowRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		k.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()), proto.RewardIndexes) {
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
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
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
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)

	return proto.RewardIndexes, true
}

// SetDelegatorRewardIndexes sets the current reward indexes for an individual denom
func (k Keeper) SetDelegatorRewardIndexes(ctx sdk.Context, denom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set([]byte(denom), bz)
}

// IterateDelegatorRewardIndexes iterates over all delegator reward index objects in the store and preforms a callback function
func (k Keeper) IterateDelegatorRewardIndexes(ctx sdk.Context, cb func(denom string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		k.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()), proto.RewardIndexes) {
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
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
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
	if err := blockTime.UnmarshalBinary(bz); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetPreviousHardSupplyRewardAccrualTime sets the last time a denom accrued Hard protocol supply-side rewards
func (k Keeper) SetPreviousHardSupplyRewardAccrualTime(ctx sdk.Context, denom string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardSupplyRewardAccrualTimeKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(denom), bz)
}

// GetPreviousHardBorrowRewardAccrualTime returns the last time a denom accrued Hard protocol borrow-side rewards
func (k Keeper) GetPreviousHardBorrowRewardAccrualTime(ctx sdk.Context, denom string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardBorrowRewardAccrualTimeKeyPrefix)
	b := store.Get([]byte(denom))
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetPreviousHardBorrowRewardAccrualTime sets the last time a denom accrued Hard protocol borrow-side rewards
func (k Keeper) SetPreviousHardBorrowRewardAccrualTime(ctx sdk.Context, denom string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousHardBorrowRewardAccrualTimeKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(denom), bz)
}

// GetPreviousDelegatorRewardAccrualTime returns the last time a denom accrued protocol delegator rewards
func (k Keeper) GetPreviousDelegatorRewardAccrualTime(ctx sdk.Context, denom string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDelegatorRewardAccrualTimeKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(bz); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetPreviousDelegatorRewardAccrualTime sets the last time a denom accrued protocol delegator rewards
func (k Keeper) SetPreviousDelegatorRewardAccrualTime(ctx sdk.Context, denom string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDelegatorRewardAccrualTimeKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(denom), bz)
}

// SetSwapRewardIndexes stores the global reward indexes that track total rewards to a swap pool.
func (k Keeper) SetSwapRewardIndexes(ctx sdk.Context, poolID string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set([]byte(poolID), bz)
}

// GetSwapRewardIndexes fetches the global reward indexes that track total rewards to a swap pool.
func (k Keeper) GetSwapRewardIndexes(ctx sdk.Context, poolID string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapRewardIndexesKeyPrefix)
	bz := store.Get([]byte(poolID))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)
	return proto.RewardIndexes, true
}

// IterateSwapRewardIndexes iterates over all swap reward index objects in the store and preforms a callback function
func (k Keeper) IterateSwapRewardIndexes(ctx sdk.Context, cb func(poolID string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		k.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()), proto.RewardIndexes) {
			break
		}
	}
}

// GetSwapRewardAccrualTime fetches the last time rewards were accrued for a swap pool.
func (k Keeper) GetSwapRewardAccrualTime(ctx sdk.Context, poolID string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSwapRewardAccrualTimeKeyPrefix)
	b := store.Get([]byte(poolID))
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetSwapRewardAccrualTime stores the last time rewards were accrued for a swap pool.
func (k Keeper) SetSwapRewardAccrualTime(ctx sdk.Context, poolID string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSwapRewardAccrualTimeKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(poolID), bz)
}

func (k Keeper) IterateSwapRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSwapRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		poolID := string(iterator.Key())
		var accrualTime time.Time
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
		if cb(poolID, accrualTime) {
			break
		}
	}
}

// SetSavingsRewardIndexes stores the global reward indexes that rewards for an individual denom type
func (k Keeper) SetSavingsRewardIndexes(ctx sdk.Context, denom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set([]byte(denom), bz)
}

// GetSavingsRewardIndexes fetches the global reward indexes that track rewards for an individual denom type
func (k Keeper) GetSavingsRewardIndexes(ctx sdk.Context, denom string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsRewardIndexesKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)
	return proto.RewardIndexes, true
}

// IterateSavingsRewardIndexes iterates over all savings reward index objects in the store and preforms a callback function
func (k Keeper) IterateSavingsRewardIndexes(ctx sdk.Context, cb func(poolID string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		k.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()), proto.RewardIndexes) {
			break
		}
	}
}

// GetSavingsRewardAccrualTime fetches the last time rewards were accrued for an individual denom type
func (k Keeper) GetSavingsRewardAccrualTime(ctx sdk.Context, poolID string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSavingsRewardAccrualTimeKeyPrefix)
	b := store.Get([]byte(poolID))
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetSavingsRewardAccrualTime stores the last time rewards were accrued for a savings deposit denom type
func (k Keeper) SetSavingsRewardAccrualTime(ctx sdk.Context, poolID string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSavingsRewardAccrualTimeKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(poolID), bz)
}

// IterateSavingsRewardAccrualTimesiterates over all the previous savings reward accrual times in the store
func (k Keeper) IterateSavingsRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousSavingsRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		poolID := string(iterator.Key())
		var accrualTime time.Time
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
		if cb(poolID, accrualTime) {
			break
		}
	}
}

// SetEarnRewardIndexes stores the global reward indexes that track total rewards to a earn vault.
func (k Keeper) SetEarnRewardIndexes(ctx sdk.Context, vaultDenom string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.EarnRewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set([]byte(vaultDenom), bz)
}

// GetEarnRewardIndexes fetches the global reward indexes that track total rewards to a earn vault.
func (k Keeper) GetEarnRewardIndexes(ctx sdk.Context, vaultDenom string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.EarnRewardIndexesKeyPrefix)
	bz := store.Get([]byte(vaultDenom))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)
	return proto.RewardIndexes, true
}

// IterateEarnRewardIndexes iterates over all earn reward index objects in the store and preforms a callback function
func (k Keeper) IterateEarnRewardIndexes(ctx sdk.Context, cb func(vaultDenom string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.EarnRewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		k.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()), proto.RewardIndexes) {
			break
		}
	}
}

// GetEarnRewardAccrualTime fetches the last time rewards were accrued for an earn vault.
func (k Keeper) GetEarnRewardAccrualTime(ctx sdk.Context, vaultDenom string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousEarnRewardAccrualTimeKeyPrefix)
	b := store.Get([]byte(vaultDenom))
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetEarnRewardAccrualTime stores the last time rewards were accrued for a earn vault.
func (k Keeper) SetEarnRewardAccrualTime(ctx sdk.Context, vaultDenom string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousEarnRewardAccrualTimeKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(vaultDenom), bz)
}

func (k Keeper) IterateEarnRewardAccrualTimes(ctx sdk.Context, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousEarnRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		poolID := string(iterator.Key())
		var accrualTime time.Time
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
		if cb(poolID, accrualTime) {
			break
		}
	}
}
