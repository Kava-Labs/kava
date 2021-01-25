package harvest

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/harvest/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, k Keeper, supplyKeeper types.SupplyKeeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	k.SetParams(ctx, gs.Params)

	// only set the previous block time if it's different than default
	if !gs.PreviousBlockTime.Equal(DefaultPreviousBlockTime) {
		k.SetPreviousBlockTime(ctx, gs.PreviousBlockTime)
	}

	for _, pdt := range gs.PreviousDistributionTimes {
		if !pdt.PreviousDistributionTime.Equal(DefaultPreviousBlockTime) {
			k.SetPreviousDelegationDistribution(ctx, pdt.PreviousDistributionTime, pdt.Denom)
		}
	}

	// check if the module account exists
	LPModuleAcc := supplyKeeper.GetModuleAccount(ctx, LPAccount)
	if LPModuleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", LPAccount))
	}

	// check if the module account exists
	DelegatorModuleAcc := supplyKeeper.GetModuleAccount(ctx, DelegatorAccount)
	if DelegatorModuleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", DelegatorAccount))
	}

	// check if the module account exists
	DepositModuleAccount := supplyKeeper.GetModuleAccount(ctx, ModuleAccountName)
	if DepositModuleAccount == nil {
		panic(fmt.Sprintf("%s module account has not been set", DepositModuleAccount))
	}

	for _, dep := range gs.Deposits {
		k.SetDeposit(ctx, dep)
	}

	for _, claim := range gs.Claims {
		k.SetClaim(ctx, claim)
	}

}

// ExportGenesis export genesis state for harvest module
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = DefaultPreviousBlockTime
	}
	previousDistTimes := GenesisDistributionTimes{}
	for _, dds := range params.DelegatorDistributionSchedules {
		previousDistTime, found := k.GetPreviousDelegatorDistribution(ctx, dds.DistributionSchedule.DepositDenom)
		if found {
			previousDistTimes = append(previousDistTimes, GenesisDistributionTime{PreviousDistributionTime: previousDistTime, Denom: dds.DistributionSchedule.DepositDenom})
		}
	}
	deposits := types.Deposits{}
	k.IterateDeposits(ctx, func(deposit types.Deposit) (stop bool) {
		deposits = append(deposits, deposit)
		return false
	})

	claims := types.Claims{}
	k.IterateClaims(ctx, func(claim types.Claim) (stop bool) {
		claims = append(claims, claim)
		return false
	})
	return NewGenesisState(params, previousBlockTime, previousDistTimes, deposits, claims)
}
