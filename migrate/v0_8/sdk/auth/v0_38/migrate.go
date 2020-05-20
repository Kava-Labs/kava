package v038

import (
	v038auth "github.com/cosmos/cosmos-sdk/x/auth"
	v038authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
)

func Migrate(oldGenState v18de63.GenesisState) v038auth.GenesisState {

	// old and new types are almost identical, just different (un)marshalJSON methods
	return v038auth.GenesisState{
		Params:   v038auth.Params(oldGenState.Params),
		Accounts: v038authexported.GenesisAccounts(oldGenState.Accounts),
	}
}
