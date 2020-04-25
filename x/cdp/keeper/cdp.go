package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/cdp/types"
)

// BaseDigitFactor is 10**18, used during coin calculations
const BaseDigitFactor = 1000000000000000000

// AddCdp adds a cdp for a specific owner and collateral type
func (k Keeper) AddCdp(ctx sdk.Context, owner sdk.AccAddress, collateral sdk.Coin, principal sdk.Coin) error {
	// validation
	err := k.ValidateCollateral(ctx, collateral)
	if err != nil {
		return err
	}
	_, found := k.GetCdpByOwnerAndDenom(ctx, owner, collateral.Denom)
	if found {
		return sdkerrors.Wrapf(types.ErrCdpAlreadyExists, "owner %s, denom %s", owner, collateral.Denom)
	}
	err = k.ValidatePrincipalAdd(ctx, principal)
	if err != nil {
		return err
	}

	err = k.ValidateDebtLimit(ctx, collateral.Denom, principal)
	if err != nil {
		return err
	}
	err = k.ValidateCollateralizationRatio(ctx, collateral, principal, sdk.NewCoin(principal.Denom, sdk.ZeroInt()))
	if err != nil {
		return err
	}

	// send coins from the owners account to the cdp module
	id := k.GetNextCdpID(ctx)
	cdp := types.NewCDP(id, owner, collateral, principal, ctx.BlockHeader().Time)
	deposit := types.NewDeposit(cdp.ID, owner, collateral)
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, sdk.NewCoins(collateral))
	if err != nil {
		return err
	}

	// mint the principal and send to the owners account
	err = k.supplyKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(principal))
	if err != nil {
		panic(err)
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, sdk.NewCoins(principal))
	if err != nil {
		panic(err)
	}

	// mint the corresponding amount of debt coins
	err = k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), principal)
	if err != nil {
		panic(err)
	}

	// emit events for cdp creation, deposit, and draw
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateCdp,
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, collateral.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpDraw,
			sdk.NewAttribute(sdk.AttributeKeyAmount, principal.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)

	// update total principal for input collateral type
	k.IncrementTotalPrincipal(ctx, collateral.Denom, principal)

	// set the cdp, deposit, and indexes in the store
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, collateral, principal)
	k.SetCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio)
	k.IndexCdpByOwner(ctx, cdp)
	k.SetDeposit(ctx, deposit)
	k.SetNextCdpID(ctx, id+1)
	return nil
}

// SetCdpAndCollateralRatioIndex sets the cdp and collateral ratio index in the store
func (k Keeper) SetCdpAndCollateralRatioIndex(ctx sdk.Context, cdp types.CDP, ratio sdk.Dec) {
	k.SetCDP(ctx, cdp)
	k.IndexCdpByCollateralRatio(ctx, cdp.Collateral.Denom, cdp.ID, ratio)
}

// MintDebtCoins mints debt coins in the cdp module account
func (k Keeper) MintDebtCoins(ctx sdk.Context, moduleAccount string, denom string, principalCoins sdk.Coin) error {
	debtCoins := sdk.NewCoins(sdk.NewCoin(denom, principalCoins.Amount))
	err := k.supplyKeeper.MintCoins(ctx, moduleAccount, debtCoins)
	if err != nil {
		return err
	}
	return nil
}

// BurnDebtCoins burns debt coins from the cdp module account
func (k Keeper) BurnDebtCoins(ctx sdk.Context, moduleAccount string, denom string, paymentCoins sdk.Coin) error {
	debtCoins := sdk.NewCoins(sdk.NewCoin(denom, paymentCoins.Amount))
	err := k.supplyKeeper.BurnCoins(ctx, moduleAccount, debtCoins)
	if err != nil {
		return err
	}
	return nil
}

// GetCdpID returns the id of the cdp corresponding to a specific owner and collateral denom
func (k Keeper) GetCdpID(ctx sdk.Context, owner sdk.AccAddress, denom string) (uint64, bool) {

	cdpIDs, found := k.GetCdpIdsByOwner(ctx, owner)
	if !found {
		return 0, false
	}
	for _, id := range cdpIDs {
		_, found = k.GetCDP(ctx, denom, id)
		if found {
			return id, true
		}
	}
	return 0, false

}

// GetCdpIdsByOwner returns all the ids of cdps corresponding to a particular owner
func (k Keeper) GetCdpIdsByOwner(ctx sdk.Context, owner sdk.AccAddress) ([]uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIDKeyPrefix)
	bz := store.Get(owner)
	// TODO figure out why this is necessary
	if bz == nil || bytes.Equal(bz, []byte{0}) {
		return []uint64{}, false
	}
	var cdpIDs []uint64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cdpIDs)
	return cdpIDs, true
}

// GetCdpByOwnerAndDenom queries cdps owned by owner and returns the cdp with matching denom
func (k Keeper) GetCdpByOwnerAndDenom(ctx sdk.Context, owner sdk.AccAddress, denom string) (types.CDP, bool) {
	cdpIDs, found := k.GetCdpIdsByOwner(ctx, owner)
	if !found {
		return types.CDP{}, false
	}
	for _, id := range cdpIDs {
		cdp, found := k.GetCDP(ctx, denom, id)
		if found {
			return cdp, true
		}
	}
	return types.CDP{}, false
}

// GetCDP returns the cdp associated with a particular collateral denom and id
func (k Keeper) GetCDP(ctx sdk.Context, collateralDenom string, cdpID uint64) (types.CDP, bool) {
	// get store
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db, found := k.GetDenomPrefix(ctx, collateralDenom)
	if !found {
		return types.CDP{}, false
	}
	// get CDP
	bz := store.Get(types.CdpKey(db, cdpID))
	// unmarshal
	if bz == nil {
		return types.CDP{}, false
	}
	var cdp types.CDP
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cdp)
	return cdp, true
}

// SetCDP sets a cdp in the store
func (k Keeper) SetCDP(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db, found := k.GetDenomPrefix(ctx, cdp.Collateral.Denom)
	if !found {
		panic(fmt.Sprintf("invalid collateral denom %s", cdp.Collateral.Denom))
	}
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cdp)
	store.Set(types.CdpKey(db, cdp.ID), bz)
	return
}

// DeleteCDP deletes a cdp from the store
func (k Keeper) DeleteCDP(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db, found := k.GetDenomPrefix(ctx, cdp.Collateral.Denom)
	if !found {
		panic(fmt.Sprintf("invalid collateral denom %s", cdp.Collateral.Denom))
	}
	store.Delete(types.CdpKey(db, cdp.ID))

}

// GetAllCdps returns all cdps from the store
func (k Keeper) GetAllCdps(ctx sdk.Context) (cdps types.CDPs) {
	k.IterateAllCdps(ctx, func(cdp types.CDP) bool {
		cdps = append(cdps, cdp)
		return false
	})
	return
}

// GetAllCdpsByDenom returns all cdps of a particular collateral type from the store
func (k Keeper) GetAllCdpsByDenom(ctx sdk.Context, denom string) (cdps types.CDPs) {
	k.IterateCdpsByDenom(ctx, denom, func(cdp types.CDP) bool {
		cdps = append(cdps, cdp)
		return false
	})
	return
}

// GetAllCdpsByDenomAndRatio returns all cdps of a particular collateral type and below a certain collateralization ratio
func (k Keeper) GetAllCdpsByDenomAndRatio(ctx sdk.Context, denom string, targetRatio sdk.Dec) (cdps types.CDPs) {
	k.IterateCdpsByCollateralRatio(ctx, denom, targetRatio, func(cdp types.CDP) bool {
		cdps = append(cdps, cdp)
		return false
	})
	return
}

// SetNextCdpID sets the highest cdp id in the store
func (k Keeper) SetNextCdpID(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIDKey)
	store.Set([]byte{}, types.GetCdpIDBytes(id))
}

// GetNextCdpID returns the highest cdp id from the store
func (k Keeper) GetNextCdpID(ctx sdk.Context) (id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIDKey)
	bz := store.Get([]byte{})
	if bz == nil {
		panic("starting cdp id not set in genesis")
	}
	id = types.GetCdpIDFromBytes(bz)
	return
}

// IndexCdpByOwner sets the cdp id in the store, indexed by the owner
func (k Keeper) IndexCdpByOwner(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIDKeyPrefix)
	cdpIDs, found := k.GetCdpIdsByOwner(ctx, cdp.Owner)

	if !found {
		idBytes := k.cdc.MustMarshalBinaryLengthPrefixed([]uint64{cdp.ID})
		store.Set(cdp.Owner, idBytes)
		return
	}
	for _, id := range cdpIDs {
		if id == cdp.ID {
			return
		}
		cdpIDs = append(cdpIDs, cdp.ID)
		store.Set(cdp.Owner, k.cdc.MustMarshalBinaryLengthPrefixed(cdpIDs))
	}
}

// RemoveCdpOwnerIndex deletes the cdp id from the store's index of cdps by owner
func (k Keeper) RemoveCdpOwnerIndex(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIDKeyPrefix)
	cdpIDs, found := k.GetCdpIdsByOwner(ctx, cdp.Owner)
	if !found {
		return
	}
	updatedCdpIds := []uint64{}
	for _, id := range cdpIDs {
		if id != cdp.ID {
			updatedCdpIds = append(updatedCdpIds, id)
		}
	}
	if len(updatedCdpIds) == 0 {
		store.Delete(cdp.Owner)
	}
	store.Set(cdp.Owner, k.cdc.MustMarshalBinaryLengthPrefixed(updatedCdpIds))

}

// IndexCdpByCollateralRatio sets the cdp id in the store, indexed by the collateral type and collateral to debt ratio
func (k Keeper) IndexCdpByCollateralRatio(ctx sdk.Context, denom string, id uint64, collateralRatio sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	db, _ := k.GetDenomPrefix(ctx, denom)
	store.Set(types.CollateralRatioKey(db, id, collateralRatio), types.GetCdpIDBytes(id))
}

// RemoveCdpCollateralRatioIndex deletes the cdp id from the store's index of cdps by collateral type and collateral to debt ratio
func (k Keeper) RemoveCdpCollateralRatioIndex(ctx sdk.Context, denom string, id uint64, collateralRatio sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	db, _ := k.GetDenomPrefix(ctx, denom)
	store.Delete(types.CollateralRatioKey(db, id, collateralRatio))
}

// GetDebtDenom returns the denom of debt in the system
func (k Keeper) GetDebtDenom(ctx sdk.Context) (denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DebtDenomKey)
	bz := store.Get([]byte{})
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &denom)
	return
}

// GetGovDenom returns the denom of debt in the system
func (k Keeper) GetGovDenom(ctx sdk.Context) (denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GovDenomKey)
	bz := store.Get([]byte{})
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &denom)
	return
}

// SetDebtDenom set the denom of debt in the system
func (k Keeper) SetDebtDenom(ctx sdk.Context, denom string) {
	if denom == "" {
		panic("debt denom not set in genesis")
	}
	store := prefix.NewStore(ctx.KVStore(k.key), types.DebtDenomKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryLengthPrefixed(denom))
	return
}

// SetGovDenom set the denom of the governance token in the system
func (k Keeper) SetGovDenom(ctx sdk.Context, denom string) {
	if denom == "" {
		panic("gov denom not set in genesis")
	}
	store := prefix.NewStore(ctx.KVStore(k.key), types.GovDenomKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryLengthPrefixed(denom))
	return
}

// ValidateCollateral validates that a collateral is valid for use in cdps
func (k Keeper) ValidateCollateral(ctx sdk.Context, collateral sdk.Coin) error {
	_, found := k.GetCollateral(ctx, collateral.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrCollateralNotSupported, collateral.Denom)
	}
	return nil
}

// ValidatePrincipalAdd validates that an asset is valid for use as debt when creating a new cdp
func (k Keeper) ValidatePrincipalAdd(ctx sdk.Context, principal sdk.Coin) error {
	dp, found := k.GetDebtParam(ctx, principal.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrDebtNotSupported, principal.Denom)
	}
	if principal.Amount.LT(dp.DebtFloor) {
		return sdkerrors.Wrapf(types.ErrBelowDebtFloor, "proposed %s < minimum %s", principal, dp.DebtFloor)
	}
	return nil
}

// ValidatePrincipalDraw validates that an asset is valid for use as debt when drawing debt off an existing cdp
func (k Keeper) ValidatePrincipalDraw(ctx sdk.Context, principal sdk.Coin, expectedDenom string) error {
	if principal.Denom != expectedDenom {
		return sdkerrors.Wrapf(types.ErrInvalidDebtRequest, "proposed %s, expected %s", principal.Denom, expectedDenom)
	}
	_, found := k.GetDebtParam(ctx, principal.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrDebtNotSupported, principal.Denom)
	}
	return nil
}

// ValidateDebtLimit validates that the input debt amount does not exceed the global debt limit or the debt limit for that collateral
func (k Keeper) ValidateDebtLimit(ctx sdk.Context, collateralDenom string, principal sdk.Coin) error {
	cp, found := k.GetCollateral(ctx, collateralDenom)
	if !found {
		return sdkerrors.Wrap(types.ErrCollateralNotSupported, collateralDenom)
	}
	totalPrincipal := k.GetTotalPrincipal(ctx, collateralDenom, principal.Denom).Add(principal.Amount)
	collateralLimit := cp.DebtLimit.Amount
	if totalPrincipal.GT(collateralLimit) {
		return sdkerrors.Wrapf(types.ErrExceedsDebtLimit, "debt increase %s > collateral debt limit %s", sdk.NewCoins(sdk.NewCoin(principal.Denom, totalPrincipal)), sdk.NewCoins(sdk.NewCoin(principal.Denom, collateralLimit)))
	}
	globalLimit := k.GetParams(ctx).GlobalDebtLimit.Amount
	if totalPrincipal.GT(globalLimit) {
		return sdkerrors.Wrapf(types.ErrExceedsDebtLimit, "debt increase %s > global debt limit  %s", sdk.NewCoin(principal.Denom, totalPrincipal), sdk.NewCoin(principal.Denom, globalLimit))
	}
	return nil
}

// ValidateCollateralizationRatio validate that adding the input principal doesn't put the cdp below the liquidation ratio
func (k Keeper) ValidateCollateralizationRatio(ctx sdk.Context, collateral sdk.Coin, principal sdk.Coin, fees sdk.Coin) error {
	//
	collateralizationRatio, err := k.CalculateCollateralizationRatio(ctx, collateral, principal, fees)
	if err != nil {
		return err
	}
	liquidationRatio := k.getLiquidationRatio(ctx, collateral.Denom)
	if collateralizationRatio.LT(liquidationRatio) {
		return sdkerrors.Wrapf(types.ErrInvalidCollateralRatio, "collateral %s, collateral ratio %s, liquidation ratio %s", collateral.Denom, collateralizationRatio, liquidationRatio)
	}
	return nil
}

// CalculateCollateralToDebtRatio returns the collateral to debt ratio of the input collateral and debt amounts
func (k Keeper) CalculateCollateralToDebtRatio(ctx sdk.Context, collateral sdk.Coin, debt sdk.Coin) sdk.Dec {
	debtTotal := k.convertDebtToBaseUnits(ctx, debt)

	if debtTotal.IsZero() || debtTotal.GTE(types.MaxSortableDec) {
		return types.MaxSortableDec.Sub(sdk.SmallestDec())
	}

	collateralBaseUnits := k.convertCollateralToBaseUnits(ctx, collateral)
	return collateralBaseUnits.Quo(debtTotal)
}

// LoadAugmentedCDP creates a new augmented CDP from an existing CDP
func (k Keeper) LoadAugmentedCDP(ctx sdk.Context, cdp types.CDP) (types.AugmentedCDP, error) {
	// calculate collateralization ratio
	collateralizationRatio, err := k.CalculateCollateralizationRatio(ctx, cdp.Collateral, cdp.Principal, cdp.AccumulatedFees)
	if err != nil {
		return types.AugmentedCDP{}, err
	}

	// total debt is the sum of all oustanding principal and fees
	var totalDebt int64
	totalDebt += cdp.Principal.Amount.Int64()
	totalDebt += cdp.AccumulatedFees.Amount.Int64()

	// convert collateral value to debt coin
	debtBaseAdjusted := sdk.NewDec(totalDebt).QuoInt64(BaseDigitFactor)
	collateralValueInDebtDenom := collateralizationRatio.Mul(debtBaseAdjusted)
	collateralValueInDebt := sdk.NewInt64Coin(cdp.Principal.Denom, collateralValueInDebtDenom.Int64())

	// create new augmuented cdp
	augmentedCDP := types.NewAugmentedCDP(cdp, collateralValueInDebt, collateralizationRatio)
	return augmentedCDP, nil
}

// CalculateCollateralizationRatio returns the collateralization ratio of the input collateral to the input debt plus fees
func (k Keeper) CalculateCollateralizationRatio(ctx sdk.Context, collateral sdk.Coin, principal sdk.Coin, fees sdk.Coin) (sdk.Dec, error) {
	if collateral.IsZero() {
		return sdk.ZeroDec(), nil
	}
	marketID := k.getMarketID(ctx, collateral.Denom)
	price, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
	if err != nil {
		return sdk.Dec{}, err
	}
	collateralBaseUnits := k.convertCollateralToBaseUnits(ctx, collateral)
	collateralValue := collateralBaseUnits.Mul(price.Price)

	prinicpalBaseUnits := k.convertDebtToBaseUnits(ctx, principal)
	principalTotal := prinicpalBaseUnits
	feeBaseUnits := k.convertDebtToBaseUnits(ctx, fees)
	principalTotal = principalTotal.Add(feeBaseUnits)

	collateralRatio := collateralValue.Quo(principalTotal)
	return collateralRatio, nil
}

// CalculateCollateralizationRatioFromAbsoluteRatio takes a coin's denom and an absolute ratio and returns the respective collateralization ratio
func (k Keeper) CalculateCollateralizationRatioFromAbsoluteRatio(ctx sdk.Context, collateralDenom string, absoluteRatio sdk.Dec) (sdk.Dec, error) {
	// get price collateral
	marketID := k.getMarketID(ctx, collateralDenom)
	price, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
	if err != nil {
		return sdk.Dec{}, err
	}
	// convert absolute ratio to collateralization ratio
	respectiveCollateralRatio := absoluteRatio.Quo(price.Price)
	return respectiveCollateralRatio, nil
}

// converts the input collateral to base units (ie multiplies the input by 10^(-ConversionFactor))
func (k Keeper) convertCollateralToBaseUnits(ctx sdk.Context, collateral sdk.Coin) (baseUnits sdk.Dec) {
	cp, _ := k.GetCollateral(ctx, collateral.Denom)
	return sdk.NewDecFromInt(collateral.Amount).Mul(sdk.NewDecFromIntWithPrec(sdk.OneInt(), cp.ConversionFactor.Int64()))
}

// converts the input debt to base units (ie multiplies the input by 10^(-ConversionFactor))
func (k Keeper) convertDebtToBaseUnits(ctx sdk.Context, debt sdk.Coin) (baseUnits sdk.Dec) {
	dp, _ := k.GetDebtParam(ctx, debt.Denom)
	return sdk.NewDecFromInt(debt.Amount).Mul(sdk.NewDecFromIntWithPrec(sdk.OneInt(), dp.ConversionFactor.Int64()))
}
