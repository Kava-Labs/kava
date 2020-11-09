package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/harvest/types"
)

// Borrow funds
func (k Keeper) Borrow(ctx sdk.Context, borrower sdk.AccAddress, coins sdk.Coins) error {
	err := k.ValidateBorrow(ctx, borrower, coins)
	if err != nil {
		return err
	}

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, borrower, coins)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient account funds") {
			return sdkerrors.Wrapf(types.ErrBorrowExceedsAvailableBalance,
				"the requested borrow amount exceeds the total amount of %s available to borrow",
				k.supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName).GetCoins(),
			)
		}
	}

	borrow, found := k.GetBorrow(ctx, borrower)
	if !found {
		borrow = types.NewBorrow(borrower, coins)
	} else {
		borrow.Amount = borrow.Amount.Add(coins...)
	}
	k.SetBorrow(ctx, borrow)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestBorrow,
			sdk.NewAttribute(types.AttributeKeyBorrower, borrow.Borrower.String()),
			sdk.NewAttribute(types.AttributeKeyBorrowCoins, coins.String()),
		),
	)

	return nil
}

// ValidateBorrow validates a borrow request against borrower and protocol requirements
func (k Keeper) ValidateBorrow(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coins) error {
	// Get the proposed borrow USD value
	moneyMarketCache := map[string]types.MoneyMarket{}
	proprosedBorrowUSDValue := sdk.ZeroDec()
	for _, coin := range amount {
		moneyMarket, ok := moneyMarketCache[coin.Denom]
		// Fetch money market and store in local cache
		if !ok {
			newMoneyMarket, found := k.GetMoneyMarket(ctx, coin.Denom)
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
		proprosedBorrowUSDValue = proprosedBorrowUSDValue.Add(coinUSDValue)
	}

	// Get the total borrowable USD amount at user's existing deposits
	deposits := k.GetDepositsByUser(ctx, borrower)
	if len(deposits) == 0 {
		return sdkerrors.Wrapf(types.ErrDepositsNotFound, "no deposits found for %s", borrower)
	}
	totalBorrowableAmount := sdk.ZeroDec()
	for _, deposit := range deposits {
		moneyMarket, ok := moneyMarketCache[deposit.Amount.Denom]
		// Fetch money market and store in local cache
		if !ok {
			newMoneyMarket, found := k.GetMoneyMarket(ctx, deposit.Amount.Denom)
			if !found {
				return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", deposit.Amount.Denom)
			}
			moneyMarketCache[deposit.Amount.Denom] = newMoneyMarket
			moneyMarket = newMoneyMarket
		}

		// Calculate the borrowable amount and add it to the user's total borrowable amount
		assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
		if err != nil {
			sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
		}
		depositUSDValue := sdk.NewDecFromInt(deposit.Amount.Amount).Quo(sdk.NewDecFromInt(moneyMarket.ConversionFactor)).Mul(assetPriceInfo.Price)
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
				newMoneyMarket, found := k.GetMoneyMarket(ctx, borrowedCoin.Denom)
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
		return sdkerrors.Wrapf(types.ErrInsufficientLoanToValue, "requested borrow %s is greater than maximum valid borrow", amount)
	}
	return nil
}
