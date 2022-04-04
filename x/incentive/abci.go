package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// BeginBlocker runs at the start of every block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)

	for _, rp := range params.USDXMintingRewardPeriods {
		k.AccumulateRewards(ctx, types.NewMultiRewardPeriodFromRewardPeriod(rp))
	}
	for _, rp := range params.HardSupplyRewardPeriods {
		k.AccumulateRewards(ctx, types.HardSupply, rp)
	}
	for _, rp := range params.HardBorrowRewardPeriods {
		k.AccumulateRewards(ctx, types.HardBorrow, rp)
	}
	for _, rp := range params.DelegatorRewardPeriods {
		k.AccumulateRewards(ctx, types.Delegator, rp)
	}
	for _, rp := range params.SwapRewardPeriods {
		k.AccumulateRewards(ctx, types.Swap, rp)
	}
}
