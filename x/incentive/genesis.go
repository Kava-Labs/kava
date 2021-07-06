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
		k.SetUSDXMintingRewardFactor(ctx, rp.CollateralType, sdk.ZeroDec())
	}

	for _, mrp := range gs.Params.HardSupplyRewardPeriods {
		newRewardIndexes := types.RewardIndexes{}
		for _, rc := range mrp.RewardsPerSecond {
			ri := types.NewRewardIndex(rc.Denom, sdk.ZeroDec())
			newRewardIndexes = append(newRewardIndexes, ri)
		}
		k.SetHardSupplyRewardIndexes(ctx, mrp.CollateralType, newRewardIndexes)
	}

	for _, mrp := range gs.Params.HardBorrowRewardPeriods {
		newRewardIndexes := types.RewardIndexes{}
		for _, rc := range mrp.RewardsPerSecond {
			ri := types.NewRewardIndex(rc.Denom, sdk.ZeroDec())
			newRewardIndexes = append(newRewardIndexes, ri)
		}
		k.SetHardBorrowRewardIndexes(ctx, mrp.CollateralType, newRewardIndexes)
	}

	for _, drp := range gs.Params.DelegatorRewardPeriods {
		newRewardIndexes := types.RewardIndexes{}
		for _, rc := range drp.RewardsPerSecond {
			ri := types.NewRewardIndex(rc.Denom, sdk.ZeroDec())
			newRewardIndexes = append(newRewardIndexes, ri)
		}
		k.SetDelegatorRewardIndexes(ctx, drp.CollateralType, newRewardIndexes)
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

	for _, gat := range gs.DelegatorAccumulationTimes {
		k.SetPreviousDelegatorRewardAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}

	for i, claim := range gs.USDXMintingClaims {
		for j, ri := range claim.RewardIndexes {
			if ri.RewardFactor != sdk.ZeroDec() {
				gs.USDXMintingClaims[i].RewardIndexes[j].RewardFactor = sdk.ZeroDec()
			}
		}
		k.SetUSDXMintingClaim(ctx, claim)
	}

	for i, claim := range gs.HardLiquidityProviderClaims {
		for j, mri := range claim.SupplyRewardIndexes {
			for k, ri := range mri.RewardIndexes {
				if ri.RewardFactor != sdk.ZeroDec() {
					gs.HardLiquidityProviderClaims[i].SupplyRewardIndexes[j].RewardIndexes[k].RewardFactor = sdk.ZeroDec()
				}
			}
		}
		for j, mri := range claim.BorrowRewardIndexes {
			for k, ri := range mri.RewardIndexes {
				if ri.RewardFactor != sdk.ZeroDec() {
					gs.HardLiquidityProviderClaims[i].BorrowRewardIndexes[j].RewardIndexes[k].RewardFactor = sdk.ZeroDec()
				}
			}
		}
		k.SetHardLiquidityProviderClaim(ctx, claim)
	}

	for i, claim := range gs.DelegatorClaims {
		for j, mri := range claim.RewardIndexes {
			for k, ri := range mri.RewardIndexes {
				if ri.RewardFactor != sdk.ZeroDec() {
					gs.DelegatorClaims[i].RewardIndexes[j].RewardIndexes[k].RewardFactor = sdk.ZeroDec()
				}
			}
		}
		k.SetDelegatorClaim(ctx, claim)
	}
}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)

	usdxClaims := k.GetAllUSDXMintingClaims(ctx)
	hardClaims := k.GetAllHardLiquidityProviderClaims(ctx)
	delegatorClaims := k.GetAllDelegatorClaims(ctx)

	synchronizedUsdxClaims := types.USDXMintingClaims{}
	synchronizedHardClaims := types.HardLiquidityProviderClaims{}
	synchronizedDelegatorClaims := types.DelegatorClaims{}

	for _, usdxClaim := range usdxClaims {
		claim, err := k.SynchronizeUSDXMintingClaim(ctx, usdxClaim)
		if err != nil {
			panic(err)
		}
		for i := range claim.RewardIndexes {
			claim.RewardIndexes[i].RewardFactor = sdk.ZeroDec()
		}
		synchronizedUsdxClaims = append(synchronizedUsdxClaims, claim)
	}

	for _, hardClaim := range hardClaims {
		k.SynchronizeHardLiquidityProviderClaim(ctx, hardClaim.Owner)
		claim, found := k.GetHardLiquidityProviderClaim(ctx, hardClaim.Owner)
		if !found {
			panic("hard liquidity provider claim should always be found after synchronization")
		}
		for i, bri := range claim.BorrowRewardIndexes {
			for j := range bri.RewardIndexes {
				claim.BorrowRewardIndexes[i].RewardIndexes[j].RewardFactor = sdk.ZeroDec()
			}
		}
		for i, sri := range claim.SupplyRewardIndexes {
			for j := range sri.RewardIndexes {
				claim.SupplyRewardIndexes[i].RewardIndexes[j].RewardFactor = sdk.ZeroDec()
			}
		}
		synchronizedHardClaims = append(synchronizedHardClaims, claim)
	}

	for _, delegatorClaim := range delegatorClaims {
		claim, err := k.SynchronizeDelegatorClaim(ctx, delegatorClaim)
		if err != nil {
			panic(err)
		}
		for i, ri := range claim.RewardIndexes {
			for j := range ri.RewardIndexes {
				claim.RewardIndexes[i].RewardIndexes[j].RewardFactor = sdk.ZeroDec()
			}
		}
		synchronizedDelegatorClaims = append(synchronizedDelegatorClaims, delegatorClaim)
	}

	var usdxMintingGats GenesisAccumulationTimes
	for _, rp := range params.USDXMintingRewardPeriods {
		pat, found := k.GetPreviousUSDXMintingAccrualTime(ctx, rp.CollateralType)
		if !found {
			panic(fmt.Sprintf("expected previous usdx minting reward accrual time to be set in state for %s", rp.CollateralType))
		}
		gat := types.NewGenesisAccumulationTime(rp.CollateralType, pat)
		usdxMintingGats = append(usdxMintingGats, gat)
	}

	var hardSupplyGats GenesisAccumulationTimes
	for _, rp := range params.HardSupplyRewardPeriods {
		pat, found := k.GetPreviousHardSupplyRewardAccrualTime(ctx, rp.CollateralType)
		if !found {
			panic(fmt.Sprintf("expected previous hard supply reward accrual time to be set in state for %s", rp.CollateralType))
		}
		gat := types.NewGenesisAccumulationTime(rp.CollateralType, pat)
		hardSupplyGats = append(hardSupplyGats, gat)
	}

	var hardBorrowGats GenesisAccumulationTimes
	for _, rp := range params.HardBorrowRewardPeriods {
		pat, found := k.GetPreviousHardBorrowRewardAccrualTime(ctx, rp.CollateralType)
		if !found {
			panic(fmt.Sprintf("expected previous hard borrow reward accrual time to be set in state for %s", rp.CollateralType))
		}
		gat := types.NewGenesisAccumulationTime(rp.CollateralType, pat)
		hardBorrowGats = append(hardBorrowGats, gat)
	}

	var delegatorGats GenesisAccumulationTimes
	for _, rp := range params.DelegatorRewardPeriods {
		pat, found := k.GetPreviousDelegatorRewardAccrualTime(ctx, rp.CollateralType)
		if !found {
			panic(fmt.Sprintf("expected previous delegator reward accrual time to be set in state for %s", rp.CollateralType))
		}
		gat := types.NewGenesisAccumulationTime(rp.CollateralType, pat)
		delegatorGats = append(delegatorGats, gat)
	}

	return types.NewGenesisState(params, usdxMintingGats, hardSupplyGats, hardBorrowGats,
		delegatorGats, synchronizedUsdxClaims, synchronizedHardClaims, synchronizedDelegatorClaims)
}
