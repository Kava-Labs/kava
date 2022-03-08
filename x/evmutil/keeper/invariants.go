package keeper

import (
	"github.com/kava-labs/kava/x/evmutil/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers the swap module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, bankK EvmBankKeeper, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "balances", BalancesInvariant(bankK, k))
}

// AllInvariants runs all invariants of the swap module
func AllInvariants(bankK EvmBankKeeper, k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := BalancesInvariant(bankK, k)(ctx)
		return res, stop
	}
}

// BalancesInvariant ensures all minor balances are backed exactly by the coins in the module account.
func BalancesInvariant(bankK EvmBankKeeper, k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "balances broken", "minor balances do not match module account")

	return func(ctx sdk.Context) (string, bool) {

		totalMinorBalances := sdk.ZeroInt()
		k.IterateAllAccounts(ctx, func(acc types.Account) bool {
			totalMinorBalances = totalMinorBalances.Add(acc.Balance)
			return false
		})

		bankAddr := bankK.GetModuleAddress(types.ModuleName)
		bankBalance := bankK.GetBalance(ctx, bankAddr, EvmDenom)

		broken = !totalMinorBalances.Equal(bankBalance.Amount)

		return message, broken
	}
}
