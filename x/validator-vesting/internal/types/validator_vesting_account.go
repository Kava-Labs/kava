package types

import (
	"errors"
	"time"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// Assert ValidatorVestingAccount implements the vestexported.VestingAccount interface
// Assert ValidatorVestingAccount implements the authexported.GenesisAccount interface
var _ vestexported.VestingAccount = (*ValidatorVestingAccount)(nil)
var _ authexported.GenesisAccount = (*ValidatorVestingAccount)(nil)

// Register the ValidatorVestingAccount type on the auth module codec
func init() {
	authtypes.RegisterAccountTypeCodec(&ValidatorVestingAccount{}, "cosmos-sdk/ValidatorVestingAccount")
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
	ValidatorAddress       sdk.ConsAddress `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress  `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64           `json:"signing_threshold" yaml:"signing_threshold"`
	MissingSignCount       []int64         `json:"missing_sign_count" yaml:"missing_sign_count"`
	VestingPeriodProgress  [][]int         `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins       `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}

// NewValidatorVestingAccountRaw creates a new ValidatorVestingAccount object from BaseVestingAccount
func NewValidatorVestingAccountRaw(bva *vestingtypes.BaseVestingAccount,
	startTime int64, periods vestingtypes.Periods, validatorAddress sdk.ConsAddress, returnAddress sdk.AccAddress, signingThreshold int64) *ValidatorVestingAccount {

	pva := &vestingtypes.PeriodicVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}
	var vestingPeriodProgress = make([][]int, len(periods))
	for i := range vestingPeriodProgress {
		vestingPeriodProgress[i] = make([]int, 2)
	}

	return &ValidatorVestingAccount{
		PeriodicVestingAccount: pva,
		ValidatorAddress:       validatorAddress,
		ReturnAddress:          returnAddress,
		SigningThreshold:       signingThreshold,
		MissingSignCount:       []int64{0, 0},
		VestingPeriodProgress:  vestingPeriodProgress,
		DebtAfterFailedVesting: sdk.NewCoins(),
	}
}

// NewValidatorVestingAccount creates a ValidatorVestingAccount object from a BaseAccount
func NewValidatorVestingAccount(baseAcc *authtypes.BaseAccount, startTime int64, periods vestingtypes.Periods, validatorAddress sdk.ConsAddress, returnAddress sdk.AccAddress, signingThreshold int64) *ValidatorVestingAccount {

	endTime := startTime
	for _, p := range periods {
		endTime += p.Length
	}
	baseVestingAcc := &vestingtypes.BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: baseAcc.Coins,
		EndTime:         endTime,
	}
	pva := &vestingtypes.PeriodicVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}
	var vestingPeriodProgress = make([][]int, len(periods))
	for i := range vestingPeriodProgress {
		vestingPeriodProgress[i] = make([]int, 2)
	}

	debt := sdk.NewCoins()

	return &ValidatorVestingAccount{
		PeriodicVestingAccount: pva,
		ValidatorAddress:       validatorAddress,
		ReturnAddress:          returnAddress,
		SigningThreshold:       signingThreshold,
		MissingSignCount:       []int64{0, 0},
		VestingPeriodProgress:  vestingPeriodProgress,
		DebtAfterFailedVesting: debt,
	}
}

// GetVestedCoins returns the total number of vested coins.
func (vva ValidatorVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins
	if blockTime.Unix() <= vva.StartTime {
		return vestedCoins
	}
	currentPeriodStartTime := vva.StartTime
	numberPeriods := len(vva.VestingPeriods)
	for i := 0; i < numberPeriods; i++ {
		x := blockTime.Unix() - currentPeriodStartTime
		if x >= vva.VestingPeriods[i].Length {
			vestingComplete := vva.VestingPeriodProgress[i][0] == 1
			if vestingComplete {
				vestedCoins = vestedCoins.Add(vva.VestingPeriods[i].Amount)
			}
			currentPeriodStartTime += vva.VestingPeriods[i].Length
		} else {
			break
		}
	}
	return vestedCoins

}

// GetFailedVestedCoins returns the total number of coins for which the vesting period has passed but the vesting threshold was not met.
func (vva ValidatorVestingAccount) GetFailedVestedCoins() sdk.Coins {
	var failedVestedCoins sdk.Coins
	numberPeriods := len(vva.VestingPeriods)
	for i := 0; i < numberPeriods; i++ {
		if vva.VestingPeriodProgress[i][0] == 1 {
			vestedFailure := vva.VestingPeriodProgress[i][1] == 0
			if vestedFailure {
				failedVestedCoins = failedVestedCoins.Add(vva.VestingPeriods[i].Amount)
			}
		} else {
			break
		}
	}
	return failedVestedCoins
}

// GetVestingCoins returns the total number of vesting coins. For validator vesting accounts, this excludes coins for which the vesting period has passed, but the vesting threshold was not met.
func (vva ValidatorVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return vva.OriginalVesting.Sub(vva.GetVestedCoins(blockTime))
}

// SpendableCoins returns the total number of spendable coins per denom for a
// periodic vesting account.
func (vva ValidatorVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return vva.BaseVestingAccount.SpendableCoinsFromVestingCoins(vva.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (vva *ValidatorVestingAccount) TrackDelegation(blockTime time.Time, amount sdk.Coins) {
	vva.BaseVestingAccount.TrackDelegation(vva.GetVestingCoins(blockTime), amount)
}

// Validate checks for errors on the account fields
func (vva ValidatorVestingAccount) Validate() error {
	if vva.SigningThreshold > 100 || vva.SigningThreshold < 0 {
		return errors.New("signing threshold must be between 0 and 100")
	}
	if vva.ReturnAddress.Equals(vva.Address) {
		return errors.New("return address cannot be the same as the account address")
	}
	return vva.PeriodicVestingAccount.Validate()
}

// MarshalYAML returns the YAML representation of an account.
func (vva ValidatorVestingAccount) MarshalYAML() (interface{}, error) {
	var bs []byte
	var err error
	var pubkey string

	if vva.PubKey != nil {
		pubkey, err = sdk.Bech32ifyAccPub(vva.PubKey)
		if err != nil {
			return nil, err
		}
	}

	bs, err = yaml.Marshal(struct {
		Address                sdk.AccAddress
		Coins                  sdk.Coins
		PubKey                 string
		AccountNumber          uint64
		Sequence               uint64
		OriginalVesting        sdk.Coins
		DelegatedFree          sdk.Coins
		DelegatedVesting       sdk.Coins
		EndTime                int64
		StartTime              int64
		VestingPeriods         vestingtypes.Periods
		ValidatorAddress       sdk.ConsAddress
		ReturnAddress          sdk.AccAddress
		SigningThreshold       int64
		MissingSignCount       []int64
		VestingPeriodProgress  [][]int
		DebtAfterFailedVesting sdk.Coins
	}{
		Address:                vva.Address,
		Coins:                  vva.Coins,
		PubKey:                 pubkey,
		AccountNumber:          vva.AccountNumber,
		Sequence:               vva.Sequence,
		OriginalVesting:        vva.OriginalVesting,
		DelegatedFree:          vva.DelegatedFree,
		DelegatedVesting:       vva.DelegatedVesting,
		EndTime:                vva.EndTime,
		StartTime:              vva.StartTime,
		VestingPeriods:         vva.VestingPeriods,
		ValidatorAddress:       vva.ValidatorAddress,
		ReturnAddress:          vva.ReturnAddress,
		SigningThreshold:       vva.SigningThreshold,
		MissingSignCount:       vva.MissingSignCount,
		VestingPeriodProgress:  vva.VestingPeriodProgress,
		DebtAfterFailedVesting: vva.DebtAfterFailedVesting,
	})
	if err != nil {
		return nil, err
	}

	return string(bs), err
}
