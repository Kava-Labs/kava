package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// TestKeeper is a test wrapper for the keeper which contains useful methods for testing
type TestKeeper struct {
	keeper.Keeper
}

func (keeper TestKeeper) StoreGlobalBorrowIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetHardBorrowRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) StoreGlobalSupplyIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetHardSupplyRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) StoreGlobalDelegatorIndexes(ctx sdk.Context, multiRewardIndexes types.MultiRewardIndexes) {
	// Hardcoded to use bond denom
	multiRewardIndex, _ := multiRewardIndexes.GetRewardIndex(types.BondDenom)
	keeper.SetDelegatorRewardIndexes(ctx, types.BondDenom, multiRewardIndex.RewardIndexes)
}

func (keeper TestKeeper) StoreGlobalSwapIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetSwapRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) StoreGlobalSavingsIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetSavingsRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) StoreGlobalEarnIndexes(ctx sdk.Context, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.SetEarnRewardIndexes(ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (keeper TestKeeper) StoreGlobalIndexes(ctx sdk.Context, claimType types.ClaimType, indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		keeper.Store.SetRewardIndexes(ctx, claimType, i.CollateralType, i.RewardIndexes)
	}
}
