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
	KeyUSDXMintingRewardPeriods     = []byte("USDXMintingRewardPeriods")
	KeyHardSupplyRewardPeriods      = []byte("HardSupplyRewardPeriods")
	KeyHardBorrowRewardPeriods      = []byte("HardBorrowRewardPeriods")
	KeyHardDelegatorRewardPeriods   = []byte("HardDelegatorRewardPeriods")
	KeyClaimEnd                     = []byte("ClaimEnd")
	KeyMultipliers                  = []byte("ClaimMultipliers")
	DefaultActive                   = false
	DefaultRewardPeriods            = RewardPeriods{}
	DefaultMultiRewardPeriods       = MultiRewardPeriods{}
	DefaultMultipliers              = Multipliers{}
	DefaultUSDXClaims               = USDXMintingClaims{}
	DefaultHardClaims               = HardLiquidityProviderClaims{}
	DefaultGenesisAccumulationTimes = GenesisAccumulationTimes{}
	DefaultClaimEnd                 = tmtime.Canonical(time.Unix(1, 0))
	GovDenom                        = cdptypes.DefaultGovDenom
	PrincipalDenom                  = "usdx"
	IncentiveMacc                   = kavadistTypes.ModuleName
)

// Params governance parameters for the incentive module
type Params struct {
	USDXMintingRewardPeriods   RewardPeriods      `json:"usdx_minting_reward_periods" yaml:"usdx_minting_reward_periods"`
	HardSupplyRewardPeriods    MultiRewardPeriods `json:"hard_supply_reward_periods" yaml:"hard_supply_reward_periods"`
	HardBorrowRewardPeriods    MultiRewardPeriods `json:"hard_borrow_reward_periods" yaml:"hard_borrow_reward_periods"`
	HardDelegatorRewardPeriods RewardPeriods      `json:"hard_delegator_reward_periods" yaml:"hard_delegator_reward_periods"`
	ClaimMultipliers           Multipliers        `json:"claim_multipliers" yaml:"claim_multipliers"`
	ClaimEnd                   time.Time          `json:"claim_end" yaml:"claim_end"`
}

// NewParams returns a new params object
func NewParams(usdxMinting RewardPeriods, hardSupply, hardBorrow MultiRewardPeriods,
	hardDelegator RewardPeriods, multipliers Multipliers, claimEnd time.Time) Params {
	return Params{
		USDXMintingRewardPeriods:   usdxMinting,
		HardSupplyRewardPeriods:    hardSupply,
		HardBorrowRewardPeriods:    hardBorrow,
		HardDelegatorRewardPeriods: hardDelegator,
		ClaimMultipliers:           multipliers,
		ClaimEnd:                   claimEnd,
	}
}

// DefaultParams returns default params for incentive module
func DefaultParams() Params {
	return NewParams(DefaultRewardPeriods, DefaultMultiRewardPeriods,
		DefaultMultiRewardPeriods, DefaultRewardPeriods, DefaultMultipliers, DefaultClaimEnd)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	USDX Minting Reward Periods: %s
	Hard Supply Reward Periods: %s
	Hard Borrow Reward Periods: %s
	Hard Delegator Reward Periods: %s
	Claim Multipliers :%s
	Claim End Time: %s
	`, p.USDXMintingRewardPeriods, p.HardSupplyRewardPeriods, p.HardBorrowRewardPeriods,
		p.HardDelegatorRewardPeriods, p.ClaimMultipliers, p.ClaimEnd)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyUSDXMintingRewardPeriods, &p.USDXMintingRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyHardSupplyRewardPeriods, &p.HardSupplyRewardPeriods, validateMultiRewardPeriodsParam),
		params.NewParamSetPair(KeyHardBorrowRewardPeriods, &p.HardBorrowRewardPeriods, validateMultiRewardPeriodsParam),
		params.NewParamSetPair(KeyHardDelegatorRewardPeriods, &p.HardDelegatorRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyClaimEnd, &p.ClaimEnd, validateClaimEndParam),
		params.NewParamSetPair(KeyMultipliers, &p.ClaimMultipliers, validateMultipliersParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {

	if err := validateMultipliersParam(p.ClaimMultipliers); err != nil {
		return err
	}

	if err := validateRewardPeriodsParam(p.USDXMintingRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.HardSupplyRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.HardBorrowRewardPeriods); err != nil {
		return err
	}

	return validateRewardPeriodsParam(p.HardDelegatorRewardPeriods)
}

func validateRewardPeriodsParam(i interface{}) error {
	rewards, ok := i.(RewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

func validateMultiRewardPeriodsParam(i interface{}) error {
	rewards, ok := i.(MultiRewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

func validateMultipliersParam(i interface{}) error {
	multipliers, ok := i.(Multipliers)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return multipliers.Validate()
}

func validateClaimEndParam(i interface{}) error {
	endTime, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if endTime.Unix() <= 0 {
		return fmt.Errorf("end time should not be zero")
	}
	return nil
}

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	Active           bool      `json:"active" yaml:"active"`
	CollateralType   string    `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time `json:"start" yaml:"start"`
	End              time.Time `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coin  `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}

// String implements fmt.Stringer
func (rp RewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Collateral Type: %s,
	Start: %s,
	End: %s,
	Rewards Per Second: %s,
	Active %t,
	`, rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond, rp.Active)
}

// NewRewardPeriod returns a new RewardPeriod
func NewRewardPeriod(active bool, collateralType string, start time.Time, end time.Time, reward sdk.Coin) RewardPeriod {
	return RewardPeriod{
		Active:           active,
		CollateralType:   collateralType,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
	}
}

// Validate performs a basic check of a RewardPeriod fields.
func (rp RewardPeriod) Validate() error {
	if rp.Start.Unix() <= 0 {
		return errors.New("reward period start time cannot be 0")
	}
	if rp.End.Unix() <= 0 {
		return errors.New("reward period end time cannot be 0")
	}
	if rp.Start.After(rp.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", rp.End, rp.Start)
	}
	if !rp.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", rp.RewardsPerSecond)
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
