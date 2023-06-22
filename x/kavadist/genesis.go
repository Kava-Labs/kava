package kavadist

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, accountKeeper types.AccountKeeper, gs *types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	k.SetParams(ctx, gs.Params)

	// only set the previous block time if it's different than default
	if !gs.PreviousBlockTime.Equal(types.DefaultPreviousBlockTime) {
		k.SetPreviousBlockTime(ctx, gs.PreviousBlockTime)
	}

	// check if the module account exists
	moduleAcc := accountKeeper.GetModuleAccount(ctx, types.KavaDistMacc)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.KavaDistMacc))
	}

	// check if the fund account exists
	fundModuleAcc := accountKeeper.GetModuleAccount(ctx, types.FundModuleAccount)
	if fundModuleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.FundModuleAccount))
	}
}

// ExportGenesis export genesis state for cdp module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = types.DefaultPreviousBlockTime
	}
	return &types.GenesisState{
		Params:            params,
		PreviousBlockTime: previousBlockTime,
	}
}
