package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyCommunityPoolInflation = []byte("CommunityPoolInflation")
	KeyStakingRewardsApy      = []byte("StakingRewardsApy")
)

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

// Validate checks that the parameters have valid values.
func (p *Params) Validate() error {
	if err := validateCommunityPoolInflation(p.CommunityPoolInflation); err != nil {
		return err
	}
	return validateStakingRewardsApy(p.StakingRewardsApy)
}

func validateCommunityPoolInflation(i interface{}) error {
	_, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateStakingRewardsApy(i interface{}) error {
	_, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
