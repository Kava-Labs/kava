package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

	// Only burn extended coin if the amount is positive
	if extendedAmount.IsPositive() {
		if err := k.burnExtendedCoin(ctx, moduleName, extendedAmount); err != nil {
			return err
		}
	}

	fullEmissionCoins := sdk.NewCoins(types.SumExtendedCoin(amt))
	if fullEmissionCoins.IsZero() {
		return nil
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		banktypes.NewCoinBurnEvent(acc.GetAddress(), fullEmissionCoins),
		banktypes.NewCoinSpentEvent(acc.GetAddress(), fullEmissionCoins),
	})

	return nil
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

	// If true, fractional balance is insufficient and will require an integer
	// borrow.
	requiresBorrow := newFractionalBalance.IsNegative()

	// Add to new remainder with burned fractional amount.
	newRemainder := prevRemainder.Add(fractionalBurnAmount)

	// If true, remainder has accumulated enough fractional amounts to burn 1
	// integer coin.
	overflowingRemainder := newRemainder.GTE(types.ConversionFactor())

	// -------------------------------------------------------------------------
	// Stateful operations for burn

	// Not enough fractional balance:
	// 1. If the new remainder incurs an additional reserve burn, we can just
	//    burn an additional integer coin from the account directly instead as
	//    an optimization.
	// 2. If the new remainder is still under conversion factor (no extra
	//    reserve burn) then we need to transfer 1 integer coin to the reserve
	//    for the integer borrow.

	// Case #1: (optimization) direct burn instead of borrow (reserve transfer)
	// & reserve burn. No additional reserve burn would be necessary after this.
	if requiresBorrow && overflowingRemainder {
		newFractionalBalance = newFractionalBalance.Add(types.ConversionFactor())
		newRemainder = newRemainder.Sub(types.ConversionFactor())

		integerBurnAmount = integerBurnAmount.AddRaw(1)
	}

	// Case #2: Transfer 1 integer coin to reserve for integer borrow to ensure
	// reserve fully backs the fractional amount.
	if requiresBorrow && !overflowingRemainder {
		newFractionalBalance = newFractionalBalance.Add(types.ConversionFactor())

		// Transfer 1 integer coin to reserve to cover the borrowed fractional
		// amount. SendCoinsFromModuleToModule will return an error if the
		// module account has insufficient funds and an error with the full
		// extended balance will be returned.
		borrowCoin := sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt())
		if err := k.bk.SendCoinsFromModuleToModule(
			ctx,
			moduleName,
			types.ModuleName, // borrowed integer is transferred to reserve
			sdk.NewCoins(borrowCoin),
		); err != nil {
			return k.updateInsufficientFundsError(ctx, moduleAddr, amt, err)
		}
	}

	// Case #3: Does not require borrow, but remainder has accumulated enough
	// fractional amounts to burn 1 integer coin.
	if !requiresBorrow && overflowingRemainder {
		reserveBurnCoins := sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt()))
		if err := k.bk.BurnCoins(ctx, types.ModuleName, reserveBurnCoins); err != nil {
			return fmt.Errorf("failed to burn %s for reserve: %w", reserveBurnCoins, err)
		}

		newRemainder = newRemainder.Sub(types.ConversionFactor())
	}

	// Case #4: No additional work required, no borrow needed and no additional
	// reserve burn

	// Burn the integer amount - this may include the extra optimization burn
	// from case #1
	if !integerBurnAmount.IsZero() {
		coin := sdk.NewCoin(types.IntegerCoinDenom, integerBurnAmount)
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(coin)); err != nil {
			return k.updateInsufficientFundsError(ctx, moduleAddr, amt, err)
		}
	}

	// Assign new fractional balance in x/precisebank
	k.SetFractionalBalance(ctx, moduleAddr, newFractionalBalance)

	// Update remainder for burned fractional coins
	k.SetRemainderAmount(ctx, newRemainder)

	return nil
}
