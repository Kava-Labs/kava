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
		k.SetHardSupplyRewardIndexes(ctx, mrp.CollateralType, newRewardIndexes)
	}

	for _, gat := range gs.HardDelegatorAccumulationTimes {
		k.SetPreviousHardDelegatorRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
		k.SetHardBorrowRewardIndexes(ctx, mrp.CollateralType, newRewardIndexes)
	}

	for _, rp := range gs.Params.HardDelegatorRewardPeriods {
		k.SetHardDelegatorRewardFactor(ctx, rp.CollateralType, sdk.ZeroDec())
	}

	k.SetParams(ctx, gs.Params)

	for _, gat := range gs.USDXAccumulationTimes {
		k.SetPreviousUSDXMintingAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
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

	hardDelegatorGats := GenesisAccumulationTimes{}
	k.IterateHardDelegatorRewardAccrualTimes(ctx, func(ct string, accTime time.Time) bool {
		hardDelegatorGats = append(hardDelegatorGats, types.NewGenesisAccumulationTime(ct, accTime))
		return false
	})

	hardSupplyGats := GenesisAccumulationTimes{}
	k.IterateHardSupplyRewardAccrualTimes(ctx, func(denom string, accTime time.Time) bool {
		hardSupplyGats = append(hardSupplyGats, types.NewGenesisAccumulationTime(denom, accTime))
		return false
	})

	hardBorrowGats := GenesisAccumulationTimes{}
	k.IterateHardBorrowRewardAccrualTimes(ctx, func(denom string, accTime time.Time) bool {
		hardBorrowGats = append(hardBorrowGats, types.NewGenesisAccumulationTime(denom, accTime))
		return false
	})

	return types.NewGenesisState(
		params,
		usdxMintingGats, hardSupplyGats, hardBorrowGats, hardDelegatorGats,
		usdxClaims,
		hardClaims,
	)
}
