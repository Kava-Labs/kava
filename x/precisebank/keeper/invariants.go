package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// RegisterInvariants registers the x/precisebank module invariants
func RegisterInvariants(
	ir sdk.InvariantRegistry,
	k Keeper,
	bk types.BankKeeper,
) {
	ir.RegisterRoute(types.ModuleName, "invalid-fractional-total", BalancedFractionalTotalInvariant(k))
}

// AllInvariants runs all invariants of the X/precisebank module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := BalancedFractionalTotalInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		return "", false
	}
}

// BalancedFractionalTotalInvariant checks that the sum of fractional balances
// and the remainder amount is divisible by the conversion factor without any
// leftover amount.
func BalancedFractionalTotalInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		fractionalBalSum := k.GetAggregateSumFractionalBalances(ctx)
		remainderAmount := k.GetRemainderAmount(ctx)

		total := fractionalBalSum.Add(remainderAmount)
		splitBal := types.NewSplitBalanceFromFullAmount(total)

		broken := false
		msg := ""

		if !splitBal.FractionalAmount.IsZero() {
			broken = true
			msg = fmt.Sprintf(
				"(sum(FractionalBalances) + remainder) %% conversionFactor should be 0 but got %v",
				splitBal.FractionalAmount,
			)
		}

		return sdk.FormatInvariant(
			types.ModuleName, "invalid-fractional-total",
			msg,
		), broken
	}
}
