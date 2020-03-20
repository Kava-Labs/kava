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
	if !found {
		previousBlockTime = types.DefaultPreviousBlockTime
	}

	rps := types.RewardPeriods{}
	k.IterateRewardPeriods(ctx, func(rp types.RewardPeriod) (stop bool) {
		rps = append(rps, rp)
		return false
	})

	cps := types.ClaimPeriods{}
	k.IterateClaimPeriods(ctx, func(cp types.ClaimPeriod) (stop bool) {
		cps = append(cps, cp)
		return false
	})

	cs := types.Claims{}
	k.IterateClaims(ctx, func(c types.Claim) (stop bool) {
		cs = append(cs, c)
		return false
	})

	ids := types.GenesisClaimPeriodIDs{}
	k.IterateClaimPeriodIDKeysAndValues(ctx, func(denom string, id uint64) (stop bool) {
		genID := types.GenesisClaimPeriodID{
			Denom: denom,
			ID:    id,
		}
		ids = append(ids, genID)
		return false
	})

	// return them as a new genesis state
	return types.NewGenesisState(params, previousBlockTime, rps, cps, cs, ids)
}
