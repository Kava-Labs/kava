package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction/types"
)

// RegisterInvariants registers all staking invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {

	ir.RegisterRoute(types.ModuleName, "module-account",
		ModuleAccountInvariants(k))
	ir.RegisterRoute(types.ModuleName, "valid-auctions",
		ValidAuctionInvariant(k))
	ir.RegisterRoute(types.ModuleName, "valid-index",
		ValidIndexInvariant(k))
}

// ModuleAccountInvariant checks that the module account's coins matches those stored in auctions
func ModuleAccountInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {

		totalAuctionCoins := sdk.NewCoins()
		k.IterateAuctions(ctx, func(auction types.Auction) bool {
			a, ok := auction.(types.GenesisAuction)
			if !ok {
				panic("stored auction type does not fulfill GenesisAuction interface")
			}
			totalAuctionCoins = totalAuctionCoins.Add(a.GetModuleAccountCoins())
			return false
		})
		moduleAccCoins := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins()
		broken := !moduleAccCoins.IsEqual(totalAuctionCoins)

		// Bonded tokens should equal sum of tokens with bonded validators
		// Not-bonded tokens should equal unbonding delegations	plus tokens on unbonded validators
		invariantMessage := sdk.FormatInvariant(
			types.ModuleName,
			"ModuleAccount coins",
			fmt.Sprintf(
				"\texpected ModuleAccount coins: %s\n"+
					"\tactual ModuleAccount coins:   %s\n",
				totalAuctionCoins, moduleAccCoins),
		)

		return invariantMessage, broken
	}
}

func ValidAuctionInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var validationErr error
		var invalidAuction types.Auction
		k.IterateAuctions(ctx, func(auction types.Auction) bool {
			a, ok := auction.(types.GenesisAuction)
			if !ok {
				panic("stored auction type does not fulfill GenesisAuction interface")
			}
			// TODO other validation? like end times compared to current block time
			if err := a.Validate(); err != nil {
				validationErr = err
				invalidAuction = a
				return true
			}
			return false
		})
		broken := validationErr != nil
		invariantMessage := sdk.FormatInvariant(
			types.ModuleName,
			"Valid Auctions",
			fmt.Sprintf(
				"\tfound invalid auction, reason: %s\n"+
					"\tauction:\n%s\n",
				validationErr, invalidAuction),
		)
		return invariantMessage, broken
	}
}

// TODO check all auctions are in the index / all ids in the index exist in the auction store
func ValidIndexInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return "", false
	}
}
