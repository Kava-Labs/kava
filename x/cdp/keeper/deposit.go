package keeper

import (
	"context"
	storetypes "cosmossdk.io/store/types"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// DepositCollateral adds collateral to a cdp
func (k Keeper) DepositCollateral(ctx context.Context, owner, depositor sdk.AccAddress, collateral sdk.Coin, collateralType string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// check that collateral exists and has a functioning pricefeed
	err := k.ValidateCollateral(sdkCtx, collateral, collateralType)
	if err != nil {
		return err
	}
	cdp, found := k.GetCdpByOwnerAndCollateralType(ctx, owner, collateralType)
	if !found {
		return errorsmod.Wrapf(types.ErrCdpNotFound, "owner %s, collateral %s", owner, collateralType)
	}
	err = k.ValidateBalance(sdkCtx, collateral, depositor)
	if err != nil {
		return err
	}
	k.hooks.BeforeCDPModified(sdkCtx, cdp)
	cdp = k.SynchronizeInterest(ctx, cdp)

	deposit, found := k.GetDeposit(ctx, cdp.ID, depositor)
	if found {
		deposit.Amount = deposit.Amount.Add(collateral)
	} else {
		deposit = types.NewDeposit(cdp.ID, depositor, collateral)
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, sdk.NewCoins(collateral))
	if err != nil {
		return err
	}

	k.SetDeposit(ctx, deposit)

	cdp.Collateral = cdp.Collateral.Add(collateral)
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, collateral.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)

	return k.UpdateCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio)
}

// WithdrawCollateral removes collateral from a cdp if it does not put the cdp below the liquidation ratio
func (k Keeper) WithdrawCollateral(ctx context.Context, owner, depositor sdk.AccAddress, collateral sdk.Coin, collateralType string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.ValidateCollateral(sdkCtx, collateral, collateralType)
	if err != nil {
		return err
	}
	cdp, found := k.GetCdpByOwnerAndCollateralType(ctx, owner, collateralType)
	if !found {
		return errorsmod.Wrapf(types.ErrCdpNotFound, "owner %s, collateral %s", owner, collateral.Denom)
	}
	deposit, found := k.GetDeposit(ctx, cdp.ID, depositor)
	if !found {
		return errorsmod.Wrapf(types.ErrDepositNotFound, "depositor %s, collateral %s %s", depositor, collateral.Denom, collateralType)
	}
	if collateral.Amount.GT(deposit.Amount.Amount) {
		return errorsmod.Wrapf(types.ErrInvalidWithdrawAmount, "collateral %s, deposit %s", collateral, deposit.Amount)
	}
	k.hooks.BeforeCDPModified(sdkCtx, cdp)
	cdp = k.SynchronizeInterest(ctx, cdp)

	collateralizationRatio, err := k.CalculateCollateralizationRatio(sdkCtx, cdp.Collateral.Sub(collateral), cdp.Type, cdp.Principal, cdp.AccumulatedFees, spot)
	if err != nil {
		return err
	}
	liquidationRatio := k.getLiquidationRatio(ctx, cdp.Type)
	if collateralizationRatio.LT(liquidationRatio) {
		return errorsmod.Wrapf(types.ErrInvalidCollateralRatio, "collateral %s, collateral ratio %s, liquidation ration %s", collateral.Denom, collateralizationRatio, liquidationRatio)
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, sdk.NewCoins(collateral))
	if err != nil {
		panic(err)
	}

	cdp.Collateral = cdp.Collateral.Sub(collateral)
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(sdkCtx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())
	err = k.UpdateCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio)
	if err != nil {
		return err
	}

	deposit.Amount = deposit.Amount.Sub(collateral)
	// delete deposits if amount is 0
	if deposit.Amount.IsZero() {
		k.DeleteDeposit(ctx, deposit.CdpID, deposit.Depositor)
	} else {
		k.SetDeposit(ctx, deposit)
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpWithdrawal,
			sdk.NewAttribute(sdk.AttributeKeyAmount, collateral.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)

	return nil
}

// GetDeposit returns the deposit of a depositor on a particular cdp from the store
func (k Keeper) GetDeposit(ctx context.Context, cdpID uint64, depositor sdk.AccAddress) (deposit types.Deposit, found bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.DepositKeyPrefix)
	bz := store.Get(types.DepositKey(cdpID, depositor))
	if bz == nil {
		return deposit, false
	}
	k.cdc.MustUnmarshal(bz, &deposit)
	return deposit, true
}

// SetDeposit sets the deposit in the store
func (k Keeper) SetDeposit(ctx context.Context, deposit types.Deposit) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.DepositKeyPrefix)
	bz := k.cdc.MustMarshal(&deposit)

	store.Set(types.DepositKey(deposit.CdpID, deposit.Depositor), bz)
}

// DeleteDeposit deletes a deposit from the store
func (k Keeper) DeleteDeposit(ctx context.Context, cdpID uint64, depositor sdk.AccAddress) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.DepositKeyPrefix)
	store.Delete(types.DepositKey(cdpID, depositor))
}

// IterateDeposits iterates over the all the deposits of a cdp and performs a callback function
func (k Keeper) IterateDeposits(ctx context.Context, cdpID uint64, cb func(deposit types.Deposit) (stop bool)) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.DepositKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, types.GetCdpIDBytes(cdpID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

// GetDeposits returns all the deposits to a cdp
func (k Keeper) GetDeposits(ctx context.Context, cdpID uint64) (deposits types.Deposits) {
	k.IterateDeposits(ctx, cdpID, func(deposit types.Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return
}
