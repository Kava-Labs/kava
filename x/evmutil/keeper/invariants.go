package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// RegisterInvariants registers the swap module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, bankK types.BankKeeper, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "fully-backed", FullyBackedInvariant(bankK, k))
	ir.RegisterRoute(types.ModuleName, "small-balances", SmallBalancesInvariant(bankK, k))
	ir.RegisterRoute(types.ModuleName, "cosmos-coins-fully-backed", CosmosCoinsFullyBackedInvariant(bankK, k))
	// Disable this invariant due to some issues with it requiring some staking params to be set in genesis.
	// ir.RegisterRoute(types.ModuleName, "backed-conversion-coins", BackedCoinsInvariant(bankK, k))
}

// AllInvariants runs all invariants of the swap module
func AllInvariants(bankK types.BankKeeper, k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if res, stop := FullyBackedInvariant(bankK, k)(ctx); stop {
			return res, stop
		}
		if res, stop := BackedCoinsInvariant(bankK, k)(ctx); stop {
			return res, stop
		}
		if res, stop := CosmosCoinsFullyBackedInvariant(bankK, k)(ctx); stop {
			return res, stop
		}
		return SmallBalancesInvariant(bankK, k)(ctx)
	}
}

// FullyBackedInvariant ensures all minor balances are backed by the coins in the module account.
//
// The module balance can be greater than the sum of all minor balances. This can happen in rare cases
// where the evm module burns tokens.
func FullyBackedInvariant(bankK types.BankKeeper, k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "fully backed broken", "sum of minor balances greater than module account")

	return func(ctx sdk.Context) (string, bool) {
		totalMinorBalances := sdk.ZeroInt()
		k.IterateAllAccounts(ctx, func(acc types.Account) bool {
			totalMinorBalances = totalMinorBalances.Add(acc.Balance)
			return false
		})

		bankAddr := authtypes.NewModuleAddress(types.ModuleName)
		bankBalance := bankK.GetBalance(ctx, bankAddr, CosmosDenom).Amount.Mul(ConversionMultiplier)

		broken = totalMinorBalances.GT(bankBalance)

		return message, broken
	}
}

// SmallBalancesInvariant ensures all minor balances are less than the overflow amount, beyond this they should be converted to the major denom.
func SmallBalancesInvariant(_ types.BankKeeper, k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "small balances broken", "minor balances not all less than overflow")

	return func(ctx sdk.Context) (string, bool) {
		k.IterateAllAccounts(ctx, func(account types.Account) bool {
			if account.Balance.GTE(ConversionMultiplier) {
				broken = true
				return true
			}
			return false
		})
		return message, broken
	}
}

// BackedCoinsInvariant iterates all conversion pairs and asserts that the
// sdk.Coin balances are less than the module ERC20 balance.
// **Note:** This compares <= and not == as anyone can send tokens to the
// ERC20 contract address and break the invariant if a strict equal check.
func BackedCoinsInvariant(_ types.BankKeeper, k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(
		types.ModuleName,
		"backed coins broken",
		"coin supply is greater than module account ERC20 tokens",
	)

	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		for _, pair := range params.EnabledConversionPairs {
			erc20Balance, err := k.QueryERC20BalanceOf(
				ctx,
				pair.GetAddress(),
				types.NewInternalEVMAddress(types.ModuleEVMAddress),
			)
			if err != nil {
				panic(err)
			}

			supply := k.bankKeeper.GetSupply(ctx, pair.Denom)

			// Must be true: sdk.Coin supply < ERC20 balanceOf(module account)
			if supply.Amount.BigInt().Cmp(erc20Balance) > 0 {
				broken = true
				break
			}
		}

		return message, broken
	}
}

// CosmosCoinsFullyBackedInvariant ensures the total supply of ERC20 representations of sdk.Coins
// match the balances in the module account.
//
// This invariant depends on the fact that coins can only become part of the balance through
// conversion to ERC20s.
// If in the future sdk.Coins can be sent directly to the module account,
// or the module account balance can be increased in any other way,
// this invariant should be changed from checking that the balance equals the total supply,
// to check that the balance is greater than or equal to the total supply.
func CosmosCoinsFullyBackedInvariant(bankK types.BankKeeper, k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(
		types.ModuleName,
		"cosmos coins fully-backed broken",
		"ERC20 total supply is not equal to module account balance",
	)
	maccAddress := authtypes.NewModuleAddress(types.ModuleName)

	return func(ctx sdk.Context) (string, bool) {
		k.IterateAllDeployedCosmosCoinContracts(ctx, func(c types.DeployedCosmosCoinContract) bool {
			moduleBalance := bankK.GetBalance(ctx, maccAddress, c.CosmosDenom).Amount
			totalSupply, err := k.QueryERC20TotalSupply(ctx, *c.Address)
			if err != nil {
				panic(fmt.Sprintf("failed to query total supply for %+v", c))
			}
			// expect total supply to equal balance in the module
			if totalSupply.Cmp(moduleBalance.BigInt()) != 0 {
				broken = true
			}
			return broken
		})
		return message, broken
	}
}
