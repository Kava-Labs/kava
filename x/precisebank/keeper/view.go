package keeper

import (
	"context"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// GetBalance returns the balance of a specific denom for an address. This will
// return the extended balance for the ExtendedCoinDenom, and the regular
// balance for all other denoms.
func (k Keeper) GetBalance(
	ctx context.Context,
	addr sdk.AccAddress,
	denom string,
) sdk.Coin {
	// Module balance should display as empty for extended denom. Module
	// balances are **only** for the reserve which backs the fractional
	// balances. Returning the backing balances if querying extended denom would
	// result in a double counting of the fractional balances.
	if denom == types.ExtendedCoinDenom && addr.Equals(k.ak.GetModuleAddress(types.ModuleName)) {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Pass through to x/bank for denoms except ExtendedCoinDenom
	if denom != types.ExtendedCoinDenom {
		return k.bk.GetBalance(sdkCtx, addr, denom)
	}

	// x/bank for integer balance - full balance, including locked
	integerCoins := k.bk.GetBalance(sdkCtx, addr, types.IntegerCoinDenom)

	// x/precisebank for fractional balance
	fractionalAmount := k.GetFractionalBalance(sdkCtx, addr)

	// (Integer * ConversionFactor) + Fractional
	fullAmount := integerCoins.
		Amount.
		Mul(types.ConversionFactor()).
		Add(fractionalAmount)

	return sdk.NewCoin(types.ExtendedCoinDenom, fullAmount)
}

// SpendableCoins returns the total balances of spendable coins for an account
// by address. If the account has no spendable coins, an empty Coins slice is
// returned.
func (k Keeper) SpendableCoin(
	ctx context.Context,
	addr sdk.AccAddress,
	denom string,
) sdk.Coin {
	// Same as GetBalance, extended denom balances are transparent to consumers.
	if denom == types.ExtendedCoinDenom && addr.Equals(k.ak.GetModuleAddress(types.ModuleName)) {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	}

	// Pass through to x/bank for denoms except ExtendedCoinDenom
	if denom != types.ExtendedCoinDenom {
		return k.bk.SpendableCoin(ctx, addr, denom)
	}

	// x/bank for integer balance - excluding locked
	integerCoin := k.bk.SpendableCoin(ctx, addr, types.IntegerCoinDenom)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// x/precisebank for fractional balance
	fractionalAmount := k.GetFractionalBalance(sdkCtx, addr)

	// Spendable = (Integer * ConversionFactor) + Fractional
	fullAmount := integerCoin.Amount.
		Mul(types.ConversionFactor()).
		Add(fractionalAmount)

	return sdk.NewCoin(types.ExtendedCoinDenom, fullAmount)
}
