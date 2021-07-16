package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
)

// BeginBlocker runs at the start of every block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)
	for _, rp := range params.USDXMintingRewardPeriods {
		err := k.AccumulateUSDXMintingRewards(ctx, rp)
		if err != nil {
			panic(err)
		}
	}
	for _, rp := range params.HardSupplyRewardPeriods {
		k.AccumulateHardSupplyRewards(ctx, rp)
	}
	for _, rp := range params.HardBorrowRewardPeriods {
		k.AccumulateHardBorrowRewards(ctx, rp)
	}
	for _, rp := range params.DelegatorRewardPeriods {
		k.AccumulateDelegatorRewards(ctx, rp)
	}
	for _, rp := range params.SwapRewardPeriods {
		k.AccumulateSwapRewards(ctx, rp)
	}
}
