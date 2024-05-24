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

	// No more processing required if no ExtendedCoinDenom
	if extendedAmount.IsZero() {
		return nil
	}

	return k.mintExtendedCoin(ctx, moduleName, extendedAmount)
}

// mintExtendedCoin manages the minting of extended coins, and no other coins.
func (k Keeper) mintExtendedCoin(
	ctx sdk.Context,
	moduleName string,
	amt sdkmath.Int,
) error {
	moduleAddr := k.ak.GetModuleAddress(moduleName)

	// Get current module account fractional balance - 0 if not found
	fractionalAmount := k.GetFractionalBalance(ctx, moduleAddr)

	// Get separated mint amounts
	integerMintAmount := amt.Quo(types.ConversionFactor())
	fractionalMintAmount := amt.Mod(types.ConversionFactor())

	// Get new fractional balance after minting, this could be greater than
	// the conversion factor and must be checked for carry over to integer mint
	// amount as being set as-is may cause fractional balance exceeding max.
	newFractionalBalance := fractionalAmount.Add(fractionalMintAmount)

	// If it carries over, add 1 to integer mint amount. In this case, it will
	// always be 1:
	// fractional amounts x and y where both x and y < ConversionFactor
	// x + y < (2 * ConversionFactor) - 2
	// x + y < 1 integer amount + fractional amount
	if newFractionalBalance.GTE(types.ConversionFactor()) {
		// Carry over to integer mint amount
		integerMintAmount = integerMintAmount.AddRaw(1)
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
			moduleName,
			sdk.NewCoins(integerMintCoin),
		); err != nil {
			return err
		}
	}

	// Assign new fractional balance in x/precisebank
	k.SetFractionalBalance(ctx, moduleAddr, newFractionalBalance)

	// ----------------------------------------
	// Update remainder & reserves to back minted fractional coins
	prevRemainder := k.GetRemainderAmount(ctx)
	// Deduct new remainder with minted fractional amount
	newRemainder := prevRemainder.Sub(fractionalMintAmount)

	if prevRemainder.LT(fractionalMintAmount) {
		// Need additional 1 integer coin in reserve to back minted fractional
		reserveMintCoins := sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.OneInt()))
		if err := k.bk.MintCoins(ctx, types.ModuleName, reserveMintCoins); err != nil {
			return fmt.Errorf("failed to mint %s for reserve: %w", reserveMintCoins, err)
		}

		// Update remainder with value of minted integer coin. newRemainder is
		// currently negative at this point. This also means that it will always
		// be < conversionFactor after this operation and not require a Mod().
		newRemainder = newRemainder.Add(types.ConversionFactor())
	}

	k.SetRemainderAmount(ctx, newRemainder)

	return nil
}
