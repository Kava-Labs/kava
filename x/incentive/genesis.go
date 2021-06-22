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
	k.SetParams(ctx, gs.Params)

	for _, gat := range gs.USDXAccumulationTimes {
		k.SetPreviousUSDXMintingAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, gri := range gs.USDXRewardIndexes {
		k.SetUSDXMintingRewardFactor(ctx, gri.CollateralType, gri.RewardIndexes[0].RewardFactor)
	}

	for _, gat := range gs.HardDelegatorAccumulationTimes {
		k.SetPreviousHardDelegatorRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, gri := range gs.HardDelegatorRewardIndexes {
		k.SetHardDelegatorRewardFactor(ctx, gri.CollateralType, gri.RewardIndexes[0].RewardFactor)
	}

	for _, gat := range gs.HardSupplyAccumulationTimes {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, gri := range gs.HardSupplyRewardIndexes {
		k.SetHardSupplyRewardIndexes(ctx, gri.CollateralType, gri.RewardIndexes)
	}

	for _, gat := range gs.HardBorrowAccumulationTimes {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, gri := range gs.HardBorrowRewardIndexes {
		k.SetHardBorrowRewardIndexes(ctx, gri.CollateralType, gri.RewardIndexes)
	}

	for _, claim := range gs.USDXMintingClaims {
		k.SetUSDXMintingClaim(ctx, claim)
	}
	for _, claim := range gs.HardLiquidityProviderClaims {
		k.SetHardLiquidityProviderClaim(ctx, claim)
	}
}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)

	usdxClaims := k.GetAllUSDXMintingClaims(ctx)
	hardClaims := k.GetAllHardLiquidityProviderClaims(ctx)

	// Not using nil for initial slice values as it makes the exported genesis json a bit nicer - represented as `[]` rather than `null`

	usdxMintingGats := GenesisAccumulationTimes{}
	k.IterateUSDXMintingAccrualTimes(ctx, func(ct string, accTime time.Time) bool {
		usdxMintingGats = append(usdxMintingGats, types.NewGenesisAccumulationTime(ct, accTime))
		return false
	})
	usdxMintingGris := types.GenesisRewardIndexesSlice{}
	k.IterateUSDXMintingRewardFactors(ctx, func(ct string, factor sdk.Dec) bool {
		usdxMintingGris = append(usdxMintingGris, types.NewGenesisRewardIndexes(ct, types.RewardIndexes{types.NewRewardIndex(types.USDXMintingRewardDenom, factor)}))
		return false
	})

	hardDelegatorGats := GenesisAccumulationTimes{}
	k.IterateHardDelegatorRewardAccrualTimes(ctx, func(ct string, accTime time.Time) bool {
		hardDelegatorGats = append(hardDelegatorGats, types.NewGenesisAccumulationTime(ct, accTime))
		return false
	})
	hardDelegatorGris := types.GenesisRewardIndexesSlice{}
	k.IterateHardDelegatorRewardFactors(ctx, func(ct string, factor sdk.Dec) bool {
		hardDelegatorGris = append(hardDelegatorGris, types.NewGenesisRewardIndexes(ct, types.RewardIndexes{types.NewRewardIndex(types.HardLiquidityRewardDenom, factor)}))
		return false
	})

	hardSupplyGats := GenesisAccumulationTimes{}
	k.IterateHardSupplyRewardAccrualTimes(ctx, func(denom string, accTime time.Time) bool {
		hardSupplyGats = append(hardSupplyGats, types.NewGenesisAccumulationTime(denom, accTime))
		return false
	})
	hardSupplyGris := types.GenesisRewardIndexesSlice{}
	k.IterateHardSupplyRewardIndexes(ctx, func(ct string, indexes types.RewardIndexes) bool {
		hardSupplyGris = append(hardSupplyGris, types.NewGenesisRewardIndexes(ct, indexes))
		return false
	})

	hardBorrowGats := GenesisAccumulationTimes{}
	k.IterateHardBorrowRewardAccrualTimes(ctx, func(denom string, accTime time.Time) bool {
		hardBorrowGats = append(hardBorrowGats, types.NewGenesisAccumulationTime(denom, accTime))
		return false
	})
	hardBorrowGris := types.GenesisRewardIndexesSlice{}
	k.IterateHardBorrowRewardIndexes(ctx, func(ct string, indexes types.RewardIndexes) bool {
		hardBorrowGris = append(hardBorrowGris, types.NewGenesisRewardIndexes(ct, indexes))
		return false
	})

	return types.NewGenesisState(
		params,
		usdxMintingGats, hardSupplyGats, hardBorrowGats, hardDelegatorGats,
		usdxMintingGris, hardSupplyGris, hardBorrowGris, hardDelegatorGris,
		usdxClaims,
		hardClaims,
	)
}
