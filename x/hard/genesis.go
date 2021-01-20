package hard

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hard/types"
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

	for _, mm := range gs.Params.MoneyMarkets {
		k.SetMoneyMarket(ctx, mm.Denom, mm)
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

	// check if the module account exists
	LiquidatorModuleAcc := supplyKeeper.GetModuleAccount(ctx, LiquidatorAccount)
	if LiquidatorModuleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", LiquidatorAccount))
	}

}

// ExportGenesis export genesis state for hard module
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = DefaultPreviousBlockTime
	}
	return NewGenesisState(params, previousBlockTime)
}
