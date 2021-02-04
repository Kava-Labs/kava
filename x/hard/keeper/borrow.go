package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/types"
)

// Borrow funds
func (k Keeper) Borrow(ctx sdk.Context, borrower sdk.AccAddress, coins sdk.Coins) error {
	// Set any new denoms' global borrow index to 1.0
	for _, coin := range coins {
		_, foundInterestFactor := k.GetBorrowInterestFactor(ctx, coin.Denom)
		if !foundInterestFactor {
			_, foundMm := k.GetMoneyMarket(ctx, coin.Denom)
			if foundMm {
				k.SetBorrowInterestFactor(ctx, coin.Denom, sdk.OneDec())
			}
		}
	}

	// Get current stored LTV based on stored borrows/deposits
	prevLtv, err := k.GetStoreLTV(ctx, borrower)
	if err != nil {
		return err
	}

	// Call incentive hook
	existingBorrow, hasExistingBorrow := k.GetBorrow(ctx, borrower)
	if hasExistingBorrow {
		k.BeforeBorrowModified(ctx, existingBorrow)
	}

	k.SyncBorrowInterest(ctx, borrower)

	// Validate borrow amount within user and protocol limits
	err = k.ValidateBorrow(ctx, borrower, coins)
	if err != nil {
		return err
	}

	// Sends coins from Hard module account to user
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, borrower, coins)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient account funds") {
			modAccCoins := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName).GetCoins()
			for _, coin := range coins {
				_, isNegative := modAccCoins.SafeSub(sdk.NewCoins(coin))
				if isNegative {
					return sdkerrors.Wrapf(types.ErrBorrowExceedsAvailableBalance,
						"the requested borrow amount of %s exceeds the total amount of %s%s available to borrow",
						coin, modAccCoins.AmountOf(coin.Denom), coin.Denom,
					)
				}
			}
		}
	}

	// The first time a user borrows a denom we add it the user's borrow interest factor index
	var borrowInterestFactors types.BorrowInterestFactors
	currBorrow, foundBorrow := k.GetBorrow(ctx, borrower)
	// On user's first borrow, build borrow index list containing denoms and current global borrow index value
	if foundBorrow {
		// If the coin denom to be borrowed is not in the user's existing borrow, we add it borrow index
		for _, coin := range coins {
			if !sdk.NewCoins(coin).DenomsSubsetOf(currBorrow.Amount) {
				borrowInterestFactorValue, _ := k.GetBorrowInterestFactor(ctx, coin.Denom)
				borrowInterestFactor := types.NewBorrowInterestFactor(coin.Denom, borrowInterestFactorValue)
				borrowInterestFactors = append(borrowInterestFactors, borrowInterestFactor)
			}
		}
		// Concatenate new borrow interest factors to existing borrow interest factors
		borrowInterestFactors = append(borrowInterestFactors, currBorrow.Index...)
	} else {
		for _, coin := range coins {
			borrowInterestFactorValue, _ := k.GetBorrowInterestFactor(ctx, coin.Denom)
			borrowInterestFactor := types.NewBorrowInterestFactor(coin.Denom, borrowInterestFactorValue)
			borrowInterestFactors = append(borrowInterestFactors, borrowInterestFactor)
		}
	}

	// Calculate new borrow amount
	var amount sdk.Coins
	if foundBorrow {
		amount = currBorrow.Amount.Add(coins...)
	} else {
		amount = coins
	}
	// Construct the user's new/updated borrow with amount and interest factors
	borrow := types.NewBorrow(borrower, amount, borrowInterestFactors)

	// Calculate the new Loan-to-Value ratio of Deposit-to-Borrow
	deposit, foundDeposit := k.GetDeposit(ctx, borrower)
	if !foundDeposit {
		return types.ErrDepositNotFound
	}
	newLtv, err := k.CalculateLtv(ctx, deposit, borrow)
	if err != nil {
		return err
	}

	k.UpdateBorrowAndLtvIndex(ctx, borrow, newLtv, prevLtv)

	// Update total borrowed amount by newly borrowed coins. Don't add user's pending interest as
	// it has already been included in the total borrowed coins by the BeginBlocker.
	k.IncrementBorrowedCoins(ctx, coins)

	if !hasExistingBorrow {
		k.AfterBorrowCreated(ctx, borrow)
	} else {
		k.AfterBorrowModified(ctx, borrow)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardBorrow,
			sdk.NewAttribute(types.AttributeKeyBorrower, borrower.String()),
			sdk.NewAttribute(types.AttributeKeyBorrowCoins, coins.String()),
		),
	)

	return nil
}

// ValidateBorrow validates a borrow request against borrower and protocol requirements
func (k Keeper) ValidateBorrow(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coins) error {
	if amount.IsZero() {
		return types.ErrBorrowEmptyCoins
	}

	// Get the proposed borrow USD value
	moneyMarketCache := map[string]types.MoneyMarket{}
	proprosedBorrowUSDValue := sdk.ZeroDec()
	for _, coin := range amount {
		moneyMarket, ok := moneyMarketCache[coin.Denom]
		// Fetch money market and store in local cache
		if !ok {
			newMoneyMarket, found := k.GetMoneyMarketParam(ctx, coin.Denom)
			if !found {
				return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", coin.Denom)
			}
			moneyMarketCache[coin.Denom] = newMoneyMarket
			moneyMarket = newMoneyMarket
		}

		// Calculate this coin's USD value and add it borrow's total USD value
		assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
		if err != nil {
			return sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
		}
		coinUSDValue := sdk.NewDecFromInt(coin.Amount).Quo(sdk.NewDecFromInt(moneyMarket.ConversionFactor)).Mul(assetPriceInfo.Price)

		// Validate the requested borrow value for the asset against the money market's global borrow limit
		if moneyMarket.BorrowLimit.HasMaxLimit {
			var assetTotalBorrowedAmount sdk.Int
			totalBorrowedCoins, found := k.GetBorrowedCoins(ctx)
			if !found {
				assetTotalBorrowedAmount = sdk.ZeroInt()
			} else {
				assetTotalBorrowedAmount = totalBorrowedCoins.AmountOf(coin.Denom)
			}
			newProposedAssetTotalBorrowedAmount := sdk.NewDecFromInt(assetTotalBorrowedAmount.Add(coin.Amount))
			if newProposedAssetTotalBorrowedAmount.GT(moneyMarket.BorrowLimit.MaximumLimit) {
				return sdkerrors.Wrapf(types.ErrGreaterThanAssetBorrowLimit,
					"proposed borrow would result in %s borrowed, but the maximum global asset borrow limit is %s",
					newProposedAssetTotalBorrowedAmount, moneyMarket.BorrowLimit.MaximumLimit)
			}
		}
		proprosedBorrowUSDValue = proprosedBorrowUSDValue.Add(coinUSDValue)
	}

	// Get the total borrowable USD amount at user's existing deposits
	deposit, found := k.GetDeposit(ctx, borrower)
	if !found {
		return sdkerrors.Wrapf(types.ErrDepositsNotFound, "no deposits found for %s", borrower)
	}
	totalBorrowableAmount := sdk.ZeroDec()
	for _, depCoin := range deposit.Amount {
		moneyMarket, ok := moneyMarketCache[depCoin.Denom]
		// Fetch money market and store in local cache
		if !ok {
			newMoneyMarket, found := k.GetMoneyMarketParam(ctx, depCoin.Denom)
			if !found {
				return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", depCoin.Denom)
			}
			moneyMarketCache[depCoin.Denom] = newMoneyMarket
			moneyMarket = newMoneyMarket
		}

		// Calculate the borrowable amount and add it to the user's total borrowable amount
		assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
		if err != nil {
			return sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
		}
		depositUSDValue := sdk.NewDecFromInt(depCoin.Amount).Quo(sdk.NewDecFromInt(moneyMarket.ConversionFactor)).Mul(assetPriceInfo.Price)
		borrowableAmountForDeposit := depositUSDValue.Mul(moneyMarket.BorrowLimit.LoanToValue)
		totalBorrowableAmount = totalBorrowableAmount.Add(borrowableAmountForDeposit)
	}

	// Get the total USD value of user's existing borrows
	existingBorrowUSDValue := sdk.ZeroDec()
	existingBorrow, found := k.GetBorrow(ctx, borrower)
	if found {
		for _, borrowedCoin := range existingBorrow.Amount {
			moneyMarket, ok := moneyMarketCache[borrowedCoin.Denom]
			// Fetch money market and store in local cache
			if !ok {
				newMoneyMarket, found := k.GetMoneyMarketParam(ctx, borrowedCoin.Denom)
				if !found {
					return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", borrowedCoin.Denom)
				}
				moneyMarketCache[borrowedCoin.Denom] = newMoneyMarket
				moneyMarket = newMoneyMarket
			}

			// Calculate this borrow coin's USD value and add it to the total previous borrowed USD value
			assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
			if err != nil {
				return sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
			}
			coinUSDValue := sdk.NewDecFromInt(borrowedCoin.Amount).Quo(sdk.NewDecFromInt(moneyMarket.ConversionFactor)).Mul(assetPriceInfo.Price)
			existingBorrowUSDValue = existingBorrowUSDValue.Add(coinUSDValue)
		}
	}

	// Validate that the proposed borrow's USD value is within user's borrowable limit
	if proprosedBorrowUSDValue.GT(totalBorrowableAmount.Sub(existingBorrowUSDValue)) {
		return sdkerrors.Wrapf(types.ErrInsufficientLoanToValue, "requested borrow %s exceeds the allowable amount as determined by the collateralization ratio", amount)
	}
	return nil
}

// IncrementBorrowedCoins increments the total amount of borrowed coins by the newCoins parameter
func (k Keeper) IncrementBorrowedCoins(ctx sdk.Context, newCoins sdk.Coins) {
	borrowedCoins, found := k.GetBorrowedCoins(ctx)
	if !found {
		if !newCoins.Empty() {
			k.SetBorrowedCoins(ctx, newCoins)
		}
	} else {
		k.SetBorrowedCoins(ctx, borrowedCoins.Add(newCoins...))
	}
}

// DecrementBorrowedCoins decrements the total amount of borrowed coins by the coins parameter
func (k Keeper) DecrementBorrowedCoins(ctx sdk.Context, coins sdk.Coins) error {
	borrowedCoins, found := k.GetBorrowedCoins(ctx)
	if !found {
		return sdkerrors.Wrapf(types.ErrBorrowedCoinsNotFound, "cannot repay coins if no coins are currently borrowed")
	}

	updatedBorrowedCoins, isAnyNegative := borrowedCoins.SafeSub(coins)
	if isAnyNegative {
		return types.ErrNegativeBorrowedCoins
	}

	k.SetBorrowedCoins(ctx, updatedBorrowedCoins)
	return nil
}

// GetSyncedBorrow returns a borrow object containing current balances and indexes
func (k Keeper) GetSyncedBorrow(ctx sdk.Context, borrower sdk.AccAddress) (types.Borrow, bool) {
	borrow, found := k.GetBorrow(ctx, borrower)
	if !found {
		return types.Borrow{}, false
	}

	return k.loadSyncedBorrow(ctx, borrow), true
}

// loadSyncedBorrow calculates a user's synced borrow, but does not update state
func (k Keeper) loadSyncedBorrow(ctx sdk.Context, borrow types.Borrow) types.Borrow {
	totalNewInterest := sdk.Coins{}
	newBorrowIndexes := types.BorrowInterestFactors{}
	for _, coin := range borrow.Amount {
		interestFactorValue, foundInterestFactorValue := k.GetBorrowInterestFactor(ctx, coin.Denom)
		if foundInterestFactorValue {
			// Locate the interest factor by coin denom in the user's list of interest factors
			foundAtIndex := -1
			for i := range borrow.Index {
				if borrow.Index[i].Denom == coin.Denom {
					foundAtIndex = i
					break
				}
			}

			// Calculate interest owed by user for this asset
			if foundAtIndex != -1 {
				storedAmount := sdk.NewDecFromInt(borrow.Amount.AmountOf(coin.Denom))
				userLastInterestFactor := borrow.Index[foundAtIndex].Value
				coinInterest := (storedAmount.Quo(userLastInterestFactor).Mul(interestFactorValue)).Sub(storedAmount)
				totalNewInterest = totalNewInterest.Add(sdk.NewCoin(coin.Denom, coinInterest.TruncateInt()))
			}
		}

		borrowIndex := types.NewBorrowInterestFactor(coin.Denom, interestFactorValue)
		newBorrowIndexes = append(newBorrowIndexes, borrowIndex)
	}

	return types.NewBorrow(borrow.Borrower, borrow.Amount.Add(totalNewInterest...), newBorrowIndexes)
}
