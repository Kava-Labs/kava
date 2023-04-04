package utils

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v040vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// ResetPeriodicVestingAccount resets a periodic vesting account to a new start
// time. The account is modified in place, and vesting periods before the new
// start time are removed from the account.
func ResetPeriodicVestingAccount(vacc *v040vesting.PeriodicVestingAccount, startTime time.Time) {
	currentPeriod := vacc.StartTime

	newOriginalVesting := sdk.Coins{}
	newStartTime := startTime.Unix()
	newPeriods := v040vesting.Periods{}

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

			newOriginalVesting = newOriginalVesting.Add(period.Amount...)
			newPeriods = append(newPeriods, period)
		}
	}

	// If the new original vesting amount is less than the delegated vesting amount, set delegated vesting
	// to the new original vesting amount, and add the difference to the delegated free amount
	for _, delegatedVestingCoin := range vacc.DelegatedVesting {
		newDelegatedVestingCoin := sdk.NewCoin(delegatedVestingCoin.Denom, sdk.MinInt(delegatedVestingCoin.Amount, newOriginalVesting.AmountOf(delegatedVestingCoin.Denom)))
		delegationAdjustment := delegatedVestingCoin.Sub(newDelegatedVestingCoin)

		if !delegationAdjustment.IsZero() {
			vacc.DelegatedVesting = vacc.DelegatedVesting.Sub(delegationAdjustment)
			vacc.DelegatedFree = vacc.DelegatedFree.Add(delegationAdjustment)
		}
	}

	// update vesting account
	vacc.StartTime = newStartTime
	vacc.OriginalVesting = newOriginalVesting
	vacc.VestingPeriods = newPeriods

	// ensure end time is >= start time
	if vacc.StartTime >= vacc.EndTime {
		vacc.EndTime = vacc.StartTime
	}
}
