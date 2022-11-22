package types

import (
	fmt "fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// Parameter keys & defaults
var (
	KeyCommunityPoolInflation = []byte("CommunityPoolInflation")
	KeyStakingRewardsApy      = []byte("StakingRewardsApy")

	DefaultPreviousBlockTime      = tmtime.Canonical(time.Unix(1, 0))
	DefaultCommunityPoolInflation = sdk.ZeroDec()
	DefaultStakingRewardsApy      = sdk.ZeroDec()

	// rates larger than 17,650% are out of bounds
	// this is due to the necessary conversion of yearly rate to per second rate
	// TODO consider lowering max rate. when it's this large the precision is very bad.
	MaxMintingRate = sdk.NewDecWithPrec(1765, 1)
)

func NewParams(communityPoolInflation sdk.Dec, stakingRewardsApy sdk.Dec) Params {
	return Params{
		CommunityPoolInflation: communityPoolInflation,
		StakingRewardsApy:      stakingRewardsApy,
	}
}

// ParamKeyTable Key declaration for parameters
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

func validateCommunityPoolInflation(i interface{}) error {
	rate, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateRateWithinBounds(rate)
}

func validateStakingRewardsApy(i interface{}) error {
	rate, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateRateWithinBounds(rate)
}

// validateRateWithinBounds ensure that the given rate falls within the allowed bounds: [0, MaxMintingRate]
func validateRateWithinBounds(rate sdk.Dec) error {
	if rate.BigInt().Sign() == -1 {
		return fmt.Errorf("rate must be >= 0")
	}
	if MaxMintingRate.LT(rate) {
		return fmt.Errorf("rate out of bounds. the max allowed rate is %s", MaxMintingRate)
	}
	return nil
}
