package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var (
	KeyCommunityPoolInflation = []byte("CommunityPoolInflation")
	KeyStakingRewardsApy      = []byte("StakingRewardsApy")

	// default inflation values are zero
	DefaultCommunityPoolInflation = sdk.ZeroDec()
	DefaultStakingRewardsApy      = sdk.ZeroDec()

	// MaxMintingRate returns the per second rate equivalent to 10,000% per year
	MaxMintingRate = sdk.NewDec(100)
)

// NewParams returns new Params with inflation rates set
func NewParams(communityPoolInflation sdk.Dec, stakingRewardsApy sdk.Dec) Params {
	return Params{
		CommunityPoolInflation: communityPoolInflation,
		StakingRewardsApy:      stakingRewardsApy,
	}
}

// ParamKeyTable returns the key table for the kavamint module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyCommunityPoolInflation, &p.CommunityPoolInflation, validateCommunityPoolInflation),
		paramtypes.NewParamSetPair(KeyStakingRewardsApy, &p.StakingRewardsApy, validateStakingRewardsApy),
	}
}

// DefaultParams returns default valid parameters for the kavamint module
func DefaultParams() Params {
	return NewParams(DefaultCommunityPoolInflation, DefaultStakingRewardsApy)
}

// Validate checks that the parameters have valid values.
func (p *Params) Validate() error {
	if err := validateCommunityPoolInflation(p.CommunityPoolInflation); err != nil {
		return err
	}
	return validateStakingRewardsApy(p.StakingRewardsApy)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	CommunityPoolInflation: %s
	StakingRewardsApy: %s`,
		p.CommunityPoolInflation, p.StakingRewardsApy)
}

func validateCommunityPoolInflation(i interface{}) error {
	rate, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateRate(rate)
}

func validateStakingRewardsApy(i interface{}) error {
	rate, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateRate(rate)
}

// validateRate ensures rate is properly initialized (non-nil), not negative, and not greater than the max rate
func validateRate(rate sdk.Dec) error {
	if rate.IsNil() || rate.IsNegative() || rate.GT(MaxMintingRate) {
		return fmt.Errorf(fmt.Sprintf("invalid rate: %s", rate))
	}
	return nil
}
