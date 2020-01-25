package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

const BaseDigitFactor = 1000000000000000000

// AddCdp adds a cdp for a specific owner and collateral type
func (k Keeper) AddCdp(ctx sdk.Context, owner sdk.AccAddress, collateral sdk.Coins, principal sdk.Coins) sdk.Error {
	// validation
	err := k.ValidateCollateral(ctx, collateral)
	if err != nil {
		return err
	}
	_, found := k.GetCdpByOwnerAndDenom(ctx, owner, collateral[0].Denom)
	if found {
		return types.ErrCdpAlreadyExists(k.codespace, owner, collateral[0].Denom)
	}
	err = k.ValidatePrincipalAdd(ctx, principal)
	if err != nil {
		return err
	}
	err = k.ValidateCollateralizationRatio(ctx, collateral, principal, sdk.NewCoins())
	if err != nil {
		return err
	}

	// send coins from the owners account to the cdp module
	id := k.GetNextCdpID(ctx)
	cdp := types.NewCDP(id, owner, collateral, principal, ctx.BlockHeader().Time)
	deposit := types.NewDeposit(cdp.ID, owner, collateral)
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, collateral)
	if err != nil {
		return err
	}

	// mint the principal and send to the owners account
	err = k.supplyKeeper.MintCoins(ctx, types.ModuleName, principal)
	if err != nil {
		panic(err)
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, principal)
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
	k.IncrementTotalPrincipal(ctx, collateral[0].Denom, principal)

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
	k.IndexCdpByCollateralRatio(ctx, cdp.Collateral[0].Denom, cdp.ID, ratio)
}

// MintDebtCoins mints debt coins in the cdp module account
func (k Keeper) MintDebtCoins(ctx sdk.Context, moduleAccount string, denom string, principalCoins sdk.Coins) sdk.Error {
	coinsToMint := sdk.NewCoins()
	for _, sc := range principalCoins {
		coinsToMint = coinsToMint.Add(sdk.NewCoins(sdk.NewCoin(denom, sc.Amount)))
	}
	err := k.supplyKeeper.MintCoins(ctx, moduleAccount, coinsToMint)
	if err != nil {
		return err
	}
	return nil
}

// BurnDebtCoins burns debts coins from the cdp module account
func (k Keeper) BurnDebtCoins(ctx sdk.Context, moduleAccount string, denom string, paymentCoins sdk.Coins) sdk.Error {
	coinsToBurn := sdk.NewCoins()
	for _, pc := range paymentCoins {
		coinsToBurn = coinsToBurn.Add(sdk.NewCoins(sdk.NewCoin(denom, pc.Amount)))
	}
	err := k.supplyKeeper.BurnCoins(ctx, moduleAccount, coinsToBurn)
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
	db, _ := k.GetDenomPrefix(ctx, collateralDenom)
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
	db, _ := k.GetDenomPrefix(ctx, cdp.Collateral[0].Denom)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cdp)
	store.Set(types.CdpKey(db, cdp.ID), bz)
	return
}

// DeleteCDP deletes a cdp from the store
func (k Keeper) DeleteCDP(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db, _ := k.GetDenomPrefix(ctx, cdp.Collateral[0].Denom)
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
func (k Keeper) ValidateCollateral(ctx sdk.Context, collateral sdk.Coins) sdk.Error {
	if len(collateral) != 1 {
		return types.ErrInvalidCollateralLength(k.codespace, len(collateral))
	}
	_, found := k.GetCollateral(ctx, collateral[0].Denom)
	if !found {
		return types.ErrCollateralNotSupported(k.codespace, collateral[0].Denom)
	}
	return nil
}

// ValidatePrincipalAdd validates that an asset is valid for use as debt when creating a new cdp
func (k Keeper) ValidatePrincipalAdd(ctx sdk.Context, principal sdk.Coins) sdk.Error {
	for _, dc := range principal {
		dp, found := k.GetDebtParam(ctx, dc.Denom)
		if !found {
			return types.ErrDebtNotSupported(k.codespace, dc.Denom)
		}
		if dc.Amount.LT(dp.DebtFloor) {
			return types.ErrBelowDebtFloor(k.codespace, sdk.NewCoins(dc), dp.DebtFloor)
		}
	}
	return nil
}

// ValidatePrincipalDraw validates that an asset is valid for use as debt when drawing debt off an existing cdp
func (k Keeper) ValidatePrincipalDraw(ctx sdk.Context, principal sdk.Coins) sdk.Error {
	for _, dc := range principal {
		_, found := k.GetDebtParam(ctx, dc.Denom)
		if !found {
			return types.ErrDebtNotSupported(k.codespace, dc.Denom)
		}
	}
	return nil
}

// ValidateDebtLimit validates that the input debt amount does not exceed the global debt limit
func (k Keeper) ValidateDebtLimit(ctx sdk.Context, collateralDenom string, principal sdk.Coins) sdk.Error {
	for _, dc := range principal {
		totalPrincipal := k.GetTotalPrincipal(ctx, collateralDenom, dc.Denom).Add(dc.Amount)
		globalLimit := k.GetParams(ctx).GlobalDebtLimit.AmountOf(dc.Denom)
		if totalPrincipal.GT(globalLimit) {
			return types.ErrExceedsDebtLimit(k.codespace, sdk.NewCoins(sdk.NewCoin(dc.Denom, totalPrincipal)), sdk.NewCoins(sdk.NewCoin(dc.Denom, globalLimit)))
		}
	}
	return nil
}

// ValidateCollateralizationRatio validate that adding the input principal doesn't put the cdp below the liquidation ratio
func (k Keeper) ValidateCollateralizationRatio(ctx sdk.Context, collateral sdk.Coins, principal sdk.Coins, fees sdk.Coins) sdk.Error {
	//
	collateralizationRatio, err := k.CalculateCollateralizationRatio(ctx, collateral, principal, fees)
	if err != nil {
		return err
	}
	liquidationRatio := k.getLiquidationRatio(ctx, collateral[0].Denom)
	if collateralizationRatio.LT(liquidationRatio) {
		return types.ErrInvalidCollateralRatio(k.codespace, collateral[0].Denom, collateralizationRatio, liquidationRatio)
	}
	return nil
}

// CalculateCollateralToDebtRatio returns the collateral to debt ratio of the input collateral and debt amounts
func (k Keeper) CalculateCollateralToDebtRatio(ctx sdk.Context, collateral sdk.Coins, debt sdk.Coins) sdk.Dec {
	debtTotal := sdk.ZeroDec()
	for _, dc := range debt {
		debtBaseUnits := k.convertDebtToBaseUnits(ctx, dc)
		debtTotal = debtTotal.Add(debtBaseUnits)
	}

	if debtTotal.IsZero() || debtTotal.GTE(types.MaxSortableDec) {
		return types.MaxSortableDec.Sub(sdk.SmallestDec())
	}

	collateralBaseUnits := k.convertCollateralToBaseUnits(ctx, collateral[0])
	return collateralBaseUnits.Quo(debtTotal)
}

// LoadAugmentedCDP creates a new augmented CDP from an existing CDP
func (k Keeper) LoadAugmentedCDP(ctx sdk.Context, cdp types.CDP) (types.AugmentedCDP, sdk.Error) {
	// calculate collateralization ratio
	collateralizationRatio, err := k.CalculateCollateralizationRatio(ctx, cdp.Collateral, cdp.Principal, cdp.AccumulatedFees)
	if err != nil {
		return types.AugmentedCDP{}, err
	}
	// calcylate collateral value in debt coin
	var totalDebt int64
	if len(cdp.AccumulatedFees) > 0 {
		totalDebt += cdp.AccumulatedFees[0].Amount.Int64()
	}
	totalDebt += cdp.Principal[0].Amount.Int64()
	debtBaseAdjusted := sdk.NewDec(totalDebt).QuoInt64(BaseDigitFactor)
	collateralValueInDebtDenom := collateralizationRatio.Mul(debtBaseAdjusted)
	collateralValueInDebt := sdk.NewInt64Coin(cdp.Principal[0].Denom, collateralValueInDebtDenom.Int64())

	// create new augmuented cdp
	augmentedCDP := types.NewAugmentedCDP(cdp, collateralValueInDebt, collateralizationRatio)
	return augmentedCDP, nil
}

// CalculateCollateralizationRatio returns the collateralization ratio of the input collateral to the input debt plus fees
func (k Keeper) CalculateCollateralizationRatio(ctx sdk.Context, collateral sdk.Coins, principal sdk.Coins, fees sdk.Coins) (sdk.Dec, sdk.Error) {
	if collateral.IsZero() {
		return sdk.ZeroDec(), nil
	}
	marketID := k.getMarketID(ctx, collateral[0].Denom)
	price, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
	if err != nil {
		return sdk.Dec{}, err
	}
	collateralBaseUnits := k.convertCollateralToBaseUnits(ctx, collateral[0])
	collateralValue := collateralBaseUnits.Mul(price.Price)

	principalTotal := sdk.ZeroDec()
	for _, pc := range principal {
		prinicpalBaseUnits := k.convertDebtToBaseUnits(ctx, pc)
		principalTotal = principalTotal.Add(prinicpalBaseUnits)
	}
	for _, fc := range fees {
		feeBaseUnits := k.convertDebtToBaseUnits(ctx, fc)
		principalTotal = principalTotal.Add(feeBaseUnits)
	}
	collateralRatio := collateralValue.Quo(principalTotal)
	return collateralRatio, nil
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
