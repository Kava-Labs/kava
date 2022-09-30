package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

// HandleCommunityPoolDepositProposal is a handler for executing a passed community pool deposit proposal
func HandleCommunityPoolDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolDepositProposal) error {
	fundAcc := k.accountKeeper.GetModuleAccount(ctx, kavadisttypes.FundModuleAccount)
	if err := k.distKeeper.DistributeFromFeePool(ctx, sdk.NewCoins(p.Amount), fundAcc.GetAddress()); err != nil {
		return err
	}

	err := k.DepositFromModuleAccount(ctx, kavadisttypes.FundModuleAccount, p.Amount, types.STRATEGY_TYPE_SAVINGS)
	if err != nil {
		return err
	}

	return nil

}

// HandleCommunityPoolWithdrawProposal is a handler for executing a passed community pool withdraw proposal.
func HandleCommunityPoolWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolWithdrawProposal) error {
	// Withdraw to fund module account
	withdrawAmount, err := k.WithdrawFromModuleAccount(ctx, kavadisttypes.FundModuleAccount, p.Amount, types.STRATEGY_TYPE_SAVINGS)
	if err != nil {
		return err
	}

	// Move funds to the community pool manually
	err = k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		kavadisttypes.FundModuleAccount,
		k.distKeeper.GetDistributionAccount(ctx).GetName(),
		sdk.NewCoins(withdrawAmount),
	)
	if err != nil {
		return err
	}
	feePool := k.distKeeper.GetFeePool(ctx)
	newCommunityPool := feePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(withdrawAmount))
	feePool.CommunityPool = newCommunityPool
	k.distKeeper.SetFeePool(ctx, feePool)
	return nil
}
