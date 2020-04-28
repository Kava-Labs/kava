package incentive

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, supplyKeeper types.SupplyKeeper, gs types.GenesisState) {

	// check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, types.IncentiveMacc)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.IncentiveMacc))
	}

	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	k.SetParams(ctx, gs.Params)

	for _, r := range gs.Params.Rewards {
		k.SetNextClaimPeriodID(ctx, r.Denom, 1)
	}

	// only set the previous block time if it's different than default
	if !gs.PreviousBlockTime.Equal(types.DefaultPreviousBlockTime) {
		k.SetPreviousBlockTime(ctx, gs.PreviousBlockTime)
	}

	// set store objects
	for _, rp := range gs.RewardPeriods {
		k.SetRewardPeriod(ctx, rp)
	}

	for _, cp := range gs.ClaimPeriods {
		k.SetClaimPeriod(ctx, cp)
	}

	for _, c := range gs.Claims {
		k.SetClaim(ctx, c)
	}

	for _, id := range gs.NextClaimPeriodIDs {
		k.SetNextClaimPeriodID(ctx, id.Denom, id.ID)
	}

}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	// get all objects out of the store
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)

	// since it is not set in genesis, if somehow the chain got started and was exported
	// immediately after InitGenesis, there would be no previousBlockTime value.
	if !found {
		previousBlockTime = types.DefaultPreviousBlockTime
	}

	// Get all objects from the store
	rewardPeriods := k.GetAllRewardPeriods(ctx)
	claimPeriods := k.GetAllClaimPeriods(ctx)
	claims := k.GetAllClaims(ctx)
	claimPeriodIDs := k.GetAllClaimPeriodIDPairs(ctx)

	return types.NewGenesisState(params, previousBlockTime, rewardPeriods, claimPeriods, claims, claimPeriodIDs)
}
