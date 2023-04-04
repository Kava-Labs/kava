package types

import (
	"encoding/json"
	fmt "fmt"
	"reflect"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	proto "github.com/gogo/protobuf/proto"
)

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
	_, ok := p.(*govv1beta1.TextProposal)
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

// Get searches the allowedParamsChange slice for the first item matching a subspace and key.
// It returns false if not found.
func (changes AllowedParamsChanges) Get(subspace, key string) (AllowedParamsChange, bool) {
	for _, apc := range changes {
		if apc.Subspace == subspace && apc.Key == key {
			return apc, true
		}
	}
	return AllowedParamsChange{}, false
}

// Set adds a new AllowedParamsChange, overwriting the first exiting item with matching subspace and key.
func (changes *AllowedParamsChanges) Set(newChange AllowedParamsChange) {
	for i, apc := range *changes {
		if apc.Subspace == newChange.Subspace && apc.Key == newChange.Key {
			(*changes)[i] = newChange
			return
		}
	}
	*changes = append(*changes, newChange)
}

// Delete removes the first AllowedParamsChange matching subspace and key.
func (changes *AllowedParamsChanges) Delete(subspace, key string) {
	var found bool
	var foundAt int

	for i, apc := range *changes {
		if apc.Subspace == subspace && apc.Key == key {
			found = true
			foundAt = i
			break
		}
	}
	if !found {
		return
	}
	*changes = append(
		(*changes)[:foundAt],
		(*changes)[foundAt+1:]...,
	)
}

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

	// Warning: ranging over maps iterates through keys in a random order.
	// All state machine code must be deterministic between validators.
	// This function's output is deterministic despite the range.
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
		if !isAllowed && !reflect.DeepEqual(v, incoming[k]) {
			return false
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

	// Allow all param changes if no subparam rules are specified.
	if len(allowed.SingleSubparamAllowedAttrs) == 0 && len(allowed.MultiSubparamsRequirements) == 0 {
		return true
	}

	subspace, found := pk.GetSubspace(paramsChange.Subspace)
	if !found {
		return false
	}
	currentRaw := subspace.GetRaw(ctx, []byte(paramsChange.Key))

	// Check if current param value is an array before unmarshalling to corresponding types
	tdata := strings.TrimLeft(string(currentRaw), "\t\r\n")
	isArray := len(tdata) > 0 && tdata[0] == '['

	// Handle multi param value validation
	if isArray {
		var changeValue MultiSubparamChanges
		if err := json.Unmarshal([]byte(paramsChange.Value), &changeValue); err != nil {
			return false
		}

		var currentValue MultiSubparamChanges
		if err := json.Unmarshal(currentRaw, &currentValue); err != nil {
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
	if err := json.Unmarshal(currentRaw, &currentValue); err != nil {
		panic(err)
	}

	return allowed.allowsSingleParamsChange(currentValue, changeValue)
}
