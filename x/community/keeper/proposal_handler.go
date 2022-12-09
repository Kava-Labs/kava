package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/types"
)

// HandleCommunityPoolLendDepositProposal is a handler for executing a passed community pool lend deposit proposal.
func HandleCommunityPoolLendDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendDepositProposal) error {
	return k.hardKeeper.Deposit(ctx, k.moduleAddress, p.Amount)
}

// HandleCommunityPoolLendWithdrawProposal is a handler for executing a passed community pool lend withdraw proposal.
func HandleCommunityPoolLendWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendWithdrawProposal) error {
	return k.hardKeeper.Withdraw(ctx, k.moduleAddress, p.Amount)
}
