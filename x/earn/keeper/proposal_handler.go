package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/types"
)

// HandleCommunityPoolDepositProposal is a handler for executing a passed community pool deposit proposal
func HandleCommunityPoolDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolDepositProposal) error {
	return k.DepositFromModuleAccount(ctx, k.communityPoolMaccName, p.Amount, types.STRATEGY_TYPE_SAVINGS)
}

// HandleCommunityPoolWithdrawProposal is a handler for executing a passed community pool withdraw proposal.
func HandleCommunityPoolWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolWithdrawProposal) error {
	_, err := k.WithdrawFromModuleAccount(ctx, k.communityPoolMaccName, p.Amount, types.STRATEGY_TYPE_SAVINGS)
	return err
}
