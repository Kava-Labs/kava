package validatorvesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/types"
)

// InitGenesis stores the account address of each ValidatorVestingAccount in the validator vesting keeper, for faster lookup.
// CONTRACT: Accounts created by the account keeper must have already been initialized/created by AccountKeeper
func InitGenesis(ctx sdk.Context, keeper Keeper, accountKeeper types.AccountKeeper, data GenesisState) {
	data.Accounts = auth.SanitizeGenesisAccounts(data.Accounts)
	for _, a := range data.Accounts {
		vv, ok := a.(ValidatorVestingAccount)
		if ok {
			keeper.SetValidatorVestingAccountKey(ctx, vv.Address)
		}
	}
}
