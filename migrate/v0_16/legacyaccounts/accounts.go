package legacyaccounts

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"
)

// legacy implemtation of accounts that allows checking the spendable balance for any account,
// in addition support periodic vesting account methods
type Account interface {
	GetAddress() sdk.AccAddress
	GetCoins() sdk.Coins
	SpendableCoins(blockTime time.Time) sdk.Coins
}
type GenesisAccounts []GenesisAccount
type GenesisAccount interface {
	Account
}
type VestingAccount interface {
	Account
	GetVestedCoins(blockTime time.Time) sdk.Coins
	GetVestingCoins(blockTime time.Time) sdk.Coins
	GetStartTime() int64
	GetEndTime() int64
	GetOriginalVesting() sdk.Coins
	GetDelegatedFree() sdk.Coins
	GetDelegatedVesting() sdk.Coins
}

type GenesisState struct {
	Params   v034auth.Params `json:"params" yaml:"params"`
	Accounts GenesisAccounts `json:"accounts" yaml:"accounts"`
}

//-----------------------------------------------------------------------------
// BaseAccount
var _ Account = (*BaseAccount)(nil)
var _ GenesisAccount = (*BaseAccount)(nil)

type BaseAccount struct {
	Address       sdk.AccAddress     `json:"address" yaml:"address"`
	Coins         sdk.Coins          `json:"coins" yaml:"coins"`
	PubKey        cryptotypes.PubKey `json:"public_key" yaml:"public_key"`
	AccountNumber uint64             `json:"account_number" yaml:"account_number"`
	Sequence      uint64             `json:"sequence" yaml:"sequence"`
}

func NewBaseAccount(
	address sdk.AccAddress, coins sdk.Coins, pk cryptotypes.PubKey, accountNumber, sequence uint64,
) *BaseAccount {

	return &BaseAccount{
		Address:       address,
		Coins:         coins,
		PubKey:        pk,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
}

func (acc BaseAccount) GetAddress() sdk.AccAddress {
	return acc.Address
}
func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.Coins
}
func (acc *BaseAccount) SpendableCoins(_ time.Time) sdk.Coins {
	return acc.GetCoins()
}

type Period struct {
	Length int64     `json:"length" yaml:"length"` // length of the period, in seconds
	Amount sdk.Coins `json:"amount" yaml:"amount"` // amount of coins vesting during this period
}

type Periods []Period

var _ VestingAccount = (*PeriodicVestingAccount)(nil)
var _ GenesisAccount = (*PeriodicVestingAccount)(nil)

type vestingAccountJSON struct {
	Address          sdk.AccAddress     `json:"address" yaml:"address"`
	Coins            sdk.Coins          `json:"coins,omitempty" yaml:"coins"`
	PubKey           cryptotypes.PubKey `json:"public_key" yaml:"public_key"`
	AccountNumber    uint64             `json:"account_number" yaml:"account_number"`
	Sequence         uint64             `json:"sequence" yaml:"sequence"`
	OriginalVesting  sdk.Coins          `json:"original_vesting" yaml:"original_vesting"`
	DelegatedFree    sdk.Coins          `json:"delegated_free" yaml:"delegated_free"`
	DelegatedVesting sdk.Coins          `json:"delegated_vesting" yaml:"delegated_vesting"`
	EndTime          int64              `json:"end_time" yaml:"end_time"`

	// custom fields based on concrete vesting type which can be omitted
	StartTime      int64   `json:"start_time,omitempty" yaml:"start_time,omitempty"`
	VestingPeriods Periods `json:"vesting_periods,omitempty" yaml:"vesting_periods,omitempty"`
}

type BaseVestingAccount struct {
	*BaseAccount

	OriginalVesting  sdk.Coins `json:"original_vesting" yaml:"original_vesting"`   // coins in account upon initialization
	DelegatedFree    sdk.Coins `json:"delegated_free" yaml:"delegated_free"`       // coins that are vested and delegated
	DelegatedVesting sdk.Coins `json:"delegated_vesting" yaml:"delegated_vesting"` // coins that vesting and delegated
	EndTime          int64     `json:"end_time" yaml:"end_time"`                   // when the coins become unlocked
}

func (bva *BaseVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	bva.BaseAccount = NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence)
	bva.OriginalVesting = alias.OriginalVesting
	bva.DelegatedFree = alias.DelegatedFree
	bva.DelegatedVesting = alias.DelegatedVesting
	bva.EndTime = alias.EndTime

	return nil
}

func (bva BaseVestingAccount) SpendableCoinsVestingAccount(vestingCoins sdk.Coins) sdk.Coins {
	var spendableCoins sdk.Coins
	bc := bva.GetCoins()

	for _, coin := range bc {
		baseAmt := coin.Amount
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt := bva.DelegatedVesting.AmountOf(coin.Denom)

		// compute min((BC + DV) - V, BC) per the specification
		min := sdk.MinInt(baseAmt.Add(delVestingAmt).Sub(vestingAmt), baseAmt)
		spendableCoin := sdk.NewCoin(coin.Denom, min)

		if !spendableCoin.IsZero() {
			spendableCoins = spendableCoins.Add(spendableCoin)
		}
	}

	return spendableCoins
}
func (bva BaseVestingAccount) GetOriginalVesting() sdk.Coins {
	return bva.OriginalVesting
}
func (bva BaseVestingAccount) GetDelegatedFree() sdk.Coins {
	return bva.DelegatedFree
}
func (bva BaseVestingAccount) GetDelegatedVesting() sdk.Coins {
	return bva.DelegatedVesting
}
func (bva BaseVestingAccount) GetEndTime() int64 {
	return bva.EndTime
}

//-----------------------------------------------------------------------------
// Continuous Vesting Account
type ContinuousVestingAccount struct {
	*BaseVestingAccount

	StartTime int64 `json:"start_time" yaml:"start_time"` // when the coins start to vest
}

func (cva *ContinuousVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	cva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	cva.StartTime = alias.StartTime

	return nil
}

func (cva ContinuousVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if blockTime.Unix() <= cva.StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= cva.EndTime {
		return cva.OriginalVesting
	}

	// calculate the vesting scalar
	x := blockTime.Unix() - cva.StartTime
	y := cva.EndTime - cva.StartTime
	s := sdk.NewDec(x).Quo(sdk.NewDec(y))

	for _, ovc := range cva.OriginalVesting {
		vestedAmt := ovc.Amount.ToDec().Mul(s).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	return vestedCoins
}
func (cva ContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return cva.OriginalVesting.Sub(cva.GetVestedCoins(blockTime))
}
func (cva ContinuousVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return cva.BaseVestingAccount.SpendableCoinsVestingAccount(cva.GetVestingCoins(blockTime))
}
func (cva ContinuousVestingAccount) GetStartTime() int64 {
	return cva.StartTime
}

//-----------------------------------------------------------------------------
// Periodic Vesting Account
type PeriodicVestingAccount struct {
	*BaseVestingAccount
	StartTime      int64   `json:"start_time" yaml:"start_time"`           // when the coins start to vest
	VestingPeriods Periods `json:"vesting_periods" yaml:"vesting_periods"` // the vesting schedule
}

// NewPeriodicVestingAccountRaw creates a new PeriodicVestingAccount object from BaseVestingAccount
func NewPeriodicVestingAccountRaw(bva *BaseVestingAccount, startTime int64, periods Periods) *PeriodicVestingAccount {
	return &PeriodicVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}
}

func (pva PeriodicVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if blockTime.Unix() <= pva.StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= pva.EndTime {
		return pva.OriginalVesting
	}

	// track the start time of the next period
	currentPeriodStartTime := pva.StartTime
	// for each period, if the period is over, add those coins as vested and check the next period.
	for _, period := range pva.VestingPeriods {
		x := blockTime.Unix() - currentPeriodStartTime
		if x < period.Length {
			break
		}
		vestedCoins = vestedCoins.Add(period.Amount...)
		// Update the start time of the next period
		currentPeriodStartTime += period.Length
	}
	return vestedCoins
}
func (pva PeriodicVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	vestedCoins := pva.GetVestedCoins(blockTime)
	return pva.OriginalVesting.Sub(vestedCoins)
}
func (pva PeriodicVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	vestingCoins := pva.GetVestingCoins(blockTime)
	return pva.BaseVestingAccount.SpendableCoinsVestingAccount(vestingCoins)
}
func (pva PeriodicVestingAccount) GetStartTime() int64 {
	return pva.StartTime
}

//-----------------------------------------------------------------------------
// Delayed Vesting Account
type DelayedVestingAccount struct {
	*BaseVestingAccount
}

// UnmarshalJSON unmarshals raw JSON bytes into a DelayedVestingAccount.
func (dva *DelayedVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	dva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}

	return nil
}

// UnmarshalJSON unmarshals raw JSON bytes into a PeriodicVestingAccount.
func (pva *PeriodicVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	pva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	pva.StartTime = alias.StartTime
	pva.VestingPeriods = alias.VestingPeriods

	return nil
}

func (dva DelayedVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	if blockTime.Unix() >= dva.EndTime {
		return dva.OriginalVesting
	}

	return nil
}
func (dva DelayedVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return dva.OriginalVesting.Sub(dva.GetVestedCoins(blockTime))
}
func (dva DelayedVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return dva.BaseVestingAccount.SpendableCoinsVestingAccount(dva.GetVestingCoins(blockTime))
}
func (dva DelayedVestingAccount) GetStartTime() int64 {
	return 0
}

type ModuleAccount struct {
	*BaseAccount

	Name        string   `json:"name" yaml:"name"`               // name of the module
	Permissions []string `json:"permissions" yaml:"permissions"` // permissions of module account
}
type moduleAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	Coins         sdk.Coins      `json:"coins,omitempty" yaml:"coins"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	Name          string         `json:"name" yaml:"name"`
	Permissions   []string       `json:"permissions" yaml:"permissions"`
}

// UnmarshalJSON unmarshals raw JSON bytes into a ModuleAccount.
func (ma *ModuleAccount) UnmarshalJSON(bz []byte) error {
	var alias moduleAccountPretty
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	ma.BaseAccount = NewBaseAccount(alias.Address, alias.Coins, nil, alias.AccountNumber, alias.Sequence)
	ma.Name = alias.Name
	ma.Permissions = alias.Permissions

	return nil
}

type VestingProgress struct {
	PeriodComplete    bool `json:"period_complete" yaml:"period_complete"`
	VestingSuccessful bool `json:"vesting_successful" yaml:"vesting_successful"`
}
type CurrentPeriodProgress struct {
	MissedBlocks int64 `json:"missed_blocks" yaml:"missed_blocks"`
	TotalBlocks  int64 `json:"total_blocks" yaml:"total_blocks"`
}
type ValidatorVestingAccount struct {
	*PeriodicVestingAccount
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
	VestingPeriods         Periods               `json:"vesting_periods" yaml:"vesting_periods"`
	ValidatorAddress       sdk.ConsAddress       `json:"validator_address" yaml:"validator_address"`
	ReturnAddress          sdk.AccAddress        `json:"return_address" yaml:"return_address"`
	SigningThreshold       int64                 `json:"signing_threshold" yaml:"signing_threshold"`
	CurrentPeriodProgress  CurrentPeriodProgress `json:"current_period_progress" yaml:"current_period_progress"`
	VestingPeriodProgress  []VestingProgress     `json:"vesting_period_progress" yaml:"vesting_period_progress"`
	DebtAfterFailedVesting sdk.Coins             `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"`
}

// UnmarshalJSON unmarshals raw JSON bytes into a ValidatorVestingAccount.
func (vva *ValidatorVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias validatorVestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	ba := NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence)
	bva := &BaseVestingAccount{
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
			if vva.VestingPeriodProgress[i].PeriodComplete {
				vestedCoins = vestedCoins.Add(vva.VestingPeriods[i].Amount...)
			}
			currentPeriodStartTime += vva.VestingPeriods[i].Length
		} else {
			break
		}
	}
	return vestedCoins

}
func (vva ValidatorVestingAccount) GetFailedVestedCoins() sdk.Coins {
	var failedVestedCoins sdk.Coins
	numberPeriods := len(vva.VestingPeriods)
	for i := 0; i < numberPeriods; i++ {
		if vva.VestingPeriodProgress[i].PeriodComplete {
			if !vva.VestingPeriodProgress[i].VestingSuccessful {
				failedVestedCoins = failedVestedCoins.Add(vva.VestingPeriods[i].Amount...)
			}
		} else {
			break
		}
	}
	return failedVestedCoins
}
func (vva ValidatorVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return vva.OriginalVesting.Sub(vva.GetVestedCoins(blockTime))
}
func (vva ValidatorVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return vva.BaseVestingAccount.SpendableCoinsVestingAccount(vva.GetVestingCoins(blockTime))
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cryptocodec.RegisterCrypto(cdc)
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
	cdc.RegisterConcrete(&ValidatorVestingAccount{}, "cosmos-sdk/ValidatorVestingAccount", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
}
