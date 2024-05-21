package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// GetBalance returns the balance of a specific denom for an address. This will
// return the extended balance for the ExtendedCoinDenom, and the regular
// balance for all other denoms.
func (k Keeper) GetBalance(
	ctx sdk.Context,
	addr sdk.AccAddress,
	denom string,
) sdk.Coin {
	// Pass through to x/bank for denoms except ExtendedCoinDenom
	if denom != types.ExtendedCoinDenom {
		return k.bk.GetBalance(ctx, addr, denom)
	}

	// x/bank for integer balance - spendable balance only
	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	integerAmount := spendableCoins.AmountOf(types.IntegerCoinDenom)

	// x/precisebank for fractional balance
	fractionalAmount, found := k.GetFractionalBalance(ctx, addr)
	if !found {
		fractionalAmount = sdk.ZeroInt()
	}

	// (Integer * ConversionFactor) + Fractional
	fullAmount := integerAmount.
		Mul(types.ConversionFactor()).
		Add(fractionalAmount)

	return sdk.NewCoin(types.ExtendedCoinDenom, fullAmount)
}
