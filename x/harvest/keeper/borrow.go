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
	proposedBorrowPrice, proposedBorrowConversionFactor, err := k.getAssetPrice(ctx, amount.Denom)
	if err != nil {
		return err
	}
	proprosedBorrowUSDValue := sdk.NewDecFromInt(amount.Amount).Quo(sdk.NewDecFromInt(proposedBorrowConversionFactor).Mul(proposedBorrowPrice))

	// Get the total value of the user's deposits
	deposits := k.GetDepositsByUser(ctx, borrower)
	if len(deposits) == 0 {
		return sdkerrors.Wrapf(types.ErrDepositsNotFound, "no deposits found for %s", borrower)
	}
	deposit := deposits[0] // TODO: Here we assume there's only one deposit. To be addressed in future cards.
	depositPrice, depositConversionFactor, err := k.getAssetPrice(ctx, deposit.Amount.Denom)
	if err != nil {
		return err
	}
	depositUSDValue := sdk.NewDecFromInt(deposit.Amount.Amount).Quo(sdk.NewDecFromInt(depositConversionFactor).Mul(depositPrice))

	previousBorrowUSDValue := sdk.ZeroDec()
	previousBorrows, found := k.GetBorrow(ctx, borrower)
	if found {
		// TODO: here we're assuming that the user only has 1 previous borrow. To be addressed in future cards.
		previousBorrow := previousBorrows.Amount[0]
		previousBorrowPrice, previousBorrowConversionFactor, err := k.getAssetPrice(ctx, previousBorrow.Denom)
		if err != nil {
			return err
		}
		previousBorrowUSDValue = sdk.NewDecFromInt(previousBorrow.Amount).Quo(sdk.NewDecFromInt(previousBorrowConversionFactor).Mul(previousBorrowPrice))
	}

	// Value of borrow cannot be greater than:
	// (total value of user's deposits * the borrow asset denom's LTV ratio) - funds already borrowed
	moneyMarket, found := k.GetMoneyMarket(ctx, amount.Denom)
	if !found { // Sanity check
		sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", amount.Denom)
	}
	borrowValueLimit := depositUSDValue.Mul(moneyMarket.BorrowLimit.LoanToValue).Sub(previousBorrowUSDValue)
	if proprosedBorrowUSDValue.GT(borrowValueLimit) {
		return sdkerrors.Wrapf(types.ErrInsufficientLoanToValue, "requested borrow %s is greater than maximum valid borrow", amount)
	}
	return nil
}

func (k Keeper) getAssetPrice(ctx sdk.Context, denom string) (sdk.Dec, sdk.Int, error) {
	moneyMarket, found := k.GetMoneyMarket(ctx, denom)
	if !found {
		return sdk.ZeroDec(), sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", denom)
	}
	assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
	if err != nil {
		return sdk.ZeroDec(), moneyMarket.ConversionFactor, sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
	}
	return assetPriceInfo.Price, moneyMarket.ConversionFactor, nil
}
