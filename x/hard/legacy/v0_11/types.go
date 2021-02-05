package v0_11

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/params"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"

	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "harvest"

	// LPAccount LP distribution module account
	LPAccount = "harvest_lp_distribution"

	// DelegatorAccount delegator distribution module account
	DelegatorAccount = "harvest_delegator_distribution"

	// ModuleAccountName name of module account used to hold deposits
	ModuleAccountName = "harvest"
)

// Parameter keys and default values
var (
	KeyActive                 = []byte("Active")
	KeyLPSchedules            = []byte("LPSchedules")
	KeyDelegatorSchedule      = []byte("DelegatorSchedule")
	DefaultActive             = true
	DefaultGovSchedules       = DistributionSchedules{}
	DefaultLPSchedules        = DistributionSchedules{}
	DefaultDelegatorSchedules = DelegatorDistributionSchedules{}
	DefaultPreviousBlockTime  = tmtime.Canonical(time.Unix(0, 0))
	DefaultDistributionTimes  = GenesisDistributionTimes{}
	GovDenom                  = cdptypes.DefaultGovDenom
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                    Params                   `json:"params" yaml:"params"`
	PreviousBlockTime         time.Time                `json:"previous_block_time" yaml:"previous_block_time"`
	PreviousDistributionTimes GenesisDistributionTimes `json:"previous_distribution_times" yaml:"previous_distribution_times"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, previousBlockTime time.Time, previousDistTimes GenesisDistributionTimes) GenesisState {
	return GenesisState{
		Params:                    params,
		PreviousBlockTime:         previousBlockTime,
		PreviousDistributionTimes: previousDistTimes,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.PreviousBlockTime.Equal(time.Time{}) {
		return fmt.Errorf("previous block time not set")
	}
	for _, gdt := range gs.PreviousDistributionTimes {
		if gdt.PreviousDistributionTime.Equal(time.Time{}) {
			return fmt.Errorf("previous distribution time not set for %s", gdt.Denom)
		}
		if err := sdk.ValidateDenom(gdt.Denom); err != nil {
			return err
		}
	}
	return nil
}

// GenesisDistributionTime stores the previous distribution time and its corresponding denom
type GenesisDistributionTime struct {
	Denom                    string    `json:"denom" yaml:"denom"`
	PreviousDistributionTime time.Time `json:"previous_distribution_time" yaml:"previous_distribution_time"`
}

// GenesisDistributionTimes slice of GenesisDistributionTime
type GenesisDistributionTimes []GenesisDistributionTime

// Params governance parameters for harvest module
type Params struct {
	Active                         bool                           `json:"active" yaml:"active"`
	LiquidityProviderSchedules     DistributionSchedules          `json:"liquidity_provider_schedules" yaml:"liquidity_provider_schedules"`
	DelegatorDistributionSchedules DelegatorDistributionSchedules `json:"delegator_distribution_schedules" yaml:"delegator_distribution_schedules"`
}

// DistributionSchedule distribution schedule for liquidity providers
type DistributionSchedule struct {
	Active           bool        `json:"active" yaml:"active"`
	DepositDenom     string      `json:"deposit_denom" yaml:"deposit_denom"`
	Start            time.Time   `json:"start" yaml:"start"`
	End              time.Time   `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coin    `json:"rewards_per_second" yaml:"rewards_per_second"`
	ClaimEnd         time.Time   `json:"claim_end" yaml:"claim_end"`
	ClaimMultipliers Multipliers `json:"claim_multipliers" yaml:"claim_multipliers"`
}

// NewDistributionSchedule returns a new DistributionSchedule
func NewDistributionSchedule(active bool, denom string, start, end time.Time, reward sdk.Coin, claimEnd time.Time, multipliers Multipliers) DistributionSchedule {
	return DistributionSchedule{
		Active:           active,
		DepositDenom:     denom,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
		ClaimEnd:         claimEnd,
		ClaimMultipliers: multipliers,
	}
}

// String implements fmt.Stringer
func (ds DistributionSchedule) String() string {
	return fmt.Sprintf(`Liquidity Provider Distribution Schedule:
	Deposit Denom: %s,
	Start: %s,
	End: %s,
	Rewards Per Second: %s,
	Claim End: %s,
	Active: %t
	`, ds.DepositDenom, ds.Start, ds.End, ds.RewardsPerSecond, ds.ClaimEnd, ds.Active)
}

// Validate performs a basic check of a distribution schedule.
func (ds DistributionSchedule) Validate() error {
	if !ds.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward coins %s for %s", ds.RewardsPerSecond, ds.DepositDenom)
	}
	if !ds.RewardsPerSecond.IsPositive() {
		return fmt.Errorf("reward amount must be positive, is %s for %s", ds.RewardsPerSecond, ds.DepositDenom)
	}
	if ds.RewardsPerSecond.Denom != "hard" {
		return fmt.Errorf("reward denom should be hard, is %s", ds.RewardsPerSecond.Denom)
	}
	if ds.Start.Unix() <= 0 {
		return errors.New("reward period start time cannot be 0")
	}
	if ds.End.Unix() <= 0 {
		return errors.New("reward period end time cannot be 0")
	}
	if ds.Start.After(ds.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", ds.End, ds.Start)
	}
	if ds.ClaimEnd.Before(ds.End) {
		return fmt.Errorf("claim end time %s cannot be before end time %s", ds.ClaimEnd, ds.End)
	}
	for _, multiplier := range ds.ClaimMultipliers {
		if err := multiplier.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DistributionSchedules slice of DistributionSchedule
type DistributionSchedules []DistributionSchedule

// Validate checks if all the LiquidityProviderSchedules are valid and there are no duplicated
// entries.
func (dss DistributionSchedules) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, ds := range dss {
		if seenPeriods[ds.DepositDenom] {
			return fmt.Errorf("duplicated distribution provider schedule with deposit denom %s", ds.DepositDenom)
		}

		if err := ds.Validate(); err != nil {
			return err
		}
		seenPeriods[ds.DepositDenom] = true
	}

	return nil
}

// String implements fmt.Stringer
func (dss DistributionSchedules) String() string {
	out := "Distribution Schedules\n"
	for _, ds := range dss {
		out += fmt.Sprintf("%s\n", ds)
	}
	return out
}

// DelegatorDistributionSchedule distribution schedule for delegators
type DelegatorDistributionSchedule struct {
	DistributionSchedule DistributionSchedule `json:"distribution_schedule" yaml:"distribution_schedule"`

	DistributionFrequency time.Duration `json:"distribution_frequency" yaml:"distribution_frequency"`
}

// NewDelegatorDistributionSchedule returns a new DelegatorDistributionSchedule
func NewDelegatorDistributionSchedule(ds DistributionSchedule, frequency time.Duration) DelegatorDistributionSchedule {
	return DelegatorDistributionSchedule{
		DistributionSchedule:  ds,
		DistributionFrequency: frequency,
	}
}

// Validate performs a basic check of a reward fields.
func (dds DelegatorDistributionSchedule) Validate() error {
	if err := dds.DistributionSchedule.Validate(); err != nil {
		return err
	}
	if dds.DistributionFrequency <= 0 {
		return fmt.Errorf("distribution frequency should be positive, got %d", dds.DistributionFrequency)
	}
	return nil
}

// DelegatorDistributionSchedules slice of DelegatorDistributionSchedule
type DelegatorDistributionSchedules []DelegatorDistributionSchedule

// Validate checks if all the LiquidityProviderSchedules are valid and there are no duplicated
// entries.
func (dds DelegatorDistributionSchedules) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, ds := range dds {
		if seenPeriods[ds.DistributionSchedule.DepositDenom] {
			return fmt.Errorf("duplicated liquidity provider schedule with deposit denom %s", ds.DistributionSchedule.DepositDenom)
		}

		if err := ds.Validate(); err != nil {
			return err
		}
		seenPeriods[ds.DistributionSchedule.DepositDenom] = true
	}

	return nil
}

// Multiplier amount the claim rewards get increased by, along with how long the claim rewards are locked
type Multiplier struct {
	Name         MultiplierName `json:"name" yaml:"name"`
	MonthsLockup int64          `json:"months_lockup" yaml:"months_lockup"`
	Factor       sdk.Dec        `json:"factor" yaml:"factor"`
}

// NewMultiplier returns a new Multiplier
func NewMultiplier(name MultiplierName, lockup int64, factor sdk.Dec) Multiplier {
	return Multiplier{
		Name:         name,
		MonthsLockup: lockup,
		Factor:       factor,
	}
}

// Validate multiplier param
func (m Multiplier) Validate() error {
	if err := m.Name.IsValid(); err != nil {
		return err
	}
	if m.MonthsLockup < 0 {
		return fmt.Errorf("expected non-negative lockup, got %d", m.MonthsLockup)
	}
	if m.Factor.IsNegative() {
		return fmt.Errorf("expected non-negative factor, got %s", m.Factor.String())
	}

	return nil
}

// GetMultiplier returns the named multiplier from the input distribution schedule
func (ds DistributionSchedule) GetMultiplier(name MultiplierName) (Multiplier, bool) {
	for _, multiplier := range ds.ClaimMultipliers {
		if multiplier.Name == name {
			return multiplier, true
		}
	}
	return Multiplier{}, false
}

// Multipliers slice of Multiplier
type Multipliers []Multiplier

// NewParams returns a new params object
func NewParams(active bool, lps DistributionSchedules, dds DelegatorDistributionSchedules) Params {
	return Params{
		Active:                         active,
		LiquidityProviderSchedules:     lps,
		DelegatorDistributionSchedules: dds,
	}
}

// DefaultParams returns default params for harvest module
func DefaultParams() Params {
	return NewParams(DefaultActive, DefaultLPSchedules, DefaultDelegatorSchedules)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Active: %t
	Liquidity Provider Distribution Schedules %s
	Delegator Distribution Schedule %s`, p.Active, p.LiquidityProviderSchedules, p.DelegatorDistributionSchedules)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyActive, &p.Active, validateActiveParam),
		params.NewParamSetPair(KeyLPSchedules, &p.LiquidityProviderSchedules, validateLPParams),
		params.NewParamSetPair(KeyDelegatorSchedule, &p.DelegatorDistributionSchedules, validateDelegatorParams),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateActiveParam(p.Active); err != nil {
		return err
	}

	if err := validateDelegatorParams(p.DelegatorDistributionSchedules); err != nil {
		return err
	}

	return validateLPParams(p.LiquidityProviderSchedules)
}

func validateActiveParam(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateLPParams(i interface{}) error {
	dss, ok := i.(DistributionSchedules)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, ds := range dss {
		err := ds.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func validateDelegatorParams(i interface{}) error {
	dds, ok := i.(DelegatorDistributionSchedules)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return dds.Validate()
}

// MultiplierName name for valid multiplier
type MultiplierName string

// DepositType type for valid deposit type strings
type DepositType string

// Valid reward multipliers and reward types
const (
	Small  MultiplierName = "small"
	Medium MultiplierName = "medium"
	Large  MultiplierName = "large"

	LP    DepositType = "lp"
	Stake DepositType = "stake"
)

// Queryable deposit types
var (
	DepositTypesDepositQuery = []DepositType{LP}
	DepositTypesClaimQuery   = []DepositType{LP, Stake}
)

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return fmt.Errorf("invalid multiplier name: %s", mn)
}

// IsValid checks if the input is one of the expected strings
func (dt DepositType) IsValid() error {
	switch dt {
	case LP, Stake:
		return nil
	}
	return fmt.Errorf("invalid deposit type: %s", dt)
}

// Deposit defines an amount of coins deposited into a harvest module account
type Deposit struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coin       `json:"amount" yaml:"amount"`
	Type      DepositType    `json:"type" yaml:"type"`
}

// NewDeposit returns a new deposit
func NewDeposit(depositor sdk.AccAddress, amount sdk.Coin, dtype DepositType) Deposit {
	return Deposit{
		Depositor: depositor,
		Amount:    amount,
		Type:      dtype,
	}
}

// Claim defines an amount of coins that the owner can claim
type Claim struct {
	Owner        sdk.AccAddress `json:"owner" yaml:"owner"`
	DepositDenom string         `json:"deposit_denom" yaml:"deposit_denom"`
	Amount       sdk.Coin       `json:"amount" yaml:"amount"`
	Type         DepositType    `json:"type" yaml:"type"`
}

// NewClaim returns a new claim
func NewClaim(owner sdk.AccAddress, denom string, amount sdk.Coin, dtype DepositType) Claim {
	return Claim{
		Owner:        owner,
		DepositDenom: denom,
		Amount:       amount,
		Type:         dtype,
	}
}

// NewPeriod returns a new vesting period
func NewPeriod(amount sdk.Coins, length int64) vesting.Period {
	return vesting.Period{Amount: amount, Length: length}
}

// GetTotalVestingPeriodLength returns the summed length of all vesting periods
func GetTotalVestingPeriodLength(periods vesting.Periods) int64 {
	length := int64(0)
	for _, period := range periods {
		length += period.Length
	}
	return length
}
