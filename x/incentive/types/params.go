package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	kavadistTypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Parameter keys and default values
var (
	KeyUSDXMintingRewardPeriods = []byte("USDXMintingRewardPeriods")
	KeyHardSupplyRewardPeriods  = []byte("HardSupplyRewardPeriods")
	KeyHardBorrowRewardPeriods  = []byte("HardBorrowRewardPeriods")
	KeyDelegatorRewardPeriods   = []byte("DelegatorRewardPeriods")
	KeySwapRewardPeriods        = []byte("SwapRewardPeriods")
	KeyClaimEnd                 = []byte("ClaimEnd")
	KeyMultipliers              = []byte("ClaimMultipliers")

	DefaultActive             = false
	DefaultRewardPeriods      = RewardPeriods{}
	DefaultMultiRewardPeriods = MultiRewardPeriods{}
	DefaultMultipliers        = MultipliersPerDenoms{}
	DefaultClaimEnd           = tmtime.Canonical(time.Unix(1, 0))

	BondDenom              = "ukava"
	USDXMintingRewardDenom = "ukava"

	IncentiveMacc = kavadistTypes.ModuleName
)

// NewParams returns a new params object
func NewParams(usdxMinting RewardPeriods, hardSupply, hardBorrow, delegator, swap MultiRewardPeriods,
	multipliers MultipliersPerDenoms, claimEnd time.Time) Params {
	return Params{
		USDXMintingRewardPeriods: usdxMinting,
		HardSupplyRewardPeriods:  hardSupply,
		HardBorrowRewardPeriods:  hardBorrow,
		DelegatorRewardPeriods:   delegator,
		SwapRewardPeriods:        swap,
		ClaimMultipliers:         multipliers,
		ClaimEnd:                 claimEnd,
	}
}

// DefaultParams returns default params for incentive module
func DefaultParams() Params {
	return NewParams(
		DefaultRewardPeriods,
		DefaultMultiRewardPeriods,
		DefaultMultiRewardPeriods,
		DefaultMultiRewardPeriods,
		DefaultMultiRewardPeriods,
		DefaultMultipliers,
		DefaultClaimEnd,
	)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyUSDXMintingRewardPeriods, &p.USDXMintingRewardPeriods, validateRewardPeriodsParam),
		paramtypes.NewParamSetPair(KeyHardSupplyRewardPeriods, &p.HardSupplyRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(KeyHardBorrowRewardPeriods, &p.HardBorrowRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(KeyDelegatorRewardPeriods, &p.DelegatorRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(KeySwapRewardPeriods, &p.SwapRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(KeyMultipliers, &p.ClaimMultipliers, validateMultipliersPerDenomParam),
		paramtypes.NewParamSetPair(KeyClaimEnd, &p.ClaimEnd, validateClaimEndParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {

	if err := validateMultipliersPerDenomParam(p.ClaimMultipliers); err != nil {
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

	if err := validateMultiRewardPeriodsParam(p.DelegatorRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.SwapRewardPeriods); err != nil {
		return err
	}

	return nil
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

func validateMultipliersPerDenomParam(i interface{}) error {
	multipliers, ok := i.(MultipliersPerDenoms)
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

// NewMultiRewardPeriodFromRewardPeriod converts a RewardPeriod into a MultiRewardPeriod.
// It's useful for compatibility between single and multi denom rewards.
func NewMultiRewardPeriodFromRewardPeriod(period RewardPeriod) MultiRewardPeriod {
	return NewMultiRewardPeriod(
		period.Active,
		period.CollateralType,
		period.Start,
		period.End,
		sdk.NewCoins(period.RewardsPerSecond),
	)
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
		// This is needed to ensure that the begin blocker accumulation does not panic.
		return fmt.Errorf("end period time %s cannot be before start time %s", rp.End, rp.Start)
	}
	if rp.RewardsPerSecond.Denom != USDXMintingRewardDenom {
		return fmt.Errorf("reward denom must be %s, got: %s", USDXMintingRewardDenom, rp.RewardsPerSecond.Denom)
	}
	if !rp.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", rp.RewardsPerSecond)
	}
	if strings.TrimSpace(rp.CollateralType) == "" {
		return fmt.Errorf("reward period collateral type cannot be blank: %v", rp)
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

// NewMultiRewardPeriod returns a new MultiRewardPeriod
func NewMultiRewardPeriod(active bool, collateralType string, start time.Time, end time.Time, reward sdk.Coins) MultiRewardPeriod {
	return MultiRewardPeriod{
		Active:           active,
		CollateralType:   collateralType,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
	}
}

// Validate performs a basic check of a MultiRewardPeriod.
func (mrp MultiRewardPeriod) Validate() error {
	if mrp.Start.IsZero() {
		return errors.New("reward period start time cannot be 0")
	}
	if mrp.End.IsZero() {
		return errors.New("reward period end time cannot be 0")
	}
	if mrp.Start.After(mrp.End) {
		// This is needed to ensure that the begin blocker accumulation does not panic.
		return fmt.Errorf("end period time %s cannot be before start time %s", mrp.End, mrp.Start)
	}
	if !mrp.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", mrp.RewardsPerSecond)
	}
	if strings.TrimSpace(mrp.CollateralType) == "" {
		return fmt.Errorf("reward period collateral type cannot be blank: %v", mrp)
	}
	return nil
}

// MultiRewardPeriods array of MultiRewardPeriod
type MultiRewardPeriods []MultiRewardPeriod

// GetMultiRewardPeriod fetches a MultiRewardPeriod from an array of MultiRewardPeriods by its denom
func (mrps MultiRewardPeriods) GetMultiRewardPeriod(denom string) (MultiRewardPeriod, bool) {
	for _, rp := range mrps {
		if rp.CollateralType == denom {
			return rp, true
		}
	}
	return MultiRewardPeriod{}, false
}

// GetMultiRewardPeriodIndex returns the index of a MultiRewardPeriod inside array MultiRewardPeriods
func (mrps MultiRewardPeriods) GetMultiRewardPeriodIndex(denom string) (int, bool) {
	for i, rp := range mrps {
		if rp.CollateralType == denom {
			return i, true
		}
	}
	return -1, false
}

// Validate checks if all the RewardPeriods are valid and there are no duplicated
// entries.
func (mrps MultiRewardPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, rp := range mrps {
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
