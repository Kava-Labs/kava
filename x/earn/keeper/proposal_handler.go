package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/types"
)

// HandleCommunityPoolDepositProposal is a handler for executing a passed community pool deposit proposal
func HandleCommunityPoolDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolDepositProposal) error {
	// deposit from community pool address (kava1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8m2splc)

	// get community pool from dist module, make sure balance is > p.Amount,
	// deduct funds from community pool manually (see DistributeFromFeePool from distribution module)
	feePool := k.distKeeper.GetFeePool(ctx)
	newCommunityPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(p.Amount))
	if negative {
		return fmt.Errorf("deposit amount %s < community pool balance: %s", p.Amount, feePool.CommunityPool)
	}
	// deposit funds
	err := k.Deposit(ctx, k.distKeeper.GetDistributionAccount(ctx).GetAddress(), p.Amount)
	if err != nil {
		return err
	}
	// store updated community pool
	feePool.CommunityPool = newCommunityPool
	k.distKeeper.SetFeePool(ctx, feePool)

	return nil

}

func HandleCommunityPoolWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolWithdrawProposal) error {

	// add funds to community pool manually
	feePool := k.distKeeper.GetFeePool(ctx)
	newCommunityPool := feePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(p.Amount))
	feePool.CommunityPool = newCommunityPool
	// store updated community pool
	k.distKeeper.SetFeePool(ctx, feePool)
	// withdraw funds from community pool via module-to-module transfer
	return k.WithdrawFromModuleAccount(ctx, k.distKeeper.GetDistributionAccount(ctx).GetAddress(), p.Amount)
}
