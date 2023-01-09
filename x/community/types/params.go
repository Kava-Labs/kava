package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

// Parameter keys and default values
var (
	KeyEnabledProposalMsgUrl   = []byte("EnabledProposalMsgUrls")
	DefaultEnabledProposalUrls = []string{sdk.MsgTypeURL(&evmutiltypes.MsgEVMCall{})}
)

// ParamKeyTable for community module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs pairs of the community module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEnabledProposalMsgUrl, &p.EnabledProposalMsgUrls, validateEnabledProposalMsgUrlsParam),
	}
}

// NewParams returns new community module Params.
func NewParams(enabledUrls []string) Params {
	return Params{
		EnabledProposalMsgUrls: enabledUrls,
	}
}

// DefaultParams returns the default parameters for community.
func DefaultParams() Params {
	return NewParams(
		DefaultEnabledProposalUrls,
	)
}

// Validate returns an error if the Params is invalid.
func (p *Params) Validate() error {
	return nil
}

func validateEnabledProposalMsgUrlsParam(i interface{}) error {
	_, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
