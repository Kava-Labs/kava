package v0_15

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
)

// ValidatorVestingAccount implements the VestingAccount interface. It
// conditionally vests by unlocking coins during each specified period, provided
// that the validator address has validated at least **SigningThreshold** blocks during
// the previous vesting period. The signing threshold takes values 0 to 100 are represents the
// percentage of blocks that must be signed each period for the vesting to complete successfully.
// If the validator has not signed at least the threshold percentage of blocks during a period,
// the coins are returned to the return address, or burned if the return address is null.
type ValidatorVestingAccount struct {
	*v039auth.PeriodicVestingAccount
	ValidatorAddress       sdk.ConsAddress       `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress        `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64                 `json:"signing_threshold" yaml:"signing_threshold"`
	CurrentPeriodProgress  CurrentPeriodProgress `json:"current_period_progress" yaml:"current_period_progress"`
	VestingPeriodProgress  []VestingProgress     `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins             `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}

type validatorVestingAccountJSON struct {
	Address          sdk.AccAddress     `json:"address" yaml:"address"`
	Coins            sdk.Coins          `json:"coins" yaml:"coins"`
	PubKey           cryptotypes.PubKey `json:"public_key" yaml:"public_key"`
	AccountNumber    uint64             `json:"account_number" yaml:"account_number"`
	Sequence         uint64             `json:"sequence" yaml:"sequence"`
	OriginalVesting  sdk.Coins          `json:"original_vesting" yaml:"original_vesting"`
	DelegatedFree    sdk.Coins          `json:"delegated_free" yaml:"delegated_free"`
	DelegatedVesting sdk.Coins          `json:"delegated_vesting" yaml:"delegated_vesting"`
	EndTime          int64              `json:"end_time" yaml:"end_time"`

	// non-base vesting account fields
	StartTime              int64                 `json:"start_time" yaml:"start_time"`
	VestingPeriods         v039auth.Periods      `json:"vesting_periods" yaml:"vesting_periods"`
	ValidatorAddress       sdk.ConsAddress       `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress        `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64                 `json:"signing_threshold" yaml:"signing_threshold"`
	CurrentPeriodProgress  CurrentPeriodProgress `json:"current_period_progress" yaml:"current_period_progress"`
	VestingPeriodProgress  []VestingProgress     `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins             `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}

// NewPeriodicVestingAccountRaw creates a new PeriodicVestingAccount object from BaseVestingAccount
func NewPeriodicVestingAccountRaw(bva *v039auth.BaseVestingAccount, startTime int64, periods v039auth.Periods) *v039auth.PeriodicVestingAccount {
	return &v039auth.PeriodicVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}
}

// UnmarshalJSON unmarshals raw JSON bytes into a ValidatorVestingAccount.
func (vva *ValidatorVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias validatorVestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	ba := v039auth.NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence)
	bva := &v039auth.BaseVestingAccount{
		BaseAccount:      ba,
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	pva := NewPeriodicVestingAccountRaw(bva, alias.StartTime, alias.VestingPeriods)
	vva.PeriodicVestingAccount = pva
	vva.ValidatorAddress = alias.ValidatorAddress
	vva.ReturnAddress = alias.ReturnAddress
	vva.SigningThreshold = alias.SigningThreshold
	vva.CurrentPeriodProgress = alias.CurrentPeriodProgress
	vva.VestingPeriodProgress = alias.VestingPeriodProgress
	vva.DebtAfterFailedVesting = alias.DebtAfterFailedVesting
	return nil
}

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

// Period defines a length of time and amount of coins that will vest
type Period struct {
	Length int64     `json:"length" yaml:"length"` // length of the period, in seconds
	Amount sdk.Coins `json:"amount" yaml:"amount"` // amount of coins vesting during this period
}

// Periods stores all vesting periods passed as part of a PeriodicVestingAccount
type Periods []Period

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&ValidatorVestingAccount{}, "cosmos-sdk/ValidatorVestingAccount", nil)
}
