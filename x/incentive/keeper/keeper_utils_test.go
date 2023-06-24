package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// TestKeeper is a test wrapper for the keeper which contains useful methods for testing
type TestKeeper struct {
	keeper.Keeper
}

func (keeper TestKeeper) storeGlobalEarnIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetEarnRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}
