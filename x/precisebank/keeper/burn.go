package keeper

import (
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

	prevBalance := k.GetBalance(ctx, moduleAddr, types.ExtendedCoinDenom)

	// Check if the balance is sufficient
	if prevBalance.Amount.LT(amt) {
		coin := sdk.NewCoin(types.ExtendedCoinDenom, amt)
		spendable := sdk.NewCoins(prevBalance)

		return errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"spendable balance %s is smaller than %s",
			spendable, coin,
		)
	}

	// Determine burn amounts
	prevFractionalBalance := prevBalance.Amount.Mod(types.ConversionFactor())

	integerBurnAmount := amt.Quo(types.ConversionFactor())
	fractionalBurnAmount := amt.Mod(types.ConversionFactor())

	newFractionalBalance := prevFractionalBalance.Sub(fractionalBurnAmount)

	// Not enough fractional balance, need to "borrow" from the integer balance
	// by adding an additional burn amount
	if fractionalBurnAmount.GT(prevFractionalBalance) {
		// Burn additional integer amount to borrow. In this case the account
		// will always have enough integer balance to cover the additional burn
		// as we checked the full balance above.
		integerBurnAmount = integerBurnAmount.AddRaw(1)

		// Add the borrowed amount to fractional balance, Sub(fractionalBurnAmount)
		// is already done earlier. So this will always be 0 < newFractionalBalance < ConversionFactor
		// Example with 2 fractional digits:
		// prevFBal: .00, Burn: .01
		// =  .00 - .01
		// = -.01 + 1.00
		// = .99
		newFractionalBalance = newFractionalBalance.Add(types.ConversionFactor())
	}

	// Burn the integer amount
	if !integerBurnAmount.IsZero() {
		coin := sdk.NewCoin(types.ExtendedCoinDenom, integerBurnAmount.Mul(types.ConversionFactor()))
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(coin)); err != nil {
			return err
		}
	}

	// Assign new fractional balance in x/precisebank
	k.SetFractionalBalance(ctx, moduleAddr, newFractionalBalance)

	// ----------------------------------------
	// Update remainder & reserves for burned fractional coins
	prevRemainder := k.GetRemainderAmount(ctx)
	// Add to new remainder with burned fractional amount.
	newRemainder := prevRemainder.Add(fractionalBurnAmount)

	// If remainder is greater than or equal to the conversion factor, burn
	// additional integer coin to make reserve just enough to back fractional
	// amounts and nothing more.
	if newRemainder.GTE(types.ConversionFactor()) {
		reserveBurnCoins := sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt()))
		if err := k.bk.BurnCoins(ctx, types.ModuleName, reserveBurnCoins); err != nil {
			return fmt.Errorf("failed to burn %s for reserve: %w", reserveBurnCoins, err)
		}

		// Update remainder with leftover fractional amount.
		// newRemainder > ConversionFactor, and we need to subtract the burned
		// 1 integer coin amount. This is equivalent to .Mod() in this case.
		newRemainder = newRemainder.Sub(types.ConversionFactor())
	}

	k.SetRemainderAmount(ctx, newRemainder)

	return nil
}
