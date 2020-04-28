package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func init() {
	// CommitteeChange/Delete proposals are registered on gov's ModuleCdc (see proposal.go).
	// But since these proposals contain Permissions, these types also need registering:
	govtypes.ModuleCdc.RegisterInterface((*Permission)(nil), nil)
	govtypes.RegisterProposalTypeCodec(GodPermission{}, "kava/GodPermission")
	govtypes.RegisterProposalTypeCodec(ParamChangePermission{}, "kava/ParamChangePermission")
	govtypes.RegisterProposalTypeCodec(TextPermission{}, "kava/TextPermission")
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(PubProposal) bool
}

// ------------------------------------------
//				GodPermission
// ------------------------------------------

// GodPermission allows any governance proposal. It is used mainly for testing.
type GodPermission struct{}

var _ Permission = GodPermission{}

func (GodPermission) Allows(PubProposal) bool { return true }

func (GodPermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type string `yaml:"type"`
	}{
		Type: "god_permission",
	}
	return valueToMarshal, nil
}

// ------------------------------------------
//				ParamChangePermission
// ------------------------------------------

// ParamChangeProposal only allows changes to certain params
type ParamChangePermission struct {
	AllowedParams AllowedParams `json:"allowed_params" yaml:"allowed_params"`
}

var _ Permission = ParamChangePermission{}

func (perm ParamChangePermission) Allows(p PubProposal) bool {
	proposal, ok := p.(paramstypes.ParameterChangeProposal)
	if !ok {
		return false
	}
	for _, change := range proposal.Changes {
		if !perm.AllowedParams.Contains(change) {
			return false
		}
	}
	return true
}

func (perm ParamChangePermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type          string        `yaml:"type"`
		AllowedParams AllowedParams `yaml:"allowed_params"`
	}{
		Type:          "param_change_permission",
		AllowedParams: perm.AllowedParams,
	}
	return valueToMarshal, nil
}

type AllowedParam struct {
	Subspace string `json:"subspace" yaml:"subspace"`
	Key      string `json:"key" yaml:"key"`
}
type AllowedParams []AllowedParam

func (allowed AllowedParams) Contains(paramChange paramstypes.ParamChange) bool {
	for _, p := range allowed {
		if paramChange.Subspace == p.Subspace && paramChange.Key == p.Key {
			return true
		}
	}
	return false
}

// ------------------------------------------
//				TextPermission
// ------------------------------------------

// TextPermission allows any text governance proposal.
type TextPermission struct{}

var _ Permission = TextPermission{}

func (TextPermission) Allows(p PubProposal) bool {
	_, ok := p.(govtypes.TextProposal)
	return ok
}

func (TextPermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type string `yaml:"type"`
	}{
		Type: "text_permission",
	}
	return valueToMarshal, nil
}
