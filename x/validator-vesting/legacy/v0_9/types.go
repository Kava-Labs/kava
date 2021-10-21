package v0_9

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/crypto"

	authtypes "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/auth"
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
	*authtypes.PeriodicVestingAccount
	ValidatorAddress       sdk.ConsAddress       `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress        `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64                 `json:"signing_threshold" yaml:"signing_threshold"`
	CurrentPeriodProgress  CurrentPeriodProgress `json:"current_period_progress" yaml:"current_period_progress"`
	VestingPeriodProgress  []VestingProgress     `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins             `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}

type validatorVestingAccountPretty struct {
	Address                sdk.AccAddress        `json:"address" yaml:"address"`
	Coins                  sdk.Coins             `json:"coins" yaml:"coins"`
	PubKey                 string                `json:"public_key" yaml:"public_key"`
	AccountNumber          uint64                `json:"account_number" yaml:"account_number"`
	Sequence               uint64                `json:"sequence" yaml:"sequence"`
	OriginalVesting        sdk.Coins             `json:"original_vesting" yaml:"original_vesting"`
	DelegatedFree          sdk.Coins             `json:"delegated_free" yaml:"delegated_free"`
	DelegatedVesting       sdk.Coins             `json:"delegated_vesting" yaml:"delegated_vesting"`
	EndTime                int64                 `json:"end_time" yaml:"end_time"`
	StartTime              int64                 `json:"start_time" yaml:"start_time"`
	VestingPeriods         authtypes.Periods     `json:"vesting_periods" yaml:"vesting_periods"`
	ValidatorAddress       sdk.ConsAddress       `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress        `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64                 `json:"signing_threshold" yaml:"signing_threshold"`
	CurrentPeriodProgress  CurrentPeriodProgress `json:"current_period_progress" yaml:"current_period_progress"`
	VestingPeriodProgress  []VestingProgress     `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins             `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}

// UnmarshalJSON unmarshals raw JSON bytes into a PeriodicVestingAccount.
func (vva *ValidatorVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias validatorVestingAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	var (
		pk  crypto.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}
	}

	ba := authtypes.NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence)
	bva := &authtypes.BaseVestingAccount{
		BaseAccount:      ba,
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	pva := authtypes.NewPeriodicVestingAccountRaw(bva, alias.StartTime, alias.VestingPeriods)
	vva.PeriodicVestingAccount = pva
	vva.ValidatorAddress = alias.ValidatorAddress
	vva.ReturnAddress = alias.ReturnAddress
	vva.SigningThreshold = alias.SigningThreshold
	vva.CurrentPeriodProgress = alias.CurrentPeriodProgress
	vva.VestingPeriodProgress = alias.VestingPeriodProgress
	vva.DebtAfterFailedVesting = alias.DebtAfterFailedVesting
	return nil
}

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&ValidatorVestingAccount{}, "cosmos-sdk/ValidatorVestingAccount", nil)
}
