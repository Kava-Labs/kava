package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavadist/types"
)

const secondsPerYear = 365 * 24 * 3600

func (k *Keeper) DistributeFunds(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found || previousBlockTime.IsZero() {
		previousBlockTime = ctx.BlockTime()
	}
	secondsPassed := ctx.BlockTime().Sub(previousBlockTime).Seconds()
	fmt.Println("SECONDS PASSED: ", secondsPassed)

	bondDenom := k.bondDenom(ctx)
	balanceToDistribute := k.getModuleAccountBalance(ctx).AmountOf(bondDenom)
	initialBalance := balanceToDistribute

	// in a non-spike world these values come from kavadist params
	stakingApy := sdk.MustNewDecFromStr("0.15")
	infraApr := sdk.MustNewDecFromStr("0.25")
	// everything else goes to community pool.
	// TODO: would need to update incentive to draw from x/community macc

	// distribute to infrastructure partners
	infraRewards := balanceToDistribute.ToDec().Mul(infraApr).TruncateInt()
	err := k.distributeInfrastructureCoins(ctx,
		params.InfrastructureParams.PartnerRewards,
		params.InfrastructureParams.CoreRewards,
		sdk.NewInt(int64(secondsPassed)),
		sdk.NewCoin(bondDenom, infraRewards),
	)
	if err != nil {
		return err
	}
	balanceToDistribute = balanceToDistribute.Sub(infraRewards)

	// distribute staking rewards
	var stakingRewards sdk.Int
	stakingRewards, balanceToDistribute, err = k.distributeStakingRewards(ctx, balanceToDistribute, stakingApy, secondsPassed)
	if err != nil {
		return err
	}

	// distribute to community pool
	communityPoolDistr := balanceToDistribute
	if balanceToDistribute.GT(sdk.ZeroInt()) {
		addr := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress()
		err = k.communityKeeper.FundCommunityPool(ctx, addr, sdk.NewCoins(sdk.NewCoin(bondDenom, balanceToDistribute)))
		if err != nil {
			return err
		}
	}

	// set previous block time for next block's staking reward accumulation.
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())

	fmt.Println("KAVADIST DISTRIBUTION MADE")
	fmt.Println("initial balance: ", initialBalance)
	fmt.Println("infra rewards: ", infraRewards)
	fmt.Println("staking rewards: ", stakingRewards)
	fmt.Println("community pool: ", communityPoolDistr)

	return nil
}

// TODO: STAKING APR OR APY?
func (k *Keeper) distributeStakingRewards(ctx sdk.Context, balanceToDistribute sdk.Int, stakingApr sdk.Dec, secondsPassed float64) (rewards sdk.Int, remaining sdk.Int, err error) {
	// determine required staking rewards
	bondDenom := k.bondDenom(ctx)
	totalBonded := k.totalBondedTokens(ctx)

	// accurate to a 10th of a second
	secondsDec := sdk.NewDec(int64(secondsPassed * 10)).QuoInt64(10)

	// ! this probs isn't right APR math. just proof of concept. real math goes here.
	stakingRewards := totalBonded.ToDec().
		Mul(stakingApr).Mul(secondsDec).
		QuoInt64(secondsPerYear).TruncateInt()

	if balanceToDistribute.LT(stakingRewards) {
		// TODO: leftovers aren't enough for staking rewards! Consider pulling from community pool.
		panic("insufficient tokens for staking rewards.")
	}

	// TODO: get fee collector name from keeper.
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, "fee_collector", sdk.NewCoins(sdk.NewCoin(bondDenom, stakingRewards)))
	if err != nil {
		return stakingRewards, balanceToDistribute, err
	}

	return stakingRewards, balanceToDistribute.Sub(stakingRewards), nil
}
