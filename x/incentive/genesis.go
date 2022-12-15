package incentive

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

const year = 365 * 24 * time.Hour

// EarliestValidAccumulationTime is how far behind the genesis time an accumulation time can be for it to be valid.
// It's a safety check to ensure rewards aren't accidentally accumulated for many years on the first block (eg since Jan 1970).
var EarliestValidAccumulationTime time.Duration = year

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	cdpKeeper types.CdpKeeper,
	gs types.GenesisState,
) {
	// check if the module account exists
	moduleAcc := accountKeeper.GetModuleAccount(ctx, types.IncentiveMacc)
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

	// Set Claims of all types
	for _, claim := range gs.Claims {
		k.SetClaim(ctx, claim)
	}

	// Set AccrualTimes of all types
	for _, accrualTime := range gs.AccrualTimes {
		k.SetRewardAccrualTime(
			ctx,
			accrualTime.ClaimType,
			accrualTime.CollateralType,
			accrualTime.PreviousAccumulationTime,
		)
	}

	// Legacy claims and indexes below

	// USDX Minting
	for _, claim := range gs.USDXMintingClaims {
		k.SetUSDXMintingClaim(ctx, claim)
	}
	for _, gat := range gs.USDXRewardState.AccumulationTimes {
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			panic(err.Error())
		}
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
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			panic(err.Error())
		}
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.HardSupplyRewardState.MultiRewardIndexes {
		k.SetHardSupplyRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}
	for _, gat := range gs.HardBorrowRewardState.AccumulationTimes {
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			panic(err.Error())
		}
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
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			panic(err.Error())
		}
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
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			panic(err.Error())
		}
		k.SetSwapRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.SwapRewardState.MultiRewardIndexes {
		k.SetSwapRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}

	// Savings
	for _, claim := range gs.SavingsClaims {
		k.SetSavingsClaim(ctx, claim)
	}
	for _, gat := range gs.SavingsRewardState.AccumulationTimes {
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			panic(err.Error())
		}
		k.SetSavingsRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.SavingsRewardState.MultiRewardIndexes {
		k.SetSavingsRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}

	// Earn
	for _, claim := range gs.EarnClaims {
		k.SetEarnClaim(ctx, claim)
	}
	for _, gat := range gs.EarnRewardState.AccumulationTimes {
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			panic(err.Error())
		}
		k.SetEarnRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range gs.EarnRewardState.MultiRewardIndexes {
		k.SetEarnRewardIndexes(ctx, mri.CollateralType, mri.RewardIndexes)
	}
}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)

	claims := k.GetAllClaims(ctx)
	accrualTimes := k.GetAllRewardAccrualTimes(ctx)

	usdxClaims := k.GetAllUSDXMintingClaims(ctx)
	usdxRewardState := getUSDXMintingGenesisRewardState(ctx, k)

	hardClaims := k.GetAllHardLiquidityProviderClaims(ctx)
	hardSupplyRewardState := getHardSupplyGenesisRewardState(ctx, k)
	hardBorrowRewardState := getHardBorrowGenesisRewardState(ctx, k)

	delegatorClaims := k.GetAllDelegatorClaims(ctx)
	delegatorRewardState := getDelegatorGenesisRewardState(ctx, k)

	swapClaims := k.GetAllSwapClaims(ctx)
	swapRewardState := getSwapGenesisRewardState(ctx, k)

	savingsClaims := k.GetAllSavingsClaims(ctx)
	savingsRewardState := getSavingsGenesisRewardState(ctx, k)

	earnClaims := k.GetAllEarnClaims(ctx)
	earnRewardState := getEarnGenesisRewardState(ctx, k)

	return types.NewGenesisState(
		params,
		// Reward states
		usdxRewardState, hardSupplyRewardState, hardBorrowRewardState, delegatorRewardState, swapRewardState, savingsRewardState, earnRewardState,
		// Claims
		claims, usdxClaims, hardClaims, delegatorClaims, swapClaims, savingsClaims, earnClaims,
		accrualTimes,
	)
}

func getUSDXMintingGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {
	var ats types.AccumulationTimes
	keeper.IterateUSDXMintingAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
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
	var ats types.AccumulationTimes
	keeper.IterateHardSupplyRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
	keeper.IterateHardSupplyRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getHardBorrowGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {
	var ats types.AccumulationTimes
	keeper.IterateHardBorrowRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
	keeper.IterateHardBorrowRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getDelegatorGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {
	var ats types.AccumulationTimes
	keeper.IterateDelegatorRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
	keeper.IterateDelegatorRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getSwapGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {
	var ats types.AccumulationTimes
	keeper.IterateSwapRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
	keeper.IterateSwapRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getSavingsGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {
	var ats types.AccumulationTimes
	keeper.IterateSavingsRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
	keeper.IterateSavingsRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func getEarnGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper) types.GenesisRewardState {
	var ats types.AccumulationTimes
	keeper.IterateEarnRewardAccrualTimes(ctx, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
	keeper.IterateEarnRewardIndexes(ctx, func(ctype string, indexes types.RewardIndexes) bool {
		mris = append(mris, types.NewMultiRewardIndex(ctype, indexes))
		return false
	})

	return types.NewGenesisRewardState(ats, mris)
}

func ValidateAccumulationTime(previousAccumulationTime, genesisTime time.Time) error {
	if previousAccumulationTime.Before(genesisTime.Add(-1 * EarliestValidAccumulationTime)) {
		return fmt.Errorf(
			"found accumulation time '%s' more than '%s' behind genesis time '%s'",
			previousAccumulationTime,
			EarliestValidAccumulationTime,
			genesisTime,
		)
	}
	return nil
}
