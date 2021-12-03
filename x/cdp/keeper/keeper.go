package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// Keeper keeper for the cdp module
type Keeper struct {
	key             sdk.StoreKey
	cdc             codec.Codec
	paramSubspace   paramtypes.Subspace
	pricefeedKeeper types.PricefeedKeeper
	auctionKeeper   types.AuctionKeeper
	bankKeeper      types.BankKeeper
	accountKeeper   types.AccountKeeper
	hooks           types.CDPHooks
	maccPerms       map[string][]string
}

// NewKeeper creates a new keeper
func NewKeeper(cdc codec.Codec, key sdk.StoreKey, paramstore paramtypes.Subspace, pfk types.PricefeedKeeper,
	ak types.AuctionKeeper, bk types.BankKeeper, ack types.AccountKeeper, maccs map[string][]string) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:             key,
		cdc:             cdc,
		paramSubspace:   paramstore,
		pricefeedKeeper: pfk,
		auctionKeeper:   ak,
		bankKeeper:      bk,
		accountKeeper:   ack,
		hooks:           nil,
		maccPerms:       maccs,
	}
}

// SetHooks adds hooks to the keeper.
func (k *Keeper) SetHooks(hooks types.CDPHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set cdp hooks twice")
	}
	k.hooks = hooks
	return k
}

// CdpDenomIndexIterator returns an sdk.Iterator for all cdps with matching collateral denom
func (k Keeper) CdpDenomIndexIterator(ctx sdk.Context, collateralType string) sdk.Iterator {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	return sdk.KVStorePrefixIterator(store, types.DenomIterKey(collateralType))
}

// CdpCollateralRatioIndexIterator returns an sdk.Iterator for all cdps that have collateral denom
// matching denom and collateral:debt ratio LESS THAN targetRatio
func (k Keeper) CdpCollateralRatioIndexIterator(ctx sdk.Context, collateralType string, targetRatio sdk.Dec) sdk.Iterator {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	return store.Iterator(types.CollateralRatioIterKey(collateralType, sdk.ZeroDec()), types.CollateralRatioIterKey(collateralType, targetRatio))
}

// IterateAllCdps iterates over all cdps and performs a callback function
func (k Keeper) IterateAllCdps(ctx sdk.Context, cb func(cdp types.CDP) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cdp types.CDP
		k.cdc.MustUnmarshal(iterator.Value(), &cdp)

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
		k.cdc.MustUnmarshal(iterator.Value(), &cdp)
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

// GetSliceOfCDPsByRatioAndType returns a slice of cdps of size equal to the input cutoffCount
// sorted by target ratio in ascending order (ie, the lowest collateral:debt ratio cdps are returned first)
func (k Keeper) GetSliceOfCDPsByRatioAndType(ctx sdk.Context, cutoffCount sdk.Int, targetRatio sdk.Dec, collateralType string) (cdps types.CDPs) {
	count := sdk.ZeroInt()
	k.IterateCdpsByCollateralRatio(ctx, collateralType, targetRatio, func(cdp types.CDP) bool {
		cdps = append(cdps, cdp)
		count = count.Add(sdk.OneInt())
		return count.GTE(cutoffCount)
	})
	return cdps
}

// GetPreviousAccrualTime returns the last time an individual market accrued interest
func (k Keeper) GetPreviousAccrualTime(ctx sdk.Context, ctype string) (time.Time, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return time.Time{}, false
	}
	var previousAccrualTime time.Time
	if err := previousAccrualTime.UnmarshalBinary(bz); err != nil {
		panic(err)
	}
	return previousAccrualTime, true
}

// SetPreviousAccrualTime sets the most recent accrual time for a particular market
func (k Keeper) SetPreviousAccrualTime(ctx sdk.Context, ctype string, previousAccrualTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz, err := previousAccrualTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
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
	if err := interestFactor.Unmarshal(bz); err != nil {
		panic(err)
	}
	return interestFactor, true
}

// SetInterestFactor sets the current interest factor for an individual collateral type
func (k Keeper) SetInterestFactor(ctx sdk.Context, ctype string, interestFactor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.InterestFactorPrefix)
	bz, err := interestFactor.Marshal()
	if err != nil {
		panic(err)
	}
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
	if err := total.Unmarshal(bz); err != nil {
		panic(err)
	}
	return total
}

// SetTotalPrincipal sets the total amount of principal that has been drawn for the input collateral
func (k Keeper) SetTotalPrincipal(ctx sdk.Context, collateralType, principalDenom string, total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	_, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	bz, err := total.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(collateralType+principalDenom), bz)
}
