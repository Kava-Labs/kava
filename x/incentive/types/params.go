package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	kavadistTypes "github.com/kava-labs/kava/x/kavadist/types"
	tmtime "github.com/tendermint/tendermint/types/time"
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
		{Key: KeyActive, Value: &p.Active},
		{Key: KeyRewards, Value: &p.Rewards},
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	rewardDenoms := make(map[string]bool)

	for _, reward := range p.Rewards {
		if rewardDenoms[reward.Denom] {
			return fmt.Errorf("cannot have duplicate reward denoms: %s", reward.Denom)
		}
		rewardDenoms[reward.Denom] = true
		if !reward.Reward.IsPositive() {
			return fmt.Errorf("reward must be positive, is %s for %s", reward.Reward, reward.Denom)
		}
		if int(reward.Duration.Seconds()) <= 0 {
			return fmt.Errorf("reward duration must be positive, is %s for %s", reward.Duration.String(), reward.Denom)
		}
		if int(reward.TimeLock.Seconds()) < 0 {
			return fmt.Errorf("reward timelock must be non-negative, is %s for %s", reward.TimeLock.String(), reward.Denom)
		}
		if int(reward.ClaimDuration.Seconds()) <= 0 {
			return fmt.Errorf("reward timelock must be positive, is %s for %s", reward.ClaimDuration.String(), reward.Denom)
		}
	}
	return nil
}

// Reward stores the specified state for a single reward period.
type Reward struct {
	Active        bool          `json:"active" yaml:"actlive"`                // governance switch to disable a period
	Denom         string        `json:"denom" yaml:"denom"`                   // the collateral denom rewards apply to, must be found in the cdp collaterals
	Reward        sdk.Coin      `json:"reward" yaml:"reward"`                 // the total amount of coins distributed per period
	Duration      time.Duration `json:"duration" yaml:"duration"`             // the duration of the period
	TimeLock      time.Duration `json:"time_lock" yaml:"time_lock"`           // how long rewards for this period are timelocked
	ClaimDuration time.Duration `json:"claim_duration" yaml:"claim_duration"` // how long users have after the period ends to claim their rewards
}

// NewReward returns a new Reward
func NewReward(active bool, denom string, reward sdk.Coin, duration time.Duration, timelock time.Duration, claimDuration time.Duration) Reward {
	return Reward{
		Active:        active,
		Denom:         denom,
		Reward:        reward,
		Duration:      duration,
		TimeLock:      timelock,
		ClaimDuration: claimDuration,
	}
}

// String implements fmt.Stringer
func (r Reward) String() string {
	return fmt.Sprintf(`Reward Period:
	Active: %t,
	Denom: %s,
	Reward: %s,
	Duration: %s,
	Time Lock: %s,
	Claim Duration: %s`,
		r.Active, r.Denom, r.Reward, r.Duration, r.TimeLock, r.ClaimDuration)
}

// Rewards array of Reward
type Rewards []Reward

// String implements fmt.Stringer
func (rs Rewards) String() string {
	out := "Reward Periods\n"
	for _, r := range rs {
		out += fmt.Sprintf("%s\n", r)
	}
	return out
}
