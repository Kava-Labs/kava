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

func (keeper TestKeeper) storeGlobalBorrowIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetHardBorrowRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) storeGlobalSupplyIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetHardSupplyRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) storeGlobalDelegatorIndexes(ctx sdk.Context, multiRewardIndexes types.MultiRewardIndexes) {
	// Hardcoded to use bond denom
	multiRewardIndex, _ := multiRewardIndexes.GetRewardIndex(types.BondDenom)
	keeper.SetDelegatorRewardIndexes(ctx, types.BondDenom, multiRewardIndex.RewardIndexes)
}

func (keeper TestKeeper) storeGlobalSwapIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetSwapRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) storeGlobalSavingsIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetSavingsRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) storeGlobalEarnIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetEarnRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}
