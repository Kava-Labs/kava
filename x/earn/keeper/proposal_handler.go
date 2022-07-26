package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/types"
)

// HandleCommunityPoolDepositProposal is a handler for executing a passed community pool deposit proposal
func HandleCommunityPoolDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolDepositProposal) error {
	// kava1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8m2splc
	return k.Deposit(ctx, k.accountKeeper.GetModuleAddress("distribution"), p.Amount)
}
