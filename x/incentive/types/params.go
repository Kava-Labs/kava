package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	tmtime "github.com/tendermint/tendermint/types/time"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	kavadistTypes "github.com/kava-labs/kava/x/kavadist/types"
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

// DefaultParams returns default params for kavadist module
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
	Denom            string        `json:"denom" yaml:"denom"`                         // the collateral denom rewards apply to, must be found in the cdp collaterals
	AvailableRewards sdk.Coin      `json:"available_rewards" yaml:"available_rewards"` // the total amount of coins distributed per period
	Duration         time.Duration `json:"duration" yaml:"duration"`                   // the duration of the period
	TimeLock         time.Duration `json:"time_lock" yaml:"time_lock"`                 // how long rewards for this period are timelocked
	ClaimDuration    time.Duration `json:"claim_duration" yaml:"claim_duration"`       // how long users have after the period ends to claim their rewards
}

// NewReward returns a new Reward
func NewReward(active bool, denom string, reward sdk.Coin, duration time.Duration, timelock time.Duration, claimDuration time.Duration) Reward {
	return Reward{
		Active:           active,
		Denom:            denom,
		AvailableRewards: reward,
		Duration:         duration,
		TimeLock:         timelock,
		ClaimDuration:    claimDuration,
	}
}

// String implements fmt.Stringer
func (r Reward) String() string {
	return fmt.Sprintf(`Reward:
	Active: %t,
	Denom: %s,
	Available Rewards: %s,
	Duration: %s,
	Time Lock: %s,
	Claim Duration: %s`,
		r.Active, r.Denom, r.AvailableRewards, r.Duration, r.TimeLock, r.ClaimDuration)
}

// Validate performs a basic check of a reward fields.
func (r Reward) Validate() error {
	if !r.AvailableRewards.IsValid() {
		return fmt.Errorf("invalid reward coins %s for %s", r.AvailableRewards, r.Denom)
	}
	if !r.AvailableRewards.IsPositive() {
		return fmt.Errorf("reward amount must be positive, is %s for %s", r.AvailableRewards, r.Denom)
	}
	if r.Duration <= 0 {
		return fmt.Errorf("reward duration must be positive, is %s for %s", r.Duration, r.Denom)
	}
	if r.TimeLock < 0 {
		return fmt.Errorf("reward timelock must be non-negative, is %s for %s", r.TimeLock, r.Denom)
	}
	if r.ClaimDuration <= 0 {
		return fmt.Errorf("claim duration must be positive, is %s for %s", r.ClaimDuration, r.Denom)
	}
	return sdk.ValidateDenom(r.Denom)
}

// Rewards array of Reward
type Rewards []Reward

// Validate checks if all the rewards are valid and there are no duplicated
// entries.
func (rs Rewards) Validate() error {
	rewardDenoms := make(map[string]bool)
	for _, r := range rs {
		if rewardDenoms[r.Denom] {
			return fmt.Errorf("cannot have duplicate reward denoms: %s", r.Denom)
		}

		if err := r.Validate(); err != nil {
			return err
		}

		rewardDenoms[r.Denom] = true
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
