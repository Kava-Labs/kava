package keeper

import (
	"github.com/kava-labs/kava/x/savings/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers the savings module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "deposits", DepositsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "solvency", SolvencyInvariant(k))
}

// AllInvariants runs all invariants of the savings module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if res, stop := DepositsInvariant(k)(ctx); stop {
			return res, stop
		}

		res, stop := SolvencyInvariant(k)(ctx)
		return res, stop
	}
}

// DepositsInvariant iterates all deposits and asserts that they are valid
func DepositsInvariant(k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "validate deposits broken", "deposit invalid")

	return func(ctx sdk.Context) (string, bool) {
		k.IterateDeposits(ctx, func(deposit types.Deposit) bool {
			if err := deposit.Validate(); err != nil {
				broken = true
				return true
			}
			if !deposit.Amount.IsAllPositive() {
				broken = true
				return true
			}
			return false
		})

		return message, broken
	}
}

// SolvencyInvariant iterates all deposits and ensures the total amount matches the module account coins
func SolvencyInvariant(k Keeper) sdk.Invariant {
	message := sdk.FormatInvariant(types.ModuleName, "module solvency broken", "total deposited amount does not match module account")

	return func(ctx sdk.Context) (string, bool) {
		balance := k.GetSavingsModuleAccountBalances(ctx)

		deposited := sdk.Coins{}
		k.IterateDeposits(ctx, func(deposit types.Deposit) bool {
			for _, coin := range deposit.Amount {
				deposited = deposited.Add(coin)
			}
			return false
		})

		broken := !deposited.IsEqual(balance)
		return message, broken
	}
}
