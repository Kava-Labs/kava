package issuance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/issuance/keeper"
)

// BeginBlocker iterates over each asset and seizes coins from blocked addresses by returning them to the asset owner
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	err := k.SeizeCoinsForBlockableAssets(ctx)
	if err != nil {
		panic(err)
	}
	k.SynchronizeBlockList(ctx)
	k.UpdateTimeBasedSupplyLimits(ctx)
}
