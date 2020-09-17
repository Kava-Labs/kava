package types

import (
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
	KeyActive                = []byte("Active")
	KeyRewards               = []byte("Rewards")
	DefaultActive            = false
	DefaultRewards           = Rewards{}
	DefaultPreviousBlockTime = tmtime.Canonical(time.Unix(0, 0))
	GovDenom                 = cdptypes.DefaultGovDenom
	PrincipalDenom           = "usdx"
	IncentiveMacc            = kavadistTypes.ModuleName
)

// Params governance parameters for the incentive module
type Params struct {
	Active  bool    `json:"active" yaml:"active"` // top level governance switch to disable all rewards
	Rewards Rewards `json:"rewards" yaml:"rewards"`
}

// NewParams returns a new params object
func NewParams(active bool, rewards Rewards) Params {
	return Params{
		Active:  active,
		Rewards: rewards,
	}
}

// DefaultParams returns default params for incentive module
func DefaultParams() Params {
	return NewParams(DefaultActive, DefaultRewards)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Active: %t
	Rewards: %s`, p.Active, p.Rewards)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyActive, &p.Active, validateActiveParam),
		params.NewParamSetPair(KeyRewards, &p.Rewards, validateRewardsParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateActiveParam(p.Active); err != nil {
		return err
	}

	return validateRewardsParam(p.Rewards)
}

func validateActiveParam(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateRewardsParam(i interface{}) error {
	rewards, ok := i.(Rewards)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

// Reward stores the specified state for a single reward period.
type Reward struct {
	Active           bool          `json:"active" yaml:"active"`                       // governance switch to disable a period
	CollateralType   string        `json:"collateral_type" yaml:"collateral_type"`     // the collateral type rewards apply to, must be found in the cdp collaterals
	AvailableRewards sdk.Coin      `json:"available_rewards" yaml:"available_rewards"` // the total amount of coins distributed per period
	Duration         time.Duration `json:"duration" yaml:"duration"`                   // the duration of the period
	ClaimMultipliers Multipliers   `json:"claim_multipliers" yaml:"claim_multipliers"` // the reward multiplier and timelock schedule - applied at the time users claim rewards
	ClaimDuration    time.Duration `json:"claim_duration" yaml:"claim_duration"`       // how long users have after the period ends to claim their rewards
}

// NewReward returns a new Reward
func NewReward(active bool, collateralType string, reward sdk.Coin, duration time.Duration, multiplier Multipliers, claimDuration time.Duration) Reward {
	return Reward{
		Active:           active,
		CollateralType:   collateralType,
		AvailableRewards: reward,
		Duration:         duration,
		ClaimMultipliers: multiplier,
		ClaimDuration:    claimDuration,
	}
}

// String implements fmt.Stringer
func (r Reward) String() string {
	return fmt.Sprintf(`Reward:
	Active: %t,
	CollateralType: %s,
	Available Rewards: %s,
	Duration: %s,
	%s,
	Claim Duration: %s`,
		r.Active, r.CollateralType, r.AvailableRewards, r.Duration, r.ClaimMultipliers, r.ClaimDuration)
}

// Validate performs a basic check of a reward fields.
func (r Reward) Validate() error {
	if !r.AvailableRewards.IsValid() {
		return fmt.Errorf("invalid reward coins %s for %s", r.AvailableRewards, r.CollateralType)
	}
	if !r.AvailableRewards.IsPositive() {
		return fmt.Errorf("reward amount must be positive, is %s for %s", r.AvailableRewards, r.CollateralType)
	}
	if r.Duration <= 0 {
		return fmt.Errorf("reward duration must be positive, is %s for %s", r.Duration, r.CollateralType)
	}
	if err := r.ClaimMultipliers.Validate(); err != nil {
		return err
	}
	if r.ClaimDuration <= 0 {
		return fmt.Errorf("claim duration must be positive, is %s for %s", r.ClaimDuration, r.CollateralType)
	}
	if strings.TrimSpace(r.CollateralType) == "" {
		return fmt.Errorf("collateral type cannot be blank: %s", r)
	}
	return nil
}

// Rewards array of Reward
type Rewards []Reward

// Validate checks if all the rewards are valid and there are no duplicated
// entries.
func (rs Rewards) Validate() error {
	rewardCollateralTypes := make(map[string]bool)
	for _, r := range rs {
		if rewardCollateralTypes[r.CollateralType] {
			return fmt.Errorf("cannot have duplicate reward collateral types: %s", r.CollateralType)
		}

		if err := r.Validate(); err != nil {
			return err
		}

		rewardCollateralTypes[r.CollateralType] = true
	}
	return nil
}

// String implements fmt.Stringer
func (rs Rewards) String() string {
	out := "Rewards\n"
	for _, r := range rs {
		out += fmt.Sprintf("%s\n", r)
	}
	return out
}

// Multiplier amount the claim rewards get increased by, along with how long the claim rewards are locked
type Multiplier struct {
	Name         MultiplierName `json:"name" yaml:"name"`
	MonthsLockup int            `json:"months_lockup" yaml:"months_lockup"`
	Factor       sdk.Dec        `json:"factor" yaml:"factor"`
}

// NewMultiplier returns a new Multiplier
func NewMultiplier(name MultiplierName, lockup int, factor sdk.Dec) Multiplier {
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
