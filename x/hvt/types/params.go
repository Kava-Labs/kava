package types

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	tmtime "github.com/tendermint/tendermint/types/time"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// Parameter keys and default values
var (
	KeyActive                 = []byte("Active")
	KeyLPSchedules            = []byte("LPSchedules")
	KeyGovSchedules           = []byte("GovSchedules")
	KeyDelegatorSchedule      = []byte("DelegatorSchedule")
	DefaultActive             = true
	DefaultGovSchedules       = DistributionSchedules{}
	DefaultLPSchedules        = DistributionSchedules{}
	DefaultDelegatorSchedules = DelegatorDistributionSchedules{}
	DefaultPreviousBlockTime  = tmtime.Canonical(time.Unix(0, 0))
	GovDenom                  = cdptypes.DefaultGovDenom
)

// Params governance parameters for kavadist module
type Params struct {
	Active                          bool                           `json:"active" yaml:"active"`
	LiquidityProviderSchedules      DistributionSchedules          `json:"liquidity_provider_schedules" yaml:"liquidity_provider_schedules"`
	GovernanceDistributionSchedules DistributionSchedules          `json:"governance_distribution_schedules" yaml:"governance_distribution_schedules"`
	DelegatorDistributionSchedules  DelegatorDistributionSchedules `json:"delegator_distribution_schedules" yaml:"delegator_distribution_schedules"`
}

// DistributionSchedule distribution schedule for liquidity providers
type DistributionSchedule struct {
	Active           bool        `json:"active" yaml:"active"`
	DepositDenom     string      `json:"deposit_denom" yaml:"deposit_denom"`
	Start            time.Time   `json:"start" yaml:"start"`
	End              time.Time   `json:"end" yaml:"end"`
	Reward           sdk.Coin    `json:"reward" yaml:"reward"`
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
		Reward:           reward,
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
	Reward: %s,
	Claim End: %s,
	Active: %t
	`, ds.DepositDenom, ds.Start, ds.End, ds.Reward, ds.ClaimEnd, ds.Active)
}

// Validate performs a basic check of a distribution schedule.
func (ds DistributionSchedule) Validate() error {
	if !ds.Reward.IsValid() {
		return fmt.Errorf("invalid reward coins %s for %s", ds.Reward, ds.DepositDenom)
	}
	if !ds.Reward.IsPositive() {
		return fmt.Errorf("reward amount must be positive, is %s for %s", ds.Reward, ds.DepositDenom)
	}
	if ds.Start.IsZero() {
		return errors.New("reward period start time cannot be 0")
	}
	if ds.End.IsZero() {
		return errors.New("reward period end time cannot be 0")
	}
	if ds.Start.After(ds.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", ds.End, ds.Start)
	}
	if ds.ClaimEnd.Before(ds.End) {
		return fmt.Errorf("claim end time %s cannot be before end time %s", ds.ClaimEnd, ds.End)
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
	Name         RewardMultiplier `json:"name" yaml:"name"`
	LockDuration time.Duration    `json:"lock_duration" yaml:"lock_duration"`
	Factor       sdk.Dec          `json:"factor" yaml:"factor"`
}

// NewMultiplier returns a new Multiplier
func NewMultiplier(name RewardMultiplier, duration time.Duration, factor sdk.Dec) Multiplier {
	return Multiplier{
		Name:         name,
		LockDuration: duration,
		Factor:       factor,
	}
}

// Multipliers slice of Multiplier
type Multipliers []Multiplier

// NewParams returns a new params object
func NewParams(active bool, lps, gds DistributionSchedules, dds DelegatorDistributionSchedules) Params {
	return Params{
		Active:                          active,
		LiquidityProviderSchedules:      lps,
		GovernanceDistributionSchedules: gds,
		DelegatorDistributionSchedules:  dds,
	}
}

// DefaultParams returns default params for kavadist module
func DefaultParams() Params {
	return NewParams(DefaultActive, DefaultLPSchedules, DefaultGovSchedules, DefaultDelegatorSchedules)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Active: %t
	Liquidity Provider Distribution Schedules %s
	Governance Distribution Schedules %s
	Delegator Distribution Schedule %s`, p.Active, p.LiquidityProviderSchedules, p.GovernanceDistributionSchedules, p.DelegatorDistributionSchedules)
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
		params.NewParamSetPair(KeyGovSchedules, &p.GovernanceDistributionSchedules, validateGovParams),
		params.NewParamSetPair(KeyDelegatorSchedule, &p.DelegatorDistributionSchedules, validateDelegatorParams),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateActiveParam(p.Active); err != nil {
		return err
	}

	if err := validateGovParams(p.GovernanceDistributionSchedules); err != nil {
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

func validateGovParams(i interface{}) error {
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
