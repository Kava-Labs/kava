package validatorvesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

// InitGenesis stores the account address of each ValidatorVestingAccount in the validator vesting keeper, for faster lookup.
// CONTRACT: Accounts must have already been initialized/created by AccountKeeper
func InitGenesis(ctx sdk.Context, keeper Keeper, accountKeeper types.AccountKeeper, data GenesisState) {

	accounts := accountKeeper.GetAllAccounts(ctx)
	for _, a := range accounts {
		vv, ok := a.(*ValidatorVestingAccount)
		if ok {
			keeper.SetValidatorVestingAccountKey(ctx, vv.Address)
		}
	}
	keeper.SetPreviousBlockTime(ctx, data.PreviousBlockTime)
}

// ExportGenesis returns empty genesis state because auth exports all the genesis state we need.
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	prevBlockTime, found := keeper.GetPreviousBlockTime(ctx)
	if !found {
		prevBlockTime = ctx.BlockTime()
	}
	return NewGenesisState(prevBlockTime)
}
