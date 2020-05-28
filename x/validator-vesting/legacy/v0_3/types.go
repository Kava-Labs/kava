package v18de63

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authvestingtypes "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63" // combined auth and vesting packages
)

// VestingProgress tracks the status of each vesting period
type VestingProgress struct {
	PeriodComplete    bool `json:"period_complete" yaml:"period_complete"`
	VestingSuccessful bool `json:"vesting_successful" yaml:"vesting_successful"`
}

// CurrentPeriodProgress tracks the progress of the current vesting period
type CurrentPeriodProgress struct {
	MissedBlocks int64 `json:"missed_blocks" yaml:"missed_blocks"`
	TotalBlocks  int64 `json:"total_blocks" yaml:"total_blocks"`
}

// ValidatorVestingAccount implements the VestingAccount interface. It
// conditionally vests by unlocking coins during each specified period, provided
// that the validator address has validated at least **SigningThreshold** blocks during
// the previous vesting period. The signing threshold takes values 0 to 100 are represents the
// percentage of blocks that must be signed each period for the vesting to complete successfully.
// If the validator has not signed at least the threshold percentage of blocks during a period,
// the coins are returned to the return address, or burned if the return address is null.
type ValidatorVestingAccount struct {
	*authvestingtypes.PeriodicVestingAccount
	ValidatorAddress       sdk.ConsAddress       `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress        `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64                 `json:"signing_threshold" yaml:"signing_threshold"`
	CurrentPeriodProgress  CurrentPeriodProgress `json:"current_period_progress" yaml:"current_period_progress"`
	VestingPeriodProgress  []VestingProgress     `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins             `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}
