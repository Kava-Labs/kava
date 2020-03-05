package kavadist

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, k Keeper) {
	err := k.MintPeriodRewards(ctx)
	if err != nil {
		panic(err)
	}
}