package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// AddCdp adds a cdp for a specific owner and collateral type
func (k Keeper) AddCdp(ctx sdk.Context, owner sdk.AccAddress, collateral sdk.Coins, principal sdk.Coins) sdk.Error {
	err := k.ValidateCollateral(ctx, collateral)
	if err != nil {
		return err
	}
	_, found := k.GetCdpID(ctx, owner, collateral[0].Denom)
	if found {
		return types.ErrCdpAlreadyExists(k.codespace, owner, collateral[0].Denom)
	}
	err = k.ValidatePrincipal(ctx, principal)
	if err != nil {
		return err
	}
	collateralRatio, err := k.CalculateCollateralizationRatio(ctx, collateral, principal, sdk.NewCoins())
	if err != nil {
		return err
	}
	liquidationRatio := k.getLiquidationRatio(ctx, collateral[0].Denom)
	if collateralRatio.LT(liquidationRatio) {
		return types.ErrInvalidCollateralRatio(k.codespace, collateral[0].Denom, collateralRatio, liquidationRatio)
	}
	id := k.GetNextCdpID(ctx)
	cdp := types.NewCDP(id, owner, collateral, principal, ctx.BlockHeader().Time)
	deposit := types.NewDeposit(cdp.ID, owner, collateral)
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, collateral)
	if err != nil {
		return err
	}
	err = k.supplyKeeper.MintCoins(ctx, types.ModuleName, principal)
	if err != nil {
		return err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, principal)
	if err != nil {
		return err
	}

	err = k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), principal)
	if err != nil {
		return err
	}

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
	k.IncrementTotalPrincipal(ctx, collateral[0].Denom, principal)
	k.SetCDP(ctx, cdp)
	k.SetDeposit(ctx, deposit, id)
	k.SetNextCdpID(ctx, id+1)
	k.IndexCdpByOwner(ctx, cdp)
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, collateral, principal)
	k.IndexCdpByCollateralRatio(ctx, cdp, collateralToDebtRatio)

	return nil
}

// MintDebtCoins mints debt coins in the cdp module account
func (k Keeper) MintDebtCoins(ctx sdk.Context, moduleAccount string, denom string, principalCoins sdk.Coins) sdk.Error {
	var coinsToMint sdk.Coins
	for _, sc := range principalCoins {
		coinsToMint.Add(sdk.NewCoins(sdk.NewCoin(denom, sc.Amount)))
	}
	err := k.supplyKeeper.MintCoins(ctx, moduleAccount, coinsToMint)
	if err != nil {
		return err
	}
	return nil
}

// BurnDebtCoins burns debts coins from the cdp module account
func (k Keeper) BurnDebtCoins(ctx sdk.Context, moduleAccount string, denom string, paymentCoins sdk.Coins) sdk.Error {
	var coinsToBurn sdk.Coins
	for _, pc := range paymentCoins {
		coinsToBurn.Add(sdk.NewCoins(sdk.NewCoin(denom, pc.Amount)))
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
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIdKeyPrefix)
	bz := store.Get(owner)
	if bz == nil {
		return []uint64{}, false
	}
	var cdpIDs []uint64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cdpIDs)
	return cdpIDs, true
}

// GetCDP returns the cdp associated with a particular collateral denom and id
func (k Keeper) GetCDP(ctx sdk.Context, collateralDenom string, cdpID uint64) (types.CDP, bool) {
	// get store
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db := k.getDenomPrefix(ctx, collateralDenom)
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
	db := k.getDenomPrefix(ctx, cdp.Collateral[0].Denom)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cdp)
	store.Set(types.CdpKey(db, cdp.ID), bz)
	return
}

// DeleteCDP deletes a cdp from the store
func (k Keeper) DeleteCDP(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db := k.getDenomPrefix(ctx, cdp.Collateral[0].Denom)
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

// SetNextCdpID sets the highest cdp id in the store
func (k Keeper) SetNextCdpID(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	store.Set([]byte{}, types.GetCdpIDBytes(id))
}

// GetNextCdpID returns the highest cdp id from the store
func (k Keeper) GetNextCdpID(ctx sdk.Context) (id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIdKey)
	bz := store.Get([]byte{})
	if bz == nil {
		panic("starting cdp id not set in genesis")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &id)
	return
}

// IndexCdpByOwner sets the cdp id in the store, indexed by the owner
func (k Keeper) IndexCdpByOwner(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIdKeyPrefix)
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
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpIdKeyPrefix)
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
func (k Keeper) IndexCdpByCollateralRatio(ctx sdk.Context, cdp types.CDP, collateralRatio sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	db := k.getDenomPrefix(ctx, cdp.Collateral[0].Denom)
	store.Set(types.LiquidationRatioKey(db, cdp.ID, collateralRatio), types.GetCdpIDBytes(cdp.ID))
}

// RemoveCdpLiquidationRatioIndex deletes the cdp id from the store's index of cdps by collateral type and collateral to debt ratio
func (k Keeper) RemoveCdpLiquidationRatioIndex(ctx sdk.Context, cdp types.CDP) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	db := k.getDenomPrefix(ctx, cdp.Collateral[0].Denom)
	iterKey := append([]byte{db}, []byte(":")...)
	iterKey = append(iterKey, cdp.Owner...)
	iterator := sdk.KVStorePrefixIterator(store, iterKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

// GetDebtDenom returns the denom of debt in the system
func (k Keeper) GetDebtDenom(ctx sdk.Context) (denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DebtDenomKey)
	bz := store.Get([]byte{})
	if bz == nil {
		panic("debt denom not set in genesis")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &denom)
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

// ValidatePrincipal validates that an asset is valid for use as debt in cdps
func (k Keeper) ValidatePrincipal(ctx sdk.Context, principal sdk.Coins) sdk.Error {
	for _, dc := range principal {
		dp, found := k.GetDebt(ctx, dc.Denom)
		if !found {
			return types.ErrDebtNotSupported(k.codespace, dc.Denom)
		}
		if dp.DebtLimit.Add(sdk.NewCoins(dc)).IsAnyGT(dp.DebtLimit) {
			return types.ErrExceedsDebtLimit(k.codespace, sdk.NewCoins(dc), dp.DebtLimit)
		}
	}
	return nil
}

// CalculateCollateralToDebtRatio returns the collateral to debt ratio of the input collateral and debt amounts
func (k Keeper) CalculateCollateralToDebtRatio(ctx sdk.Context, collateral sdk.Coins, debt sdk.Coins) sdk.Dec {
	var debtTotal sdk.Int
	for _, dc := range debt {
		debtTotal.Add(dc.Amount)
	}
	return sdk.NewDecFromInt(collateral[0].Amount).Quo(sdk.NewDecFromInt(debtTotal))
}

// CalculateCollateralizationRatio returns the collateralization ratio of the input collateral to the input debt plus fees
func (k Keeper) CalculateCollateralizationRatio(ctx sdk.Context, collateral sdk.Coins, principal sdk.Coins, fees sdk.Coins) (sdk.Dec, sdk.Error) {
	marketID := k.getMarketID(ctx, collateral[0].Denom)
	price, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
	if err != nil {
		return sdk.Dec{}, err
	}

	collateralValue := sdk.NewDecFromInt(collateral[0].Amount).Mul(price.Price)
	var principalTotal sdk.Int
	for _, pc := range principal {
		principalTotal.Add(pc.Amount)
	}
	for _, fc := range fees {
		principalTotal.Add(fc.Amount)
	}
	collateralRatio := collateralValue.Quo(sdk.NewDecFromInt(principalTotal))
	return collateralRatio, nil
}
