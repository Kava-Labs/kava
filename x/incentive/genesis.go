package incentive

import (
	"fmt"
	"time"

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
		if _, found := cdpKeeper.GetCollateral(ctx, rp.CollateralType); !found {
			panic(fmt.Sprintf("incentive params contain collateral not found in cdp params: %s", rp.CollateralType))
		}
	}
	// TODO more param validation?

	k.SetParams(ctx, gs.Params)

	// USDX Minting
	for _, claim := range gs.USDXMintingClaims {
		k.SetUSDXMintingClaim(ctx, claim)
	}
	for _, gat := range gs.USDXRewardState.AccumulationTimes {
		k.SetPreviousUSDXMintingAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.USDXRewardState.MultiRewardIndexes {
		factor, found := mri.RewardIndexes.Get(types.USDXMintingRewardDenom)
		if !found || len(mri.RewardIndexes) != 1 {
			panic(fmt.Sprintf("USDX Minting reward factors must only have denom %s", types.USDXMintingRewardDenom))
		}
		k.SetUSDXMintingRewardFactor(ctx, mri.CollateralType, factor)
	}

	// Hard Supply / Borrow
	for _, claim := range gs.HardLiquidityProviderClaims {
		k.SetHardLiquidityProviderClaim(ctx, claim)
	}
	for _, gat := range gs.HardSupplyRewardState.AccumulationTimes {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.HardSupplyRewardState.MultiRewardIndexes {
		k.SetHardSupplyRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}
	for _, gat := range gs.HardBorrowRewardState.AccumulationTimes {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.HardBorrowRewardState.MultiRewardIndexes {
		k.SetHardBorrowRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}

	// Delegator
	for _, claim := range gs.DelegatorClaims {
		k.SetDelegatorClaim(ctx, claim)
	}
	for _, gat := range gs.DelegatorRewardState.AccumulationTimes {
		k.SetPreviousDelegatorRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.DelegatorRewardState.MultiRewardIndexes {
		k.SetDelegatorRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}

	// Swap
	for _, claim := range gs.SwapClaims {
		k.SetSwapClaim(ctx, claim)
	}
	for _, gat := range gs.SwapRewardState.AccumulationTimes {
		k.SetSwapRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.SwapRewardState.MultiRewardIndexes {
		k.SetSwapRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}
}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)

	usdxClaims := k.GetAllUSDXMintingClaims(ctx)
	usdxRewardState := getUSDXMintingGenesisRewardState(ctx, k)

	hardClaims := k.GetAllHardLiquidityProviderClaims(ctx)
	hardSupplyRewardState := getHardSupplyGenesisRewardState(ctx, k)
	hardBorrowRewardState := getHardBorrowGenesisRewardState(ctx, k)

	delegatorClaims := k.GetAllDelegatorClaims(ctx)
	delegatorRewardState := getDelegatorGenesisRewardState(ctx, k)

	swapClaims := k.GetAllSwapClaims(ctx)
	swapRewardState := getSwapGenesisRewardState(ctx, k)

	return types.NewGenesisState(
		params,
		usdxRewardState, hardSupplyRewardState, hardBorrowRewardState, delegatorRewardState, swapRewardState,
		usdxClaims, hardClaims, delegatorClaims, swapClaims,
	)
}

func getUSDXMintingGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {

	var ats AccumulationTimes
	keeper.IterateUSDXMintingAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris MultiRewardIndexes
	keeper.IterateUSDXMintingRewardFactors(ctx, func(ctype string, factor sdk.Dec) bool {
		mris = append(
			mris,
			types.NewMultiRewardIndex(
				ctype,
				types.RewardIndexes{types.NewRewardIndex(types.USDXMintingRewardDenom, factor)},
			),
		)
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getHardSupplyGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {

	var ats AccumulationTimes
	keeper.IterateHardSupplyRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris MultiRewardIndexes
	keeper.IterateHardSupplyRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getHardBorrowGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {

	var ats AccumulationTimes
	keeper.IterateHardBorrowRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris MultiRewardIndexes
	keeper.IterateHardBorrowRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getDelegatorGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {

	var ats AccumulationTimes
	keeper.IterateDelegatorRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris MultiRewardIndexes
	keeper.IterateDelegatorRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getSwapGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {

	var ats AccumulationTimes
	keeper.IterateSwapRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris MultiRewardIndexes
	keeper.IterateSwapRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}
