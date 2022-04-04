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

	// USDX Minting
	for _, claim := range gs.USDXMintingClaims {
		k.SetUSDXMintingClaim(ctx, claim)
	}
	if err := setRewardState(ctx, k, types.USDXMinting, gs.USDXRewardState); err != nil {
		panic(err)
	}

	// Hard Supply / Borrow
	for _, claim := range gs.HardLiquidityProviderClaims {
		k.SetHardLiquidityProviderClaim(ctx, claim)
	}
	if err := setRewardState(ctx, k, types.HardBorrow, gs.HardBorrowRewardState); err != nil {
		panic(err)
	}
	if err := setRewardState(ctx, k, types.HardSupply, gs.HardSupplyRewardState); err != nil {
		panic(err)
	}

	// Delegator
	for _, claim := range gs.DelegatorClaims {
		k.SetDelegatorClaim(ctx, claim)
	}
	if err := setRewardState(ctx, k, types.Delegator, gs.DelegatorRewardState); err != nil {
		panic(err)
	}

	// Swap
	for _, claim := range gs.SwapClaims {
		k.SetSwapClaim(ctx, claim)
	}
	if err := setRewardState(ctx, k, types.Swap, gs.SwapRewardState); err != nil {
		panic(err)
	}

	// Savings
	for _, claim := range gs.SavingsClaims {
		k.SetSavingsClaim(ctx, claim)
	}
	if err := setRewardState(ctx, k, types.Savings, gs.SavingsRewardState); err != nil {
		panic(err)
	}
}

func setRewardState(ctx sdk.Context, k keeper.Keeper, sourceID types.SourceID, rewardState types.GenesisRewardState) error {
	for _, gat := range rewardState.AccumulationTimes {
		if err := ValidateAccumulationTime(gat.PreviousAccumulationTime, ctx.BlockTime()); err != nil {
			return err
		}
		k.SetLastAccrual(ctx, sourceID, gat.CollateralType, gat.PreviousAccumulationTime)
	}
	for _, mri := range rewardState.MultiRewardIndexes {
		k.SetRewardIndexes(ctx, sourceID, mri.CollateralType, mri.RewardIndexes)
	}
	return nil
}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)

	usdxClaims := k.GetAllUSDXMintingClaims(ctx)
	usdxRewardState := getGenesisRewardState(ctx, k)

	hardClaims := k.GetAllHardLiquidityProviderClaims(ctx)
	hardSupplyRewardState := getGenesisRewardState(ctx, k, types.HardSupply)
	hardBorrowRewardState := getGenesisRewardState(ctx, k, types.HardBorrow)

	delegatorClaims := k.GetAllDelegatorClaims(ctx)
	delegatorRewardState := getGenesisRewardState(ctx, k, types.Delegator)

	swapClaims := k.GetAllSwapClaims(ctx)
	swapRewardState := getGenesisRewardState(ctx, k, types.Swap)

	savingsClaims := k.GetAllSavingsClaims(ctx)
	savingsRewardState := getGenesisRewardState(ctx, k, types.Savings)

	return types.NewGenesisState(
		params,
		usdxRewardState, hardSupplyRewardState, hardBorrowRewardState, delegatorRewardState, swapRewardState,
		savingsRewardState, usdxClaims, hardClaims, delegatorClaims, swapClaims, savingsClaims,
	)
}

func getGenesisRewardState(ctx sdk.Context, keeper keeper.Keeper, sourceID types.SourceID) types.GenesisRewardState {
	var ats types.AccumulationTimes
	keeper.IterateLastAccruals(ctx, sourceID, func(ctype string, accTime time.Time) bool {
		ats = append(ats, types.NewAccumulationTime(ctype, accTime))
		return false
	})

	var mris types.MultiRewardIndexes
	keeper.IterateRewardIndexes(ctx, sourceID, func(ctype string, indexes types.RewardIndexes) bool {
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
