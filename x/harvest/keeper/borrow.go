package keeper

import (
	"math"

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

	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, borrower, types.ModuleAccountName, sdk.NewCoins(amount))
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

func (k Keeper) getAssetPrice(ctx sdk.Context, denom string) (sdk.Dec, error) {
	moneyMarket, found := k.GetMoneyMarket(ctx, denom)
	if !found {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", denom)
	}
	assetPriceInfo, err := k.pricefeedKeeper.GetCurrentPrice(ctx, moneyMarket.SpotMarketID)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrPriceNotFound, "no price found for market %s", moneyMarket.SpotMarketID)
	}
	return assetPriceInfo.Price, nil
}

func (k Keeper) validateBorrowUser(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {
	var borrowUSDValue sdk.Dec
	if amount.Denom == USDX {
		borrowUSDValue = sdk.NewDecFromInt(amount.Amount).Quo(sdk.NewDec(int64(math.Pow(10, 6)))) // TODO: include conversion_factor in params
	} else {
		price, err := k.getAssetPrice(ctx, amount.Denom)
		if err != nil {
			return err
		}
		borrowUSDValue = sdk.NewDecFromInt(amount.Amount).Mul(price)
	}

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
			price, err := k.getAssetPrice(ctx, deposit.Amount.Denom)
			if err != nil {
				return err
			}
			depositUSDValue := sdk.NewDecFromInt(deposit.Amount.Amount).Mul(price)
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
			price, err := k.getAssetPrice(ctx, borrow.Amount.Denom)
			if err != nil {
				return err
			}
			borrowUSDValue := sdk.NewDecFromInt(borrow.Amount.Amount).Mul(price)
			totalValueBorrows = totalValueBorrows.Add(borrowUSDValue)
		}
	}

	// Value of borrow cannot be greater than:
	// (total value of user's deposits * the borrow asset denom's LTV ratio) - funds already borrowed
	moneyMarket, found := k.GetMoneyMarket(ctx, amount.Denom)
	if !found { // Sanity check
		sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", amount.Denom)
	}
	borrowValueLimit := totalValueDeposits.Mul(moneyMarket.BorrowLimit.LoanToValue).Sub(totalValueBorrows)
	if borrowUSDValue.GT(borrowValueLimit) {
		// Here we get the price so that we can return a helpful error to the user
		price, err := k.getAssetPrice(ctx, amount.Denom)
		if err != nil {
			return err
		}
		return sdkerrors.Wrapf(types.ErrInsufficientLoanToValue,
			"requested borrow %s is greater than maximum valid borrow %s",
			amount, sdk.NewCoin(amount.Denom, borrowValueLimit.Quo(price).TruncateInt()))
	}

	return nil
}

func (k Keeper) validateBorrowSystem(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {
	// TODO: validate borrow against system requirements
	return nil
}
