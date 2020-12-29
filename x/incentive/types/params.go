package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	tmtime "github.com/tendermint/tendermint/types/time"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	kavadistTypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Valid reward multipliers
const (
	Small  MultiplierName = "small"
	Medium MultiplierName = "medium"
	Large  MultiplierName = "large"
)

// Parameter keys and default values
var (
	KeyActive                       = []byte("Active")
	KeyRewards                      = []byte("RewardPeriods")
	DefaultActive                   = false
	DefaultRewardPeriods            = RewardPeriods{}
	DefaultClaims                   = Claims{}
	DefaultGenesisAccumulationTimes = GenesisAccumulationTimes{}
	DefaultPreviousBlockTime        = tmtime.Canonical(time.Unix(0, 0))
	GovDenom                        = cdptypes.DefaultGovDenom
	PrincipalDenom                  = "usdx"
	IncentiveMacc                   = kavadistTypes.ModuleName
)

// Params governance parameters for the incentive module
type Params struct {
	Active        bool          `json:"active" yaml:"active"` // top level governance switch to disable all rewards
	RewardPeriods RewardPeriods `json:"reward_periods" yaml:"reward_periods"`
}

// NewParams returns a new params object
func NewParams(active bool, rewards RewardPeriods) Params {
	return Params{
		Active:        active,
		RewardPeriods: rewards,
	}
}

// DefaultParams returns default params for incentive module
func DefaultParams() Params {
	return NewParams(DefaultActive, DefaultRewardPeriods)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Active: %t
	Rewards: %s`, p.Active, p.RewardPeriods)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyActive, &p.Active, validateActiveParam),
		params.NewParamSetPair(KeyRewards, &p.RewardPeriods, validateRewardsParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateActiveParam(p.Active); err != nil {
		return err
	}

	return validateRewardsParam(p.RewardPeriods)
}

func validateActiveParam(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateRewardsParam(i interface{}) error {
	rewards, ok := i.(RewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	Active           bool        `json:"active" yaml:"active"`
	CollateralType   string      `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time   `json:"start" yaml:"start"`
	End              time.Time   `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coin    `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
	ClaimEnd         time.Time   `json:"claim_end" yaml:"claim_end"`
	ClaimMultipliers Multipliers `json:"claim_multipliers" yaml:"claim_multipliers"` // the reward multiplier and timelock schedule - applied at the time users claim rewards
}

// String implements fmt.Stringer
func (rp RewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Collateral Type: %s,
	Start: %s,
	End: %s,
	Rewards Per Second: %s,
	Claim End: %s,
	Active %t,
	%s
	`, rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond, rp.ClaimEnd, rp.Active, rp.ClaimMultipliers)
}

// NewRewardPeriod returns a new RewardPeriod
func NewRewardPeriod(active bool, collateralType string, start time.Time, end time.Time, reward sdk.Coin, claimEnd time.Time, claimMultipliers Multipliers) RewardPeriod {
	return RewardPeriod{
		Active:           active,
		CollateralType:   collateralType,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
		ClaimEnd:         claimEnd,
		ClaimMultipliers: claimMultipliers,
	}
}

// GetMultiplier returns the named multiplier from the input reward period
func (rp RewardPeriod) GetMultiplier(name MultiplierName) (Multiplier, bool) {
	for _, multiplier := range rp.ClaimMultipliers {
		if multiplier.Name == name {
			return multiplier, true
		}
	}
	return Multiplier{}, false
}

// Validate performs a basic check of a RewardPeriod fields.
func (rp RewardPeriod) Validate() error {
	if rp.Start.IsZero() {
		return errors.New("reward period start time cannot be 0")
	}
	if rp.End.IsZero() {
		return errors.New("reward period end time cannot be 0")
	}
	if rp.Start.After(rp.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", rp.End, rp.Start)
	}
	if !rp.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", rp.RewardsPerSecond)
	}
	if rp.ClaimEnd.IsZero() {
		return errors.New("reward period claim end time cannot be 0")
	}
	if err := rp.ClaimMultipliers.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(rp.CollateralType) == "" {
		return fmt.Errorf("reward period collateral type cannot be blank: %s", rp)
	}
	return nil
}

// RewardPeriods array of RewardPeriod
type RewardPeriods []RewardPeriod

// Validate checks if all the RewardPeriods are valid and there are no duplicated
// entries.
func (rps RewardPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, rp := range rps {
		if seenPeriods[rp.CollateralType] {
			return fmt.Errorf("duplicated reward period with collateral type %s", rp.CollateralType)
		}

		if err := rp.Validate(); err != nil {
			return err
		}
		seenPeriods[rp.CollateralType] = true
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

// String implements fmt.Stringer
func (m Multiplier) String() string {
	return fmt.Sprintf(`Claim Multiplier:
	Name: %s
	Months Lockup %d
	Factor %s
	`, m.Name, m.MonthsLockup, m.Factor)
}

// Multipliers slice of Multiplier
type Multipliers []Multiplier

// Validate validates each multiplier
func (ms Multipliers) Validate() error {
	for _, m := range ms {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// String implements fmt.Stringer
func (ms Multipliers) String() string {
	out := "Claim Multipliers\n"
	for _, s := range ms {
		out += fmt.Sprintf("%s\n", s)
	}
	return out
}

// MultiplierName name for valid multiplier
type MultiplierName string

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return fmt.Errorf("invalid multiplier name: %s", mn)
}
