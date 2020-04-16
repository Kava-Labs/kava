package types

import (
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
)

func init() {
	// CommitteeChange/Delete proposals need to be registered on gov's ModuleCdc.
	// But since these proposals contain Permissions, these types also need registering:
	gov.ModuleCdc.RegisterInterface((*Permission)(nil), nil)
	gov.RegisterProposalTypeCodec(GodPermission{}, "kava/GodPermission")
	gov.RegisterProposalTypeCodec(ParamChangePermission{}, "kava/ParamChangePermission")
	gov.RegisterProposalTypeCodec(TextPermission{}, "kava/TextPermission")
}

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

// ParamChangeProposal only allows changes to certain params
type ParamChangePermission struct {
	AllowedParams AllowedParams `json:"allowed_params" yaml:"allowed_params"`
}

var _ Permission = ParamChangePermission{}

func (perm ParamChangePermission) Allows(p PubProposal) bool {
	proposal, ok := p.(params.ParameterChangeProposal)
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
		AllowedParams AllowedParams `yaml:"allowed_params`
	}{
		Type:          "param_change_permission",
		AllowedParams: perm.AllowedParams,
	}
	return valueToMarshal, nil
}

type AllowedParam struct {
	Subspace string `json:"subspace" yaml:"subspace"`
	Key      string `json:"key" yaml:"key"`
	Subkey   string `json:"subkey,omitempty" yaml:"subkey,omitempty"`
}
type AllowedParams []AllowedParam

func (allowed AllowedParams) Contains(paramChange params.ParamChange) bool {
	for _, p := range allowed {
		if paramChange.Subspace == p.Subspace && paramChange.Key == p.Key && paramChange.Subkey == p.Subkey {
			return true
		}
	}
	return false
}

type TextPermission struct{}

var _ Permission = TextPermission{}

func (TextPermission) Allows(p PubProposal) bool {
	_, ok := p.(gov.TextProposal)
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

// TODO add more permissions?
// - limit parameter changes to be within small ranges
// - allow community spend proposals
// - allow committee change proposals
