package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// DepositCollateral adds collateral to a cdp
func (k Keeper) DepositCollateral(ctx sdk.Context, owner sdk.AccAddress, depositor sdk.AccAddress, collateral sdk.Coins) sdk.Error {
	err := k.ValidateCollateral(ctx, collateral)
	if err != nil {
		return err
	}
	cdp, found := k.GetCdpByOwnerAndDenom(ctx, owner, collateral[0].Denom)
	if !found {
		return types.ErrCdpNotFound(k.codespace, owner, collateral[0].Denom)
	}
	deposit, found := k.GetDeposit(ctx, cdp.ID, depositor)
	if found {
		// TODO should deposits be allowed when the current deposit is InLiquidation?
		deposit.Amount = deposit.Amount.Add(collateral)
		k.SetDeposit(ctx, deposit, cdp.ID)
	} else {
		deposit = types.NewDeposit(cdp.ID, depositor, collateral)
		k.SetDeposit(ctx, deposit, cdp.ID)
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, collateral.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)
	cdp.AccumulatedFees = k.CalculateFees(ctx, cdp)
	cdp.FeesUpdated = ctx.BlockTime()
	cdp.Collateral = cdp.Collateral.Add(collateral)
	k.SetCDP(ctx, cdp)
	collateralRatio := k.CalculateCollateralToDebtRatio(ctx, collateral, cdp.Principal)
	k.IndexCdpByCollateralRatio(ctx, cdp, collateralRatio)
	return nil
}

// WithdrawCollateral removes collateral from a cdp if it does not put the cdp below the liquidation ratio
func (k Keeper) WithdrawCollateral(ctx sdk.Context, owner sdk.AccAddress, depositor sdk.AccAddress, collateral sdk.Coins) sdk.Error {
	err := k.ValidateCollateral(ctx, collateral)
	if err != nil {
		return err
	}
	cdp, found := k.GetCdpByOwnerAndDenom(ctx, owner, collateral[0].Denom)
	if !found {
		return types.ErrCdpNotFound(k.codespace, owner, collateral[0].Denom)
	}
	deposit, found := k.GetDeposit(ctx, cdp.ID, depositor)
	if !found {
		return types.ErrDepositNotFound(k.codespace, depositor, cdp.ID)
	}
	if deposit.InLiquidation {
		return types.ErrDepositNotAvailable(k.codespace, cdp.ID, depositor)
	}
	cdp.AccumulatedFees = cdp.AccumulatedFees.Add(k.CalculateFees(ctx, cdp))
	cdp.FeesUpdated = ctx.BlockTime()
	collateralRatio, err := k.CalculateCollateralizationRatio(ctx, cdp.Collateral.Sub(collateral), cdp.Principal, cdp.AccumulatedFees)
	if err != nil {
		return err
	}
	liquidationRatio := k.getLiquidationRatio(ctx, collateral[0].Denom)
	if collateralRatio.LT(liquidationRatio) {
		return types.ErrInvalidCollateralRatio(k.codespace, collateral[0].Denom, collateralRatio, liquidationRatio)
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpWithdrawal,
			sdk.NewAttribute(sdk.AttributeKeyAmount, collateral.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)
	deposit.Amount = deposit.Amount.Sub(collateral)
	cdp.Collateral = cdp.Collateral.Sub(collateral)
	k.SetCDP(ctx, cdp)
	k.SetDeposit(ctx, deposit, cdp.ID)
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal)
	k.IndexCdpByCollateralRatio(ctx, cdp, collateralToDebtRatio)
	return nil
}

// GetDeposit returns the deposit of a depositor on a particular cdp from the store
func (k Keeper) GetDeposit(ctx sdk.Context, cdpID uint64, depositor sdk.AccAddress) (deposit types.Deposit, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositKeyPrefix)
	bz := store.Get(types.DepositKey(cdpID, depositor))
	if bz == nil {
		return deposit, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
	return deposit, true

}

// SetDeposit sets the deposit in the store
func (k Keeper) SetDeposit(ctx sdk.Context, deposit types.Deposit, cdpID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(deposit)
	store.Set(types.DepositKey(cdpID, deposit.Depositor), bz)
}

// DeleteDeposit deletes a deposit from the store
func (k Keeper) DeleteDeposit(ctx sdk.Context, cdpID uint64, depositor sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositKeyPrefix)
	store.Delete(types.DepositKey(cdpID, depositor))
}

// IterateDeposits iterates over the all the deposits of a cdp and performs a callback function
func (k Keeper) IterateDeposits(ctx sdk.Context, cdpID uint64, cb func(deposit types.Deposit) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, types.GetCdpIDBytes(cdpID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

// GetDeposits returns all the deposits from a proposal
func (k Keeper) GetDeposits(ctx sdk.Context, cdpID uint64) (deposits types.Deposits) {
	k.IterateDeposits(ctx, cdpID, func(deposit types.Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return
}
