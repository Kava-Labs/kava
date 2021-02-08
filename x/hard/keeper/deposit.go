package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/kava-labs/kava/x/hard/types"
)

// Deposit deposit
func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error {
	// Set any new denoms' global supply index to 1.0
	for _, coin := range coins {
		_, foundInterestFactor := k.GetSupplyInterestFactor(ctx, coin.Denom)
		if !foundInterestFactor {
			_, foundMm := k.GetMoneyMarket(ctx, coin.Denom)
			if foundMm {
				k.SetSupplyInterestFactor(ctx, coin.Denom, sdk.OneDec())
			}
		}
	}

	// Call incentive hook
	existingDeposit, hasExistingDeposit := k.GetDeposit(ctx, depositor)
	if hasExistingDeposit {
		k.BeforeDepositModified(ctx, existingDeposit)
	}

	// Sync any outstanding interest
	k.SyncBorrowInterest(ctx, depositor)
	k.SyncSupplyInterest(ctx, depositor)

	err := k.ValidateDeposit(ctx, coins)
	if err != nil {
		return err
	}

	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, coins)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient account funds") {
			accCoins := k.accountKeeper.GetAccount(ctx, depositor).SpendableCoins(ctx.BlockTime())
			for _, coin := range coins {
				_, isNegative := accCoins.SafeSub(sdk.NewCoins(coin))
				if isNegative {
					return sdkerrors.Wrapf(types.ErrBorrowExceedsAvailableBalance,
						"insufficient funds: the requested deposit amount of %s exceeds the total available account funds of %s%s",
						coin, accCoins.AmountOf(coin.Denom), coin.Denom,
					)
				}
			}
		}
	}
	if err != nil {
		return err
	}

	interestFactors := types.SupplyInterestFactors{}
	currDeposit, foundDeposit := k.GetDeposit(ctx, depositor)
	if foundDeposit {
		interestFactors = currDeposit.Index
	}
	for _, coin := range coins {
		interestFactorValue, foundValue := k.GetSupplyInterestFactor(ctx, coin.Denom)
		if foundValue {
			interestFactors = interestFactors.SetInterestFactor(coin.Denom, interestFactorValue)
		}
	}

	// Calculate new deposit amount
	var amount sdk.Coins
	if foundDeposit {
		amount = currDeposit.Amount.Add(coins...)
	} else {
		amount = coins
	}
	// Update the depositer's amount and supply interest factors in the store
	deposit := types.NewDeposit(depositor, amount, interestFactors)

	if deposit.Amount.Empty() {
		k.DeleteDeposit(ctx, deposit)
	} else {
		k.SetDeposit(ctx, deposit)
	}

	k.IncrementSuppliedCoins(ctx, coins)
	if !foundDeposit { // User's first deposit
		k.AfterDepositCreated(ctx, deposit)
	} else {
		k.AfterDepositModified(ctx, deposit)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, coins.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, deposit.Depositor.String()),
		),
	)

	return nil
}

// ValidateDeposit validates a deposit
func (k Keeper) ValidateDeposit(ctx sdk.Context, coins sdk.Coins) error {
	for _, depCoin := range coins {
		_, foundMm := k.GetMoneyMarket(ctx, depCoin.Denom)
		if !foundMm {
			return sdkerrors.Wrapf(types.ErrInvalidDepositDenom, "money market denom %s not found", depCoin.Denom)
		}
	}

	return nil
}

// GetTotalDeposited returns the total amount deposited for the input deposit type and deposit denom
func (k Keeper) GetTotalDeposited(ctx sdk.Context, depositDenom string) (total sdk.Int) {
	var macc supplyExported.ModuleAccountI
	macc = k.supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	return macc.GetCoins().AmountOf(depositDenom)
}

// IncrementSuppliedCoins increments the total amount of supplied coins by the newCoins parameter
func (k Keeper) IncrementSuppliedCoins(ctx sdk.Context, newCoins sdk.Coins) {
	suppliedCoins, found := k.GetSuppliedCoins(ctx)
	if !found {
		if !newCoins.Empty() {
			k.SetSuppliedCoins(ctx, newCoins)
		}
	} else {
		k.SetSuppliedCoins(ctx, suppliedCoins.Add(newCoins...))
	}
}

// DecrementSuppliedCoins decrements the total amount of supplied coins by the coins parameter
func (k Keeper) DecrementSuppliedCoins(ctx sdk.Context, coins sdk.Coins) error {
	suppliedCoins, found := k.GetSuppliedCoins(ctx)
	if !found {
		return sdkerrors.Wrapf(types.ErrSuppliedCoinsNotFound, "cannot withdraw if no coins are deposited")
	}

	updatedSuppliedCoins, isAnyNegative := suppliedCoins.SafeSub(coins)
	if isAnyNegative {
		return types.ErrNegativeSuppliedCoins
	}

	k.SetSuppliedCoins(ctx, updatedSuppliedCoins)
	return nil
}

// GetSyncedDeposit returns a deposit object containing current balances and indexes
func (k Keeper) GetSyncedDeposit(ctx sdk.Context, depositor sdk.AccAddress) (types.Deposit, bool) {
	deposit, found := k.GetDeposit(ctx, depositor)
	if !found {
		return types.Deposit{}, false
	}

	return k.loadSyncedDeposit(ctx, deposit), true
}

// loadSyncedDeposit calculates a user's synced deposit, but does not update state
func (k Keeper) loadSyncedDeposit(ctx sdk.Context, deposit types.Deposit) types.Deposit {
	totalNewInterest := sdk.Coins{}
	newSupplyIndexes := types.SupplyInterestFactors{}
	for _, coin := range deposit.Amount {
		interestFactorValue, foundInterestFactorValue := k.GetSupplyInterestFactor(ctx, coin.Denom)
		if foundInterestFactorValue {
			// Locate the interest factor by coin denom in the user's list of interest factors
			foundAtIndex := -1
			for i := range deposit.Index {
				if deposit.Index[i].Denom == coin.Denom {
					foundAtIndex = i
					break
				}
			}

			// Calculate interest that will be paid to user for this asset
			if foundAtIndex != -1 {
				storedAmount := sdk.NewDecFromInt(deposit.Amount.AmountOf(coin.Denom))
				userLastInterestFactor := deposit.Index[foundAtIndex].Value
				coinInterest := (storedAmount.Quo(userLastInterestFactor).Mul(interestFactorValue)).Sub(storedAmount)
				totalNewInterest = totalNewInterest.Add(sdk.NewCoin(coin.Denom, coinInterest.TruncateInt()))
			}
		}

		supplyIndex := types.NewSupplyInterestFactor(coin.Denom, interestFactorValue)
		newSupplyIndexes = append(newSupplyIndexes, supplyIndex)
	}

	return types.NewDeposit(deposit.Depositor, deposit.Amount.Add(totalNewInterest...), newSupplyIndexes)
}
