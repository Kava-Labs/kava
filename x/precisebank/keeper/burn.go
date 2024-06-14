package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// BurnCoins burns coins deletes coins from the balance of the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Custom protection for x/precisebank, no external module should be able to
	// affect reserves.
	if moduleName == types.ModuleName {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s cannot be burned from", moduleName))
	}

	// Panic errors are identical to x/bank for consistency.
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Burner) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName))
	}

	// Ensure the coins are valid before burning
	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	// Get non-ExtendedCoinDenom coins
	passthroughCoins := amt

	extendedAmount := amt.AmountOf(types.ExtendedCoinDenom)
	if extendedAmount.IsPositive() {
		// Remove ExtendedCoinDenom from the coins as it is managed by x/precisebank
		removeCoin := sdk.NewCoin(types.ExtendedCoinDenom, extendedAmount)
		passthroughCoins = amt.Sub(removeCoin)
	}

	// Coins unmanaged by x/precisebank are passed through to x/bank
	if !passthroughCoins.Empty() {
		if err := k.bk.BurnCoins(ctx, moduleName, passthroughCoins); err != nil {
			return err
		}
	}

	// No more processing required if no ExtendedCoinDenom
	if extendedAmount.IsZero() {
		return nil
	}

	return k.burnExtendedCoin(ctx, moduleName, extendedAmount)
}

// burnExtendedCoin burns the fractional amount of the ExtendedCoinDenom from the module account.
func (k Keeper) burnExtendedCoin(
	ctx sdk.Context,
	moduleName string,
	amt sdkmath.Int,
) error {
	// Get the module address
	moduleAddr := k.ak.GetModuleAddress(moduleName)

	// We only need the fractional balance to burn coins, as integer burns will
	// return errors on insufficient funds.
	prevFractionalBalance := k.GetFractionalBalance(ctx, moduleAddr)

	// Get remainder amount first to optimize direct burn.
	prevRemainder := k.GetRemainderAmount(ctx)

	// -------------------------------------------------------------------------
	// Pure stateless calculations

	integerBurnAmount := amt.Quo(types.ConversionFactor())
	fractionalBurnAmount := amt.Mod(types.ConversionFactor())

	// newFractionalBalance can be negative if fractional balance is insufficient.
	newFractionalBalance := prevFractionalBalance.Sub(fractionalBurnAmount)

	// Add to new remainder with burned fractional amount.
	newRemainder := prevRemainder.Add(fractionalBurnAmount)

	// -------------------------------------------------------------------------
	// Stateful operations for burn

	// Not enough fractional balance:
	// 1. If the new remainder incurs an additional reserve burn, we can just
	//    burn an additional integer coin from the account directly instead as
	//    an optimization.
	// 2. If the new remainder is still under conversion factor (no extra
	//    reserve burn) then we need to transfer 1 integer coin to the reserve
	//    for the integer borrow.

	// Case #1 - optimization path, direct burn instead of reserve transfer & reserve burn
	if newFractionalBalance.IsNegative() && newRemainder.GTE(types.ConversionFactor()) {
		integerBurnAmount = integerBurnAmount.AddRaw(1)
	}

	// Case #2 - transfer 1 integer coin to reserve for integer borrow.
	if newFractionalBalance.IsNegative() && newRemainder.LT(types.ConversionFactor()) {
		// Transfer 1 integer coin to reserve to cover the borrowed fractional
		// amount. SendCoinsFromModuleToModule will return an error if the
		// module account has insufficient funds and an error with the full
		// extended balance will be returned.
		borrowCoin := sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt())
		if err := k.bk.SendCoinsFromModuleToModule(
			ctx,
			moduleName,
			types.ModuleName,
			sdk.NewCoins(borrowCoin),
		); err != nil {
			return k.updateInsufficientFundsError(ctx, moduleAddr, amt, err)
		}
	}

	// Adjustment for negative fractional balance after either integer borrow
	// case above.
	if newFractionalBalance.IsNegative() {
		// Add the borrowed amount to negative fractional balance.
		// This will always be 0 < newFractionalBalance < ConversionFactor
		// as we are adding ConversionFactor to a negative amount.
		newFractionalBalance = newFractionalBalance.Add(types.ConversionFactor())
	}

	// Burn the integer amount
	if !integerBurnAmount.IsZero() {
		coin := sdk.NewCoin(types.IntegerCoinDenom, integerBurnAmount)
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(coin)); err != nil {
			return k.updateInsufficientFundsError(ctx, moduleAddr, amt, err)
		}
	}

	// Assign new fractional balance in x/precisebank
	k.SetFractionalBalance(ctx, moduleAddr, newFractionalBalance)

	// ----------------------------------------
	// Update remainder & reserves for burned fractional coins

	// If remainder is greater than or equal to the conversion factor, burn
	// additional integer coin to make reserve just enough to back fractional
	// amounts and nothing more.
	// Optimization: Only burn when BOTH:
	// - remainder > conversion factor
	// - fractional balance is not incur an integer borrow
	if newRemainder.GTE(types.ConversionFactor()) && prevFractionalBalance.Sub(fractionalBurnAmount).GTE(sdkmath.ZeroInt()) {
		reserveBurnCoins := sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt()))
		if err := k.bk.BurnCoins(ctx, types.ModuleName, reserveBurnCoins); err != nil {
			return fmt.Errorf("failed to burn %s for reserve: %w", reserveBurnCoins, err)
		}
	}

	if newRemainder.GTE(types.ConversionFactor()) {
		// Update remainder with leftover fractional amount.
		// newRemainder > ConversionFactor, and we need to subtract the burned
		// 1 integer coin amount. This is equivalent to .Mod() in this case.
		newRemainder = newRemainder.Sub(types.ConversionFactor())
	}

	k.SetRemainderAmount(ctx, newRemainder)

	return nil
}

// TODO: Duplicated method from SendCoins - deduplicate when both merged
// updateInsufficientFundsError returns a modified ErrInsufficientFunds with
// extended coin amounts if the error is due to insufficient funds. Otherwise,
// it returns the original error. This is used since x/bank transfers will
// return errors with integer coins, but we want the more accurate error that
// contains the full extended coin balance and send amounts.
func (k Keeper) updateInsufficientFundsError(
	ctx sdk.Context,
	addr sdk.AccAddress,
	amt sdkmath.Int,
	err error,
) error {
	if !errors.Is(err, sdkerrors.ErrInsufficientFunds) {
		return err
	}

	// Check balance is sufficient
	bal := k.GetBalance(ctx, addr, types.ExtendedCoinDenom)
	coin := sdk.NewCoin(types.ExtendedCoinDenom, amt)

	// TODO: This checks spendable coins and returns error with spendable
	// coins, not full balance. If GetBalance() is modified to return the
	// full, including locked, balance then this should be updated to deduct
	// locked coins.

	// Use sdk.NewCoins() so that it removes empty balances - ie. prints
	// empty string if balance is 0. This is to match x/bank behavior.
	spendable := sdk.NewCoins(bal)

	return errorsmod.Wrapf(
		sdkerrors.ErrInsufficientFunds,
		"spendable balance %s is smaller than %s",
		spendable, coin,
	)
}
