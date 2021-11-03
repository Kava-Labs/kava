package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proto "github.com/gogo/protobuf/proto"
)

func init() {
	// CommitteeChange/Delete proposals are registered on gov's ModuleCdc (see proposal.go).
	// But since these proposals contain Permissions, these types also need registering:
	govtypes.ModuleCdc.RegisterInterface((*Permission)(nil), nil)
	// govtypes.RegisterProposalTypeCodec(GodPermission{}, "kava/GodPermission")
	// govtypes.RegisterProposalTypeCodec(SimpleParamChangePermission{}, "kava/SimpleParamChangePermission")
	// govtypes.RegisterProposalTypeCodec(TextPermission{}, "kava/TextPermission")
	// govtypes.RegisterProposalTypeCodec(SoftwareUpgradePermission{}, "kava/SoftwareUpgradePermission")
	// govtypes.RegisterProposalTypeCodec(SubParamChangePermission{}, "kava/SubParamChangePermission")
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(sdk.Context, *codec.Codec, ParamKeeper, PubProposal) bool
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
			panic(err)
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

var _ Permission = GodPermission{}

// Allows implement permission interface
func (GodPermission) Allows(sdk.Context, *codec.Codec, ParamKeeper, PubProposal) bool { return true }

// MarshalYAML implement yaml marshalling
func (GodPermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type string `yaml:"type"`
	}{
		Type: "god_permission",
	}
	return valueToMarshal, nil
}
