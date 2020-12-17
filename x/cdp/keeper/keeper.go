package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/cdp/types"
)

// Keeper keeper for the cdp module
type Keeper struct {
	key             sdk.StoreKey
	cdc             *codec.Codec
	paramSubspace   subspace.Subspace
	pricefeedKeeper types.PricefeedKeeper
	supplyKeeper    types.SupplyKeeper
	auctionKeeper   types.AuctionKeeper
	accountKeeper   types.AccountKeeper
	maccPerms       map[string][]string
}

// NewKeeper creates a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, pfk types.PricefeedKeeper,
	ak types.AuctionKeeper, sk types.SupplyKeeper, ack types.AccountKeeper, maccs map[string][]string) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:             key,
		cdc:             cdc,
		paramSubspace:   paramstore,
		pricefeedKeeper: pfk,
		auctionKeeper:   ak,
		supplyKeeper:    sk,
		accountKeeper:   ack,
		maccPerms:       maccs,
	}
}

// CdpDenomIndexIterator returns an sdk.Iterator for all cdps with matching collateral denom
func (k Keeper) CdpDenomIndexIterator(ctx sdk.Context, collateralType string) sdk.Iterator {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db, found := k.GetCollateralTypePrefix(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("denom %s prefix not found", collateralType))
	}
	return sdk.KVStorePrefixIterator(store, types.DenomIterKey(db))
}

// CdpCollateralRatioIndexIterator returns an sdk.Iterator for all cdps that have collateral denom
// matching denom and collateral:debt ratio LESS THAN targetRatio
func (k Keeper) CdpCollateralRatioIndexIterator(ctx sdk.Context, collateralType string, targetRatio sdk.Dec) sdk.Iterator {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	db, found := k.GetCollateralTypePrefix(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("denom %s prefix not found", collateralType))
	}
	return store.Iterator(types.CollateralRatioIterKey(db, sdk.ZeroDec()), types.CollateralRatioIterKey(db, targetRatio))
}

// IterateAllCdps iterates over all cdps and performs a callback function
func (k Keeper) IterateAllCdps(ctx sdk.Context, cb func(cdp types.CDP) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cdp types.CDP
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &cdp)

		if cb(cdp) {
			break
		}
	}
}

// IterateCdpsByCollateralType iterates over cdps with matching denom and performs a callback function
func (k Keeper) IterateCdpsByCollateralType(ctx sdk.Context, collateralType string, cb func(cdp types.CDP) (stop bool)) {
	iterator := k.CdpDenomIndexIterator(ctx, collateralType)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cdp types.CDP
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &cdp)
		if cb(cdp) {
			break
		}
	}
}

// IterateCdpsByCollateralRatio iterate over cdps with collateral denom equal to denom and
// collateral:debt ratio LESS THAN targetRatio and performs a callback function.
func (k Keeper) IterateCdpsByCollateralRatio(ctx sdk.Context, collateralType string, targetRatio sdk.Dec, cb func(cdp types.CDP) (stop bool)) {
	iterator := k.CdpCollateralRatioIndexIterator(ctx, collateralType, targetRatio)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		_, id, _ := types.SplitCollateralRatioKey(iterator.Key())
		cdp, found := k.GetCDP(ctx, collateralType, id)
		if !found {
			panic(fmt.Sprintf("cdp %d does not exist", id))
		}
		if cb(cdp) {
			break
		}

	}
}

// SetSavingsRateDistributed sets the SavingsRateDistributed in the store
func (k Keeper) SetSavingsRateDistributed(ctx sdk.Context, totalDistributed sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsRateDistributedKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(totalDistributed)
	store.Set([]byte{}, bz)
}

// GetSavingsRateDistributed gets the SavingsRateDistributed from the store
func (k Keeper) GetSavingsRateDistributed(ctx sdk.Context) sdk.Int {
	savingsRateDistributed := sdk.ZeroInt()
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsRateDistributedKey)
	bz := store.Get([]byte{})
	if bz == nil {
		return savingsRateDistributed
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &savingsRateDistributed)
	return savingsRateDistributed
}

// GetPreviousAccrualTime returns the last time an individual market accrued interest
func (k Keeper) GetPreviousAccrualTime(ctx sdk.Context, ctype string) (time.Time, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return time.Time{}, false
	}
	var previousAccrualTime time.Time
	k.cdc.MustUnmarshalBinaryBare(bz, &previousAccrualTime)
	return previousAccrualTime, true
}

// SetPreviousAccrualTime sets the most recent accrual time for a particular market
func (k Keeper) SetPreviousAccrualTime(ctx sdk.Context, ctype string, previousAccrualTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz := k.cdc.MustMarshalBinaryBare(previousAccrualTime)
	store.Set([]byte(ctype), bz)
}

// GetInterestFactor returns the current interest factor for an individual collateral type
func (k Keeper) GetInterestFactor(ctx sdk.Context, ctype string) (sdk.Dec, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.InterestFactorPrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return sdk.ZeroDec(), false
	}
	var interestFactor sdk.Dec
	k.cdc.MustUnmarshalBinaryBare(bz, &interestFactor)
	return interestFactor, true
}

// SetInterestFactor sets the current interest factor for an individual collateral type
func (k Keeper) SetInterestFactor(ctx sdk.Context, ctype string, interestFactor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.InterestFactorPrefix)
	bz := k.cdc.MustMarshalBinaryBare(interestFactor)
	store.Set([]byte(ctype), bz)
}

// IncrementTotalPrincipal increments the total amount of debt that has been drawn with that collateral type
func (k Keeper) IncrementTotalPrincipal(ctx sdk.Context, collateralType string, principal sdk.Coin) {
	total := k.GetTotalPrincipal(ctx, collateralType, principal.Denom)
	total = total.Add(principal.Amount)
	k.SetTotalPrincipal(ctx, collateralType, principal.Denom, total)
}

// DecrementTotalPrincipal decrements the total amount of debt that has been drawn for a particular collateral type
func (k Keeper) DecrementTotalPrincipal(ctx sdk.Context, collateralType string, principal sdk.Coin) {
	total := k.GetTotalPrincipal(ctx, collateralType, principal.Denom)
	// NOTE: negative total principal can happen in tests due to rounding errors
	// in fee calculation
	total = sdk.MaxInt(total.Sub(principal.Amount), sdk.ZeroInt())
	k.SetTotalPrincipal(ctx, collateralType, principal.Denom, total)
}

// GetTotalPrincipal returns the total amount of principal that has been drawn for a particular collateral
func (k Keeper) GetTotalPrincipal(ctx sdk.Context, collateralType, principalDenom string) (total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	bz := store.Get([]byte(collateralType + principalDenom))
	if bz == nil {
		k.SetTotalPrincipal(ctx, collateralType, principalDenom, sdk.ZeroInt())
		return sdk.ZeroInt()
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &total)
	return total
}

// SetTotalPrincipal sets the total amount of principal that has been drawn for the input collateral
func (k Keeper) SetTotalPrincipal(ctx sdk.Context, collateralType, principalDenom string, total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	_, found := k.GetCollateralTypePrefix(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	store.Set([]byte(collateralType+principalDenom), k.cdc.MustMarshalBinaryLengthPrefixed(total))
}
