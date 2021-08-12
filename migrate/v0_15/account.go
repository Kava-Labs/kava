package v0_15

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

// MigrateAccount removes old vesting periods from periodic vesting accounts
// and converts any periodic vesting account with zero periods back to a base account
func MigrateAccount(acc authexported.GenesisAccount, genesisTime time.Time) authexported.GenesisAccount {
	if vacc, ok := acc.(*vesting.PeriodicVestingAccount); ok {
		ResetPeriodicVestingAccount(vacc, genesisTime)

		if genesisTime.Unix() >= vacc.EndTime {
			return vacc.BaseVestingAccount.BaseAccount
		}

		return vacc
	}

	return acc
}

// ResetPeriodicVestingAccount resets a periodic vesting account to a new start time.  The account is
// modified in place, and vesting periods before the new start time are removed from the account.
func ResetPeriodicVestingAccount(vacc *vesting.PeriodicVestingAccount, startTime time.Time) {
	currentPeriod := vacc.StartTime

	newOriginalVesting := sdk.Coins{}
	newStartTime := startTime.Unix()
	newEndTime := newStartTime
	newPeriods := vesting.Periods{}

	for _, period := range vacc.VestingPeriods {
		currentPeriod = currentPeriod + period.Length

		// Periods less than the newStartTime are still vesting,
		// so adjust their length and add them to the newPeriods
		if newStartTime < currentPeriod {
			// adjust the length of the first vesting period
			// to be relative to the new start time
			if len(newPeriods) == 0 {
				period.Length = currentPeriod - newStartTime
			}

			newEndTime = newEndTime + period.Length
			newOriginalVesting = newOriginalVesting.Add(period.Amount...)

			newPeriods = append(newPeriods, period)
		}
	}

	// In order to preserve the spendable amount of the account, we must drop
	// the vesting funds if the start and end time are equal.
	if newStartTime == newEndTime {
		newOriginalVesting = sdk.Coins{}
		newPeriods = vesting.Periods{}
	}

	// If the new original vesting amount is less than the delegated vesting amount, set delegated vesting
	// to the new original vesting amount, and add the difference to the delegated free amount
	for _, delegatedVestingCoin := range vacc.DelegatedVesting {
		newDelegatedVestingCoin := sdk.NewCoin(delegatedVestingCoin.Denom, sdk.MinInt(delegatedVestingCoin.Amount, newOriginalVesting.AmountOf(delegatedVestingCoin.Denom)))
		delegationAdjustment := delegatedVestingCoin.Sub(newDelegatedVestingCoin)

		if !delegationAdjustment.IsZero() {
			vacc.DelegatedVesting = vacc.DelegatedVesting.Sub(sdk.NewCoins(delegationAdjustment))
			vacc.DelegatedFree = vacc.DelegatedFree.Add(delegationAdjustment)
		}
	}

	vacc.StartTime = newStartTime
	vacc.EndTime = newEndTime
	vacc.OriginalVesting = newOriginalVesting
	vacc.VestingPeriods = newPeriods
}
