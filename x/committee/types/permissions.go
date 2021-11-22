package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	proto "github.com/gogo/protobuf/proto"
)

func init() {
	// CommitteeChange/Delete proposals are registered on gov's ModuleCdc (see proposal.go).
	// But since these proposals contain Permissions, these types also need registering:
	govtypes.ModuleCdc.RegisterInterface((*Permission)(nil), nil)
	govtypes.RegisterProposalTypeCodec(GodPermission{}, "kava/GodPermission")
	govtypes.RegisterProposalTypeCodec(TextPermission{}, "kava/TextPermission")
	govtypes.RegisterProposalTypeCodec(SoftwareUpgradePermission{}, "kava/SoftwareUpgradePermission")
	govtypes.RegisterProposalTypeCodec(ParamsChangePermission{}, "kava/ParamsChangePermission")
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(sdk.Context, codec.Codec, ParamKeeper, PubProposal) bool
}

func PackPermissions(permissions []Permission) ([]*types.Any, error) {
	permissionsAny := make([]*types.Any, len(permissions))
	for i, permission := range permissions {
		msg, ok := permission.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("cannot proto marshal %T", permission)
		}
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		permissionsAny[i] = any
	}
	return permissionsAny, nil
}

func UnpackPermissions(permissionsAny []*types.Any) ([]Permission, error) {
	permissions := make([]Permission, len(permissionsAny))
	for i, any := range permissionsAny {
		permission, ok := any.GetCachedValue().(Permission)
		if !ok {
			return nil, fmt.Errorf("expected base committee permission")
		}
		permissions[i] = permission
	}

	return permissions, nil
}

var (
	_ Permission = GodPermission{}
	_ Permission = TextPermission{}
	_ Permission = SoftwareUpgradePermission{}
	_ Permission = ParamsChangePermission{}
)

// Allows implement permission interface for GodPermission.
func (GodPermission) Allows(sdk.Context, codec.Codec, ParamKeeper, PubProposal) bool { return true }

// Allows implement permission interface for TextPermission.
func (TextPermission) Allows(_ sdk.Context, _ codec.Codec, _ ParamKeeper, p PubProposal) bool {
	_, ok := p.(*govtypes.TextProposal)
	return ok
}

// Allows implement permission interface for SoftwareUpgradePermission.
func (SoftwareUpgradePermission) Allows(_ sdk.Context, _ codec.Codec, _ ParamKeeper, p PubProposal) bool {
	_, ok := p.(*upgradetypes.SoftwareUpgradeProposal)
	return ok
}

// Allows implement permission interface for ParamsChangePermission.
func (perm ParamsChangePermission) Allows(_ sdk.Context, _ codec.Codec, _ ParamKeeper, p PubProposal) bool {
	proposal, ok := p.(*paramsproposal.ParameterChangeProposal)
	if !ok {
		return false
	}

	// Check if all proposal changes are allowed by this permission.
	for _, change := range proposal.Changes {
		targetedParamsChange := perm.AllowedParamsChanges.FilterByParamChange(change)

		// We allow the proposal param change if any of the targeted AllowedParamsChange allows it.
		// This enables us to have multiple different rules for the same subspace/key.
		// TODO: Note: This is here to support the sub parameter feature which is not yet implemented.
		allowed := false
		for _, pc := range targetedParamsChange {
			if pc.Allows(change) {
				allowed = true
				break
			}
		}

		// If no target param change allows the proposed change, then the proposal is rejected.
		if !allowed {
			return false
		}
	}

	return true
}

type AllowedParamsChanges []AllowedParamsChange

// FilterByParamChange returns all targeted AllowedParamsChange that matches a given ParamChange's subspace and key.
func (changes AllowedParamsChanges) FilterByParamChange(paramChange paramsproposal.ParamChange) AllowedParamsChanges {
	filtered := []AllowedParamsChange{}
	for _, p := range changes {
		if paramChange.Subspace == p.Subspace && paramChange.Key == p.Key {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// Allows returns true if the given proposal param change is allowed by the AllowedParamsChange rules.
func (allowed AllowedParamsChange) Allows(paramsChange paramsproposal.ParamChange) bool {
	// Check if param change matches target subspace and key.
	if allowed.Subspace != paramsChange.Subspace && allowed.Key != paramsChange.Key {
		return false
	}

	// TODO: Handle sub parameters and required sub param attr values

	return true
}
