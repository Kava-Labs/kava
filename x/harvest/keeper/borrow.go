package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/harvest/types"
)

// USDX is the USDX coin's denom
const USDX = "usdx"

// Borrow funds
func (k Keeper) Borrow(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {

	err := k.ValidateBorrow(ctx, borrower, amount)
	if err != nil {
		return err
	}

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, borrower, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	borrow, found := k.GetBorrow(ctx, borrower, amount.Denom)
	if !found {
		borrow = types.NewBorrow(borrower, amount)
	} else {
		borrow.Amount = borrow.Amount.Add(amount)
	}

	k.SetBorrow(ctx, borrow)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestBorrow,
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyBorrower, borrow.Borrower.String()),
			sdk.NewAttribute(types.AttributeKeyBorrowDenom, borrow.Amount.Denom),
		),
	)

	return nil
}

// ValidateBorrow validates a borrow request against borrower and protocol requirements
func (k Keeper) ValidateBorrow(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {
	if err := k.validateBorrowUser(ctx, borrower, amount); err != nil {
		return err
	}

	if err := k.validateBorrowSystem(ctx, borrower, amount); err != nil {
		return err
	}

	return nil
}

func (k Keeper) validateBorrowUser(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {
	// Get requested borrow amount's value
	borrowMarketID, found := k.pricefeedKeeper.GetLiveMarketIDByDenom(ctx, amount.Denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for %s", amount.Denom)
	}
	borrowAssetPrice, err := k.pricefeedKeeper.GetCurrentPrice(ctx, borrowMarketID)
	if err != nil {
		return err
	}
	borrowUSDValue := sdk.NewDecFromInt(amount.Amount).Mul(borrowAssetPrice.Price)

	// Get the user's deposits
	deposits := k.GetDepositsByUser(ctx, borrower)
	if len(deposits) == 0 {
		return sdkerrors.Wrapf(types.ErrDepositsNotFound, "no deposits found for %s", borrower)
	}

	// Get the total value of the user's deposits
	totalValueDeposits := sdk.ZeroDec()
	for _, deposit := range deposits {
		if deposit.Amount.Denom == USDX {
			totalValueDeposits = totalValueDeposits.Add(sdk.NewDecFromInt(deposit.Amount.Amount))
		} else {
			marketID, found := k.pricefeedKeeper.GetLiveMarketIDByDenom(ctx, deposit.Amount.Denom)
			if !found {
				return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for %s", deposit.Amount.Denom)
			}
			depositAssetPrice, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
			if err != nil {
				return err
			}
			depositUSDValue := sdk.NewDecFromInt(deposit.Amount.Amount).Mul(depositAssetPrice.Price)
			totalValueDeposits = totalValueDeposits.Add(depositUSDValue)
		}
	}

	// Get the total value of the user's borrows
	borrows := k.GetBorrowsByUser(ctx, borrower)
	totalValueBorrows := sdk.ZeroDec()
	for _, borrow := range borrows {
		if borrow.Amount.Denom == USDX {
			totalValueBorrows = totalValueBorrows.Add(sdk.NewDecFromInt(borrow.Amount.Amount))
		} else {
			marketID, found := k.pricefeedKeeper.GetLiveMarketIDByDenom(ctx, borrow.Amount.Denom)
			if !found {
				return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for %s", borrow.Amount.Denom)
			}
			borrowAssetPrice, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
			if err != nil {
				return err
			}
			borrowUSDValue := sdk.NewDecFromInt(borrow.Amount.Amount).Mul(borrowAssetPrice.Price)
			totalValueBorrows = totalValueBorrows.Add(borrowUSDValue)
		}
	}

	// Get the borrow asset's LTV param
	moneyMarket, found := k.GetMoneyMarket(ctx, amount.Denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrMoneyMarketNotFound, "no money market found for %s", amount.Denom)
	}

	// Value of borrow cannot be greater than:
	// (total value of user's deposits * the borrow asset denom's LTV ratio) - funds already borrowed
	borrowValueLimit := totalValueDeposits.Mul(moneyMarket.BorrowLimit.LoanToValue).Sub(totalValueBorrows)
	if borrowUSDValue.GT(borrowValueLimit) {
		return sdkerrors.Wrapf(types.ErrInsufficientLoanToValue,
			"requested borrow %s is greater than maximum valid borrow %s",
			amount, sdk.NewCoin(amount.Denom, borrowValueLimit.Quo(borrowAssetPrice.Price).TruncateInt()))
	}

	return nil
}

func (k Keeper) validateBorrowSystem(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {
	// TODO: validate borrow against system requirements
	return nil
}
