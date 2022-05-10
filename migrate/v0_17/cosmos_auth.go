package v0_17

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	v040vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/kava-labs/kava/migrate/utils"
)

// MigrateAuthV040 resets all periodic vesting accounts for a given
// v40 cosmos auth module genesis state, returning a copy of the original state where all
// periodic vesting accounts have been zeroed out.
func MigrateAuthV040(authGenState v040auth.GenesisState, genesisTime time.Time, ctx client.Context) *v040auth.GenesisState {
	anyAccounts := make([]*codectypes.Any, len(authGenState.Accounts))
	for i, anyAcc := range authGenState.Accounts {
		// Only need to make modifications to vesting accounts
		if anyAcc.TypeUrl != "/cosmos.vesting.v1beta1.PeriodicVestingAccount" {
			anyAccounts[i] = anyAcc
			continue
		}
		var acc v040auth.GenesisAccount
		if err := ctx.InterfaceRegistry.UnpackAny(anyAcc, &acc); err != nil {
			panic(err)
		}
		if vacc, ok := acc.(*v040vesting.PeriodicVestingAccount); ok {
			vestingPeriods := make([]v040vesting.Period, len(vacc.VestingPeriods))
			for j, period := range vacc.VestingPeriods {
				vestingPeriods[j] = v040vesting.Period{
					Length: period.Length,
					Amount: period.Amount,
				}
			}
			vacc := v040vesting.PeriodicVestingAccount{
				BaseVestingAccount: vacc.BaseVestingAccount,
				StartTime:          vacc.StartTime,
				VestingPeriods:     vestingPeriods,
			}

			utils.ResetPeriodicVestingAccount(&vacc, genesisTime)

			// If periodic vesting account has zero periods, convert back
			// to a base account
			if genesisTime.Unix() >= vacc.EndTime {
				any, err := codectypes.NewAnyWithValue(vacc.BaseVestingAccount.BaseAccount)
				if err != nil {
					panic(err)
				}
				anyAccounts[i] = any
				continue
			}
			// Convert back to any
			any, err := codectypes.NewAnyWithValue(&vacc)
			if err != nil {
				panic(err)
			}
			anyAccounts[i] = any
		}
	}

	return &v040auth.GenesisState{
		Params:   authGenState.Params,
		Accounts: anyAccounts,
	}
}
