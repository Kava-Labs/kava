package keeper

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// Keeper keeper for the cdp module
type Keeper struct {
	key             storetypes.StoreKey
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
func NewKeeper(cdc codec.Codec, key storetypes.StoreKey, paramstore paramtypes.Subspace, pfk types.PricefeedKeeper,
	ak types.AuctionKeeper, bk types.BankKeeper, ack types.AccountKeeper, maccs map[string][]string,
) Keeper {
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
func (k Keeper) CdpDenomIndexIterator(ctx context.Context, collateralType string) storetypes.Iterator {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.CdpKeyPrefix)
	return storetypes.KVStorePrefixIterator(store, types.DenomIterKey(collateralType))
}

// CdpCollateralRatioIndexIterator returns an sdk.Iterator for all cdps that have collateral denom
// matching denom and collateral:debt ratio LESS THAN targetRatio
func (k Keeper) CdpCollateralRatioIndexIterator(ctx context.Context, collateralType string, targetRatio sdkmath.LegacyDec) storetypes.Iterator {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	return store.Iterator(types.CollateralRatioIterKey(collateralType, sdkmath.LegacyZeroDec()), types.CollateralRatioIterKey(collateralType, targetRatio))
}

// IterateAllCdps iterates over all cdps and performs a callback function
func (k Keeper) IterateAllCdps(ctx context.Context, cb func(cdp types.CDP) (stop bool)) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.CdpKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
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
func (k Keeper) IterateCdpsByCollateralType(ctx context.Context, collateralType string, cb func(cdp types.CDP) (stop bool)) {
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
func (k Keeper) IterateCdpsByCollateralRatio(ctx context.Context, collateralType string, targetRatio sdkmath.LegacyDec, cb func(cdp types.CDP) (stop bool)) {
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
func (k Keeper) GetSliceOfCDPsByRatioAndType(ctx context.Context, cutoffCount sdkmath.Int, targetRatio sdkmath.LegacyDec, collateralType string) (cdps types.CDPs) {
	count := sdkmath.ZeroInt()
	k.IterateCdpsByCollateralRatio(ctx, collateralType, targetRatio, func(cdp types.CDP) bool {
		cdps = append(cdps, cdp)
		count = count.Add(sdkmath.OneInt())
		return count.GTE(cutoffCount)
	})
	return cdps
}

// GetPreviousAccrualTime returns the last time an individual market accrued interest
func (k Keeper) GetPreviousAccrualTime(ctx context.Context, ctype string) (time.Time, bool) {
	fmt.Println("GetPreviousAccrualTime: ", ctype)
	//debug.PrintStack()
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz := store.Get([]byte(ctype))
	fmt.Println("bz: ", bz)
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
func (k Keeper) SetPreviousAccrualTime(ctx context.Context, ctype string, previousAccrualTime time.Time) {
	fmt.Println("SetPreviousAccrualTime: ", ctype, previousAccrualTime)
	//debug.PrintStack()
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cdpMacc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	fmt.Println("SetPreviousAccrualTime 1", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(sdkCtx, cdpMacc.GetAddress(), "debt").Amount.Int64())
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz, err := previousAccrualTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(ctype), bz)
	fmt.Println("SetPreviousAccrualTime 2", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(sdkCtx, cdpMacc.GetAddress(), "debt").Amount.Int64())
}

// GetInterestFactor returns the current interest factor for an individual collateral type
func (k Keeper) GetInterestFactor(ctx context.Context, ctype string) (sdkmath.LegacyDec, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.InterestFactorPrefix)
	bz := store.Get([]byte(ctype))
	if bz == nil {
		return sdkmath.LegacyZeroDec(), false
	}
	var interestFactor sdkmath.LegacyDec
	if err := interestFactor.Unmarshal(bz); err != nil {
		panic(err)
	}
	return interestFactor, true
}

// SetInterestFactor sets the current interest factor for an individual collateral type
func (k Keeper) SetInterestFactor(ctx context.Context, ctype string, interestFactor sdkmath.LegacyDec) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.InterestFactorPrefix)
	bz, err := interestFactor.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(ctype), bz)
}

// IncrementTotalPrincipal increments the total amount of debt that has been drawn with that collateral type
func (k Keeper) IncrementTotalPrincipal(ctx context.Context, collateralType string, principal sdk.Coin) {
	total := k.GetTotalPrincipal(ctx, collateralType, principal.Denom)
	total = total.Add(principal.Amount)
	k.SetTotalPrincipal(ctx, collateralType, principal.Denom, total)
}

// DecrementTotalPrincipal decrements the total amount of debt that has been drawn for a particular collateral type
func (k Keeper) DecrementTotalPrincipal(ctx context.Context, collateralType string, principal sdk.Coin) {
	total := k.GetTotalPrincipal(ctx, collateralType, principal.Denom)
	// NOTE: negative total principal can happen in tests due to rounding errors
	// in fee calculation
	total = sdkmath.MaxInt(total.Sub(principal.Amount), sdkmath.ZeroInt())
	k.SetTotalPrincipal(ctx, collateralType, principal.Denom, total)
}

// GetTotalPrincipal returns the total amount of principal that has been drawn for a particular collateral
func (k Keeper) GetTotalPrincipal(ctx context.Context, collateralType, principalDenom string) (total sdkmath.Int) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.PrincipalKeyPrefix)
	bz := store.Get([]byte(collateralType + principalDenom))
	if bz == nil {
		k.SetTotalPrincipal(ctx, collateralType, principalDenom, sdkmath.ZeroInt())
		return sdkmath.ZeroInt()
	}
	if err := total.Unmarshal(bz); err != nil {
		panic(err)
	}
	return total
}

// SetTotalPrincipal sets the total amount of principal that has been drawn for the input collateral
func (k Keeper) SetTotalPrincipal(ctx context.Context, collateralType, principalDenom string, total sdkmath.Int) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.PrincipalKeyPrefix)
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
