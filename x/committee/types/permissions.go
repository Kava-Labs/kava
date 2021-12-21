package types

import (
	"encoding/json"
	fmt "fmt"
	"reflect"
	"strings"

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
	Allows(sdk.Context, ParamKeeper, PubProposal) bool
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
func (GodPermission) Allows(sdk.Context, ParamKeeper, PubProposal) bool { return true }

// Allows implement permission interface for TextPermission.
func (TextPermission) Allows(_ sdk.Context, _ ParamKeeper, p PubProposal) bool {
	_, ok := p.(*govtypes.TextProposal)
	return ok
}

// Allows implement permission interface for SoftwareUpgradePermission.
func (SoftwareUpgradePermission) Allows(_ sdk.Context, _ ParamKeeper, p PubProposal) bool {
	_, ok := p.(*upgradetypes.SoftwareUpgradeProposal)
	return ok
}

// Allows implement permission interface for ParamsChangePermission.
func (perm ParamsChangePermission) Allows(ctx sdk.Context, pk ParamKeeper, p PubProposal) bool {
	proposal, ok := p.(*paramsproposal.ParameterChangeProposal)
	if !ok {
		return false
	}

	// Check if all proposal changes are allowed by this permission.
	for _, change := range proposal.Changes {
		targetedParamsChange := perm.AllowedParamsChanges.filterByParamChange(change)

		// We allow the proposal param change if any of the targeted AllowedParamsChange allows it.
		// This give the option of having multiple rules for the same subspace/key if needed.
		allowed := false
		for _, pc := range targetedParamsChange {
			if pc.allowsParamChange(ctx, change, pk) {
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

// filterByParamChange returns all targeted AllowedParamsChange that matches a given ParamChange's subspace and key.
func (changes AllowedParamsChanges) filterByParamChange(paramChange paramsproposal.ParamChange) AllowedParamsChanges {
	filtered := []AllowedParamsChange{}
	for _, p := range changes {
		if paramChange.Subspace == p.Subspace && paramChange.Key == p.Key {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// SubparamChanges is a map of sub param change keys and its values.
type SubparamChanges map[string]interface{}

// MultiSubparamChanges is a slice of SubparamChanges.
type MultiSubparamChanges []SubparamChanges

func (allowed AllowedParamsChange) allowsMultiParamsChange(currentRecords MultiSubparamChanges, incomingRecords MultiSubparamChanges) bool {
	// do not allow new records from being added or removed for multi-subparam changes.
	if len(currentRecords) != len(incomingRecords) {
		return false
	}

	for _, current := range currentRecords {
		// find the incoming record and the requirements for each current record
		var req *SubparamRequirement
		var incoming *SubparamChanges
		for _, v := range allowed.MultiSubparamsRequirements {
			if current[v.Key] == v.Val {
				req = &v
				break
			}
		}

		// all records should have a requirement, otherwise the change is not allowed
		if req == nil {
			return false
		}

		for _, v := range incomingRecords {
			if v[req.Key] == req.Val {
				incoming = &v
				break
			}
		}

		// disallow the change if no incoming record found for current record.
		if incoming == nil {
			return false
		}

		// check incoming changes are allowed
		allowed := validateParamChangesAreAllowed(current, *incoming, req.AllowedSubparamAttrChanges)

		if !allowed {
			return false
		}
	}

	return true
}

func validateParamChangesAreAllowed(current SubparamChanges, incoming SubparamChanges, allowList []string) bool {
	// make sure we are not adding or removing any new attributes
	if len(current) != len(incoming) {
		return false
	}

	for k, v := range current {
		isAllowed := false

		// check if the param attr key is in the allow list
		for _, allowedKey := range allowList {
			if k == allowedKey {
				isAllowed = true
				break
			}
		}

		// if not allowed, incoming value needs to be the same, or it is rejected
		if !isAllowed {
			// since we cannot compare maps directly, we need to convert them to json first.
			// this should be fine since the data we are marshalling here should always be pretty small.
			if reflect.TypeOf(v).Kind() == reflect.Map {
				data, err := json.Marshal(v)
				if err != nil {
					return false
				}
				data2, err := json.Marshal(incoming[k])
				if err != nil {
					return false
				}
				if string(data) != string(data2) {
					return false
				}
			} else if v != incoming[k] {
				return false
			}
		}
	}

	return true
}

func (allowed AllowedParamsChange) allowsSingleParamsChange(current SubparamChanges, incoming SubparamChanges) bool {
	return validateParamChangesAreAllowed(current, incoming, allowed.SingleSubparamAllowedAttrs)
}

// allowsParamChange returns true if the given proposal param change is allowed by the AllowedParamsChange rules.
func (allowed AllowedParamsChange) allowsParamChange(ctx sdk.Context, paramsChange paramsproposal.ParamChange, pk ParamKeeper) bool {
	// Check if param change matches target subspace and key.
	if allowed.Subspace != paramsChange.Subspace && allowed.Key != paramsChange.Key {
		return false
	}

	// Check if param value is an array before unmarshalling to corresponding types
	tdata := strings.TrimLeft(paramsChange.Value, "\t\r\n")
	isArray := len(tdata) > 0 && tdata[0] == '['

	// Handle multi param value validation
	if isArray {
		var changeValue MultiSubparamChanges
		if err := json.Unmarshal([]byte(paramsChange.Value), &changeValue); err != nil {
			return false
		}

		var currentValue MultiSubparamChanges
		subspace, found := pk.GetSubspace(paramsChange.Subspace)
		if !found {
			return false
		}
		raw := subspace.GetRaw(ctx, []byte(paramsChange.Key))
		if err := json.Unmarshal(raw, &currentValue); err != nil {
			panic(err)
		}

		return allowed.allowsMultiParamsChange(currentValue, changeValue)
	}

	// Handle single param value validation
	var changeValue SubparamChanges
	if err := json.Unmarshal([]byte(paramsChange.Value), &changeValue); err != nil {
		return false
	}

	var currentValue SubparamChanges
	subspace, found := pk.GetSubspace(paramsChange.Subspace)
	if !found {
		return false
	}
	raw := subspace.GetRaw(ctx, []byte(paramsChange.Key))
	if err := json.Unmarshal(raw, &currentValue); err != nil {
		panic(err)
	}

	return allowed.allowsSingleParamsChange(currentValue, changeValue)
}
