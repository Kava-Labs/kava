package v0_13

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	tmtime "github.com/tendermint/tendermint/types/time"
)

// Assert ValidatorVestingAccount implements the vestexported.VestingAccount interface
// Assert ValidatorVestingAccount implements the authexported.GenesisAccount interface
var _ vestexported.VestingAccount = (*ValidatorVestingAccount)(nil)
var _ authexported.GenesisAccount = (*ValidatorVestingAccount)(nil)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	PreviousBlockTime time.Time `json:"previous_block_time" yaml:"previous_block_time"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(prevBlockTime time.Time) GenesisState {
	return GenesisState{
		PreviousBlockTime: prevBlockTime,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(tmtime.Canonical(time.Unix(0, 0)))
}

// ValidateGenesis returns nil because accounts are validated by auth
func ValidateGenesis(data GenesisState) error {
	if data.PreviousBlockTime.IsZero() {
		return errors.New("previous block time cannot be zero")
	}
	return nil
}

// ValidatorVestingAccount implements the VestingAccount interface. It
// conditionally vests by unlocking coins during each specified period, provided
// that the validator address has validated at least **SigningThreshold** blocks during
// the previous vesting period. The signing threshold takes values 0 to 100 are represents the
// percentage of blocks that must be signed each period for the vesting to complete successfully.
// If the validator has not signed at least the threshold percentage of blocks during a period,
// the coins are returned to the return address, or burned if the return address is null.
type ValidatorVestingAccount struct {
	*vestingtypes.PeriodicVestingAccount
	ValidatorAddress       sdk.ConsAddress       `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress        `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64                 `json:"signing_threshold" yaml:"signing_threshold"`
	CurrentPeriodProgress  CurrentPeriodProgress `json:"current_period_progress" yaml:"current_period_progress"`
	VestingPeriodProgress  []VestingProgress     `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins             `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}

// CurrentPeriodProgress tracks the progress of the current vesting period
type CurrentPeriodProgress struct {
	MissedBlocks int64 `json:"missed_blocks" yaml:"missed_blocks"`
	TotalBlocks  int64 `json:"total_blocks" yaml:"total_blocks"`
}

// VestingProgress tracks the status of each vesting period
type VestingProgress struct {
	PeriodComplete    bool `json:"period_complete" yaml:"period_complete"`
	VestingSuccessful bool `json:"vesting_successful" yaml:"vesting_successful"`
}
