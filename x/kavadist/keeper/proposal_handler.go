package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// HandleCommunityPoolLendDepositProposal handles a gov proposal for depositing community pool funds into x/hard
func HandleCommunityPoolLendDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendDepositProposal) error {
	moduleAddress := k.maccAddress(ctx)
	// move funds from community pool to kavadist so hard position is help by kavadist
	err := k.distKeeper.DistributeFromFeePool(ctx, p.Amount, moduleAddress)
	if err != nil {
		return err
	}
	// deposit funds into hard
	return k.hardKeeper.Deposit(ctx, moduleAddress, p.Amount)
}

// HandleCommunityPoolLendWithdrawProposal handles a gov proposal for withdrawing community pool positions in x/hard
func HandleCommunityPoolLendWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendWithdrawProposal) error {
	moduleAddress := k.maccAddress(ctx)
	// hard allows attempting to withdraw more funds than there is a deposit for.
	// this means the amount that gets withdrawn will not necessarily match the amount set in the proposal.
	// to calculate how much is withdrawn, compare this module's balance before & after withdraw.
	balanceBefore := k.bankKeeper.GetAllBalances(ctx, moduleAddress)

	// withdraw funds from x/hard to kavadist module account
	err := k.hardKeeper.Withdraw(ctx, moduleAddress, p.Amount)
	if err != nil {
		return err
	}

	balanceAfter := k.bankKeeper.GetAllBalances(ctx, moduleAddress)
	totalWithdrawn := balanceAfter.Sub(balanceBefore)

	// send all withdrawn coins back to community pool
	return k.distKeeper.FundCommunityPool(ctx, totalWithdrawn, moduleAddress)
}

// HandleCommunityPoolMultiSpendProposal is a handler for executing a passed community multi-spend proposal
func HandleCommunityPoolMultiSpendProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolMultiSpendProposal) error {
	for _, receiverInfo := range p.RecipientList {
		if k.blacklistedAddrs[receiverInfo.Address] {
			return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is blacklisted from receiving external funds", receiverInfo.Address)
		}
		err := k.distKeeper.DistributeFromFeePool(ctx, receiverInfo.Amount, receiverInfo.GetAddress())
		if err != nil {
			return err
		}
	}
	return nil
}
