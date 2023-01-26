package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/types"
)

// HandleCommunityPoolLendDepositProposal is a handler for executing a passed community pool lend deposit proposal.
func HandleCommunityPoolLendDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendDepositProposal) error {
	// move funds from community pool to x/community so hard position is held by this module.
	err := k.distrKeeper.DistributeFromFeePool(ctx, p.Amount, k.moduleAddress)
	if err != nil {
		return err
	}
	// deposit funds into hard
	return k.hardKeeper.Deposit(ctx, k.moduleAddress, p.Amount)
}

// HandleCommunityPoolLendWithdrawProposal is a handler for executing a passed community pool lend withdraw proposal.
func HandleCommunityPoolLendWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendWithdrawProposal) error {
	// hard allows attempting to withdraw more funds than there is a deposit for.
	// this means the amount that gets withdrawn will not necessarily match the amount set in the proposal.
	// to calculate how much is withdrawn, compare this module's balance before & after withdraw.
	balanceBefore := k.bankKeeper.GetAllBalances(ctx, k.moduleAddress)

	// withdraw funds from x/hard to this module account
	err := k.hardKeeper.Withdraw(ctx, k.moduleAddress, p.Amount)
	if err != nil {
		return err
	}

	balanceAfter := k.bankKeeper.GetAllBalances(ctx, k.moduleAddress)
	totalWithdrawn := balanceAfter.Sub(balanceBefore)

	// send all withdrawn coins back to community pool
	return k.distrKeeper.FundCommunityPool(ctx, totalWithdrawn, k.moduleAddress)
}
