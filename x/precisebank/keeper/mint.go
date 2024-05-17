package keeper

import (
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
	// TODO: Disallow minting to x/precisebank module
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

	// Get current fractional amount
	// TODO: GetFractionalBalance should just return 0 if not found and not return a bool
	// since we don't actually need to know if it was found or not.
	fractionalAmount, found := k.GetFractionalBalance(ctx, moduleAddr)
	if !found {
		fractionalAmount = sdk.ZeroInt()
	}

	// Get mint amounts
	integerMintAmount := amt.Quo(types.ConversionFactor())
	fractionalMintAmount := amt.Mod(types.ConversionFactor())

	// Get new fractional balance
	newFractionalBalance := fractionalAmount.Add(fractionalMintAmount)

	// If it carries over, add 1 to integer mint amount. In this case, it will
	// always be 1 as two fractional amounts can only add up < 2
	if newFractionalBalance.GTE(types.ConversionFactor()) {
		integerMintAmount = integerMintAmount.AddRaw(1)
		newFractionalBalance = newFractionalBalance.Mod(types.ConversionFactor())
	}

	// Mint new integer amounts in x/bank
	if !integerMintAmount.IsZero() {
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

	// TODO: Update remainder

	return nil
}
