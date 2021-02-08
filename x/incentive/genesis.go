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

	for _, rp := range gs.Params.USDXMintingRewardPeriods {
		_, found := cdpKeeper.GetCollateral(ctx, rp.CollateralType)
		if !found {
			panic(fmt.Sprintf("usdx minting collateral type %s not found in cdp collateral types", rp.CollateralType))
		}
	}

	k.SetParams(ctx, gs.Params)

	// TODO: previous hard module accrual times/indexes should be set here

	for _, gat := range gs.USDXAccumulationTimes {
		k.SetPreviousUSDXMintingAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
		k.SetUSDXMintingRewardFactor(ctx, gat.CollateralType, gat.RewardFactor)
	}

	for _, gat := range gs.HardSupplyAccumulationTimes {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}

	for _, gat := range gs.HardBorrowAccumulationTimes {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}

	for _, gat := range gs.HardDelegatorAccumulationTimes {
		k.SetPreviousHardDelegatorRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}

	for _, claim := range gs.USDXMintingClaims {
		k.SetUSDXMintingClaim(ctx, claim)
	}

}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)

	usdxClaims := k.GetAllUSDXMintingClaims(ctx)
	hardClaims := k.GetAllHardLiquidityProviderClaims(ctx)

	var gats GenesisAccumulationTimes

	for _, rp := range params.USDXMintingRewardPeriods {
		pat, found := k.GetPreviousUSDXMintingAccrualTime(ctx, rp.CollateralType)
		if !found {
			pat = ctx.BlockTime()
		}
		factor, found := k.GetUSDXMintingRewardFactor(ctx, rp.CollateralType)
		if !found {
			factor = sdk.ZeroDec()
		}
		gat := types.NewGenesisAccumulationTime(rp.CollateralType, pat, factor)
		gats = append(gats, gat)
	}

	return types.NewGenesisState(params, gats, DefaultGenesisAccumulationTimes, DefaultGenesisAccumulationTimes, DefaultGenesisAccumulationTimes, usdxClaims, hardClaims)
}
