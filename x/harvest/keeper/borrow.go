package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/harvest/types"
)

// Borrow funds
func (k Keeper) Borrow(ctx sdk.Context, borrower sdk.AccAddress, coins sdk.Coins) error {
	// TODO: Here we assume borrower only has one coin. To be addressed in future card.
	err := k.ValidateBorrow(ctx, borrower, coins[0])
	if err != nil {
		return err
	}

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, borrower, coins)
	if err != nil {
		return err
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
func (k Keeper) ValidateBorrow(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {
	proprosedBorrowUSDValue, err := k.calculateUSDValue(ctx, amount.Amount, amount.Denom)
	if err != nil {
		return err
	}

	// Get the total value of the user's deposits
	deposits := k.GetDepositsByUser(ctx, borrower)
	if len(deposits) == 0 {
		return sdkerrors.Wrapf(types.ErrDepositsNotFound, "no deposits found for %s", borrower)
	}
	totalBorrowableAmount := sdk.ZeroDec()
	for _, deposit := range deposits {
		borrowableAmountForDeposit, err := k.getBorrowableAmountForDeposit(ctx, deposit)
		if err != nil {
			return err
		}
		totalBorrowableAmount = totalBorrowableAmount.Add(borrowableAmountForDeposit)
	}

	previousBorrowsUSDValue := sdk.ZeroDec()
	previousBorrows, found := k.GetBorrow(ctx, borrower)
	if found {
		// TODO: here we're assuming that the user only has 1 previous borrow. To be addressed in future cards.
		previousBorrow := previousBorrows.Amount[0]
		previousBorrowUSDValue, err := k.calculateUSDValue(ctx, previousBorrow.Amount, previousBorrow.Denom)
		if err != nil {
			return err
		}
		previousBorrowsUSDValue = previousBorrowsUSDValue.Add(previousBorrowUSDValue)
	}

	// Validate that the proposed borrow's USD value is within user's borrowable limit
	if proprosedBorrowUSDValue.GT(totalBorrowableAmount.Sub(previousBorrowsUSDValue)) {
		return sdkerrors.Wrapf(types.ErrInsufficientLoanToValue, "requested borrow %s is greater than maximum valid borrow", amount)
	}
	return nil
}

func (k Keeper) calculateUSDValue(ctx sdk.Context, amount sdk.Int, denom string) (sdk.Dec, error) {
	moneyMarket, found := k.GetMoneyMarket(ctx, denom)
	if !found {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", denom)
	}
	assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
	}
	return sdk.NewDecFromInt(amount).Quo(sdk.NewDecFromInt(moneyMarket.ConversionFactor)).Mul(assetPriceInfo.Price), nil
}

func (k Keeper) getBorrowableAmountForDeposit(ctx sdk.Context, deposit types.Deposit) (sdk.Dec, error) {
	moneyMarket, found := k.GetMoneyMarket(ctx, deposit.Amount.Denom)
	if !found {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", deposit.Amount.Denom)
	}
	assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
	}
	usdValue := sdk.NewDecFromInt(deposit.Amount.Amount).Quo(sdk.NewDecFromInt(moneyMarket.ConversionFactor)).Mul(assetPriceInfo.Price)
	return usdValue.Mul(moneyMarket.BorrowLimit.LoanToValue), nil
}
