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

	// Get current stored LTV based on stored borrows/deposits
	prevLtv, err := k.GetStoreLTV(ctx, depositor)
	if err != nil {
		return err
	}

	// Sync any outstanding interest
	k.SyncBorrowInterest(ctx, depositor)
	k.SyncSupplyInterest(ctx, depositor)

	err = k.ValidateDeposit(ctx, coins)
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

	// The first time a user deposits a denom we add it the user's supply interest factor index
	var supplyInterestFactors types.SupplyInterestFactors
	currDeposit, foundDeposit := k.GetDeposit(ctx, depositor)
	// On user's first deposit, build deposit index list containing denoms and current global deposit index value
	if foundDeposit {
		// If the coin denom to be deposited is not in the user's existing deposit, we add it deposit index
		for _, coin := range coins {
			if !sdk.NewCoins(coin).DenomsSubsetOf(currDeposit.Amount) {
				supplyInterestFactorValue, _ := k.GetSupplyInterestFactor(ctx, coin.Denom)
				supplyInterestFactor := types.NewSupplyInterestFactor(coin.Denom, supplyInterestFactorValue)
				supplyInterestFactors = append(supplyInterestFactors, supplyInterestFactor)
			}
		}
		// Concatenate new deposit interest factors to existing deposit interest factors
		supplyInterestFactors = append(supplyInterestFactors, currDeposit.Index...)
	} else {
		for _, coin := range coins {
			supplyInterestFactorValue, _ := k.GetSupplyInterestFactor(ctx, coin.Denom)
			supplyInterestFactor := types.NewSupplyInterestFactor(coin.Denom, supplyInterestFactorValue)
			supplyInterestFactors = append(supplyInterestFactors, supplyInterestFactor)
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
	deposit := types.NewDeposit(depositor, amount, supplyInterestFactors)

	// Calculate the new Loan-to-Value ratio of Deposit-to-Borrow
	borrow, _ := k.GetBorrow(ctx, depositor)
	newLtv, err := k.CalculateLtv(ctx, deposit, borrow)
	if err != nil {
		return err
	}

	k.UpdateDepositAndLtvIndex(ctx, deposit, newLtv, prevLtv)

	// Update total supplied amount by newly supplied coins. Don't add user's pending interest as
	// it has already been included in the total supplied coins by the BeginBlocker.
	k.IncrementSuppliedCoins(ctx, coins)

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
	params := k.GetParams(ctx)
	for _, depCoin := range coins {
		found := false
		for _, lps := range params.LiquidityProviderSchedules {
			if lps.DepositDenom == depCoin.Denom {
				found = true
			}
		}
		if !found {
			return sdkerrors.Wrapf(types.ErrInvalidDepositDenom, "liquidity provider denom %s not found", depCoin.Denom)
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
