package hvt

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hvt/types"
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

	// check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, KavaDistMacc)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", KavaDistMacc))
	}

}

// ExportGenesis export genesis state for cdp module
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = DefaultPreviousBlockTime
	}
	return NewGenesisState(params, previousBlockTime)
}
