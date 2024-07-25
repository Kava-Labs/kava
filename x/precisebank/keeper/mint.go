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

// MintCoins creates new coins from thin air and adds it to the module account.
// If ExtendedCoinDenom is provided, the corresponding fractional amount is
// added to the module state.
// It will panic if the module account does not exist or is unauthorized.
func (k Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Disallow minting to x/precisebank module
	if moduleName == types.ModuleName {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s cannot be minted to", moduleName))
	}

	// Note: MintingRestrictionFn is not used in x/precisebank
	// Panic errors are identical to x/bank for consistency.
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Minter) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName))
	}

	// Ensure the coins are valid before minting
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
		if err := k.bk.MintCoins(ctx, moduleName, passthroughCoins); err != nil {
			return err
		}
	}

	// Only mint extended coin if the amount is positive
	if extendedAmount.IsPositive() {
		if err := k.mintExtendedCoin(ctx, moduleName, extendedAmount); err != nil {
			return err
		}
	}

	fullEmissionCoins := sdk.NewCoins(types.SumExtendedCoin(amt))
	if fullEmissionCoins.IsZero() {
		return nil
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		banktypes.NewCoinMintEvent(acc.GetAddress(), fullEmissionCoins),
		banktypes.NewCoinReceivedEvent(acc.GetAddress(), fullEmissionCoins),
	})

	return nil
}

// mintExtendedCoin manages the minting of only extended coins. This also
// handles integer carry over from fractional balance to integer balance if
// necessary depending on the fractional balance and minting amount. Ensures
// that the reserve fully backs the additional minted amount, minting any extra
// reserve integer coins if necessary.
// 4 Cases:
// 1. NO integer carry over, >= 0 remainder - no reserve mint
// 2. NO integer carry over, negative remainder - mint 1 to reserve
// 3. Integer carry over, >= 0 remainder
//   - Transfer 1 integer from reserve -> account
//
// 4. Integer carry over, negative remainder
//   - Transfer 1 integer from reserve -> account
//   - Mint 1 to reserve
//     Optimization:
//   - Increase direct account mint amount by 1, no extra reserve mint
func (k Keeper) mintExtendedCoin(
	ctx sdk.Context,
	recipientModuleName string,
	amt sdkmath.Int,
) error {
	moduleAddr := k.ak.GetModuleAddress(recipientModuleName)

	// Get current module account fractional balance - 0 if not found
	fractionalAmount := k.GetFractionalBalance(ctx, moduleAddr)

	// Get separated mint amounts
	integerMintAmount := amt.Quo(types.ConversionFactor())
	fractionalMintAmount := amt.Mod(types.ConversionFactor())

	// Get previous remainder amount, as we need to it before carry calculation
	// for the optimization path.
	prevRemainder := k.GetRemainderAmount(ctx)

	// Deduct new remainder with minted fractional amount. This will result in
	// two cases:
	// 1. Zero or positive remainder: remainder is sufficient to back the minted
	//   fractional amount. Reserve is also sufficient to back the minted amount
	//   so no additional reserve integer coin is needed.
	// 2. Negative remainder: remainder is insufficient to back the minted
	//   fractional amount. Reserve will need to be increased to back the mint
	//   amount.
	newRemainder := prevRemainder.Sub(fractionalMintAmount)

	// Get new fractional balance after minting, this could be greater than
	// the conversion factor and must be checked for carry over to integer mint
	// amount as being set as-is may cause fractional balance exceeding max.
	newFractionalBalance := fractionalAmount.Add(fractionalMintAmount)

	// Case #3 - Integer carry, remainder is sufficient (0 or positive)
	if newFractionalBalance.GTE(types.ConversionFactor()) && newRemainder.GTE(sdkmath.ZeroInt()) {
		// Carry should send from reserve -> account, instead of minting an
		// extra integer coin. Otherwise doing an extra mint will require a burn
		// from reserves to maintain exact backing.
		carryCoin := sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt())

		// SendCoinsFromModuleToModule allows for sending coins even if the
		// recipient module account is blocked.
		if err := k.bk.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			recipientModuleName,
			sdk.NewCoins(carryCoin),
		); err != nil {
			return err
		}
	}

	// Case #4 - Integer carry, remainder is insufficient
	// This is the optimization path where the integer mint amount is increased
	// by 1, instead of doing both a reserve -> account transfer and reserve mint.
	if newFractionalBalance.GTE(types.ConversionFactor()) && newRemainder.IsNegative() {
		integerMintAmount = integerMintAmount.AddRaw(1)
	}

	// If it carries over, adjust the fractional balance to account for the
	// previously added 1 integer amount.
	// fractional amounts x and y where both x and y < ConversionFactor
	// x + y < (2 * ConversionFactor) - 2
	// x + y < 1 integer amount + fractional amount
	if newFractionalBalance.GTE(types.ConversionFactor()) {
		// Subtract 1 integer equivalent amount of fractional balance. Same
		// behavior as using .Mod() in this case.
		newFractionalBalance = newFractionalBalance.Sub(types.ConversionFactor())
	}

	// Mint new integer amounts in x/bank - including carry over from fractional
	// amount if any.
	if integerMintAmount.IsPositive() {
		integerMintCoin := sdk.NewCoin(types.IntegerCoinDenom, integerMintAmount)

		if err := k.bk.MintCoins(
			ctx,
			recipientModuleName,
			sdk.NewCoins(integerMintCoin),
		); err != nil {
			return err
		}
	}

	// Assign new fractional balance in x/precisebank
	k.SetFractionalBalance(ctx, moduleAddr, newFractionalBalance)

	// ----------------------------------------
	// Update remainder & reserves to back minted fractional coins

	// Mint an additional reserve integer coin if remainder is insufficient.
	// The remainder is the amount of fractional coins that can be minted and
	// still be fully backed by reserve. If the remainder is less than the
	// minted fractional amount, then the reserve needs to be increased to
	// back the additional fractional amount.
	// Optimization: This is only done when the integer amount does NOT carry,
	// as a direct account mint is done instead of integer carry transfer +
	// insufficient remainder reserve mint.
	wasCarried := fractionalAmount.Add(fractionalMintAmount).GTE(types.ConversionFactor())
	if prevRemainder.LT(fractionalMintAmount) && !wasCarried {
		// Always only 1 integer coin, as fractionalMintAmount < ConversionFactor
		reserveMintCoins := sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt()))
		if err := k.bk.MintCoins(ctx, types.ModuleName, reserveMintCoins); err != nil {
			return fmt.Errorf("failed to mint %s for reserve: %w", reserveMintCoins, err)
		}
	}

	// newRemainder will be negative if prevRemainder < fractionalMintAmount.
	// This needs to be adjusted back to the corresponding positive value. The
	// remainder will be always < conversionFactor after add if it is negative.
	if newRemainder.IsNegative() {
		newRemainder = newRemainder.Add(types.ConversionFactor())
	}

	k.SetRemainderAmount(ctx, newRemainder)

	return nil
}
