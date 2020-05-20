package v038

import (
	"fmt"

	v038auth "github.com/cosmos/cosmos-sdk/x/auth"
	v038authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
)

func Migrate(oldGenState v18de63.GenesisState) v038auth.GenesisState {

	// old and new types are almost identical, just different (un)marshalJSON methods
	var newAccounts v038authexported.GenesisAccounts
	for _, account := range oldGenState.Accounts {
		switch acc := account.(type) {
		case *v18de63.BaseAccount:
			ba := v038auth.BaseAccount(*acc)
			newAccounts = append(newAccounts, v038authexported.GenesisAccount(&ba))
		// TODO
		// case *v18de63.ModuleAccount:
		// case *v18de63.ContinuousVestingAccount:
		// case *v18de63.PeriodicAccount:
		// case *v18de63.VestingModuleAccount:
		default:
			// TODO
			fmt.Println("different account type: ", acc)
		}
	}
	return v038auth.GenesisState{
		Params:   v038auth.Params(oldGenState.Params),
		Accounts: newAccounts,
	}
}
