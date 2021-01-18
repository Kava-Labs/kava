package incentive

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, supplyKeeper types.SupplyKeeper, cdpKeeper types.CdpKeeper, gs types.GenesisState) {

	// check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, types.IncentiveMacc)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.IncentiveMacc))
	}

	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	for _, rp := range gs.Params.RewardPeriods {
		_, found := cdpKeeper.GetCollateral(ctx, rp.CollateralType)
		if !found {
			panic(fmt.Sprintf("usdx minting collateral type %s not found in cdp collateral types", rp.CollateralType))
		}
	}

	k.SetParams(ctx, gs.Params)

	for _, gat := range gs.PreviousAccumulationTimes {
		k.SetPreviousAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
		k.SetRewardFactor(ctx, gat.CollateralType, gat.RewardFactor)
	}

	for _, claim := range gs.USDXMintingClaims {
		k.SetClaim(ctx, claim)
	}

}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)

	claims := k.GetAllClaims(ctx)

	var gats GenesisAccumulationTimes

	for _, rp := range params.RewardPeriods {
		pat, found := k.GetPreviousAccrualTime(ctx, rp.CollateralType)
		if !found {
			pat = ctx.BlockTime()
		}
		factor, found := k.GetRewardFactor(ctx, rp.CollateralType)
		if !found {
			factor = sdk.ZeroDec()
		}
		gat := types.NewGenesisAccumulationTime(rp.CollateralType, pat, factor)
		gats = append(gats, gat)
	}

	return types.NewGenesisState(params, gats, claims)
}
