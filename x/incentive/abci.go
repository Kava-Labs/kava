package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
)

// BeginBlocker runs at the start of every block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	for _, rp := range k.GetParams(ctx).USDXMintingRewardPeriods {
		err := k.AccumulateUSDXMintingRewards(ctx, rp)
		if err != nil {
			panic(err)
		}
	}
}
