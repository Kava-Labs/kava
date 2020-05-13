package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

func init() {
	// CommitteeChange/Delete proposals are registered on gov's ModuleCdc (see proposal.go).
	// But since these proposals contain Permissions, these types also need registering:
	govtypes.ModuleCdc.RegisterInterface((*Permission)(nil), nil)
	govtypes.RegisterProposalTypeCodec(GodPermission{}, "kava/GodPermission")
	govtypes.RegisterProposalTypeCodec(ParamChangePermission{}, "kava/ParamChangePermission")
	govtypes.RegisterProposalTypeCodec(TextPermission{}, "kava/TextPermission")
	govtypes.RegisterProposalTypeCodec(SubParamChangePermission{}, "kava/SubParamChangePermission")
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(sdk.Context, *codec.Codec, ParamKeeper, PubProposal) bool
}

// ------------------------------------------
//				GodPermission
// ------------------------------------------

// GodPermission allows any governance proposal. It is used mainly for testing.
type GodPermission struct{}

var _ Permission = GodPermission{}

func (GodPermission) Allows(sdk.Context, *codec.Codec, ParamKeeper, PubProposal) bool { return true }

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

func (perm ParamChangePermission) Allows(_ sdk.Context, _ *codec.Codec, _ ParamKeeper, p PubProposal) bool {
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

func (TextPermission) Allows(_ sdk.Context, _ *codec.Codec, _ ParamKeeper, p PubProposal) bool {
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

// ------------------------------------------
//				SubParamChangePermission // TODO change original to SimpleParamChangePermission
// ------------------------------------------

// ParamChangeProposal only allows changes to certain params
type SubParamChangePermission struct {
	AllowedParams           AllowedParams `json:"allowed_params" yaml:"allowed_params"`
	AllowedCollateralParams AllowedCollateralParams
	// DebtParams       AllowedDebtParams
	// SupportedAsset   AllowedSupportedAssets
	// Markets          AllowedMarkets
}

var _ Permission = SubParamChangePermission{}

func (perm SubParamChangePermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type          string        `yaml:"type"`
		AllowedParams AllowedParams `yaml:"allowed_params"`
		// TODO
	}{
		Type:          "param_change_permission",
		AllowedParams: perm.AllowedParams,
	}
	return valueToMarshal, nil
}

func (perm SubParamChangePermission) Allows(ctx sdk.Context, appCdc *codec.Codec, pk ParamKeeper, p PubProposal) bool {
	// Check pubproposal has correct type
	proposal, ok := p.(paramstypes.ParameterChangeProposal)
	if !ok {
		return false
	}
	// Check the param changes match the allowed keys
	for _, change := range proposal.Changes {
		if !perm.AllowedParams.Contains(change) {
			return false
		}
	}
	// Check any CollateralParam changes are allowed

	// Get the incoming CollaterParams value
	var foundIncomingCP bool
	var incomingCP cdptypes.CollateralParams
	for _, change := range proposal.Changes {
		if !(change.Subspace == cdptypes.ModuleName && change.Key == string(cdptypes.KeyCollateralParams)) {
			continue
		}
		// note: in case of duplicates take the last value
		foundIncomingCP = true
		if err := appCdc.UnmarshalJSON([]byte(change.Value), &incomingCP); err != nil {
			return false // invalid json value, so just disallow
		}
	}
	// only check if there was a proposed change
	if foundIncomingCP {
		// Get the current value of the CollateralParams
		cdpSubspace, found := pk.GetSubspace(cdptypes.ModuleName)
		if !found {
			panic(fmt.Sprintf("subspace doesn't exist: %s", cdptypes.ModuleName)) // TODO return false?
		}
		var currentCP cdptypes.CollateralParams
		// TODO byte type cast ok?
		cdpSubspace.Get(ctx, cdptypes.KeyCollateralParams, &currentCP) // panics if something goes wrong

		// Check all the incoming changes in the CollateralParams are allowed
		collateralParamChangesAllowed := perm.AllowedCollateralParams.Allows(currentCP, incomingCP)
		if !collateralParamChangesAllowed {
			return false
		}
	}

	// TODO repeat for DebtParams, SupportedAsset, Markets

	return true
}

type AllowedCollateralParams []AllowedCollateralParam

func (acps AllowedCollateralParams) Allows(incoming, current cdptypes.CollateralParams) bool {
	allAllowed := true

	// do not allow CollateralParams to be added or removed
	// this checks both lists are the same size, then below checks each incoming matches a current
	if len(incoming) != len(current) {
		return false
	}

	// for each param struct, check it is allowed, and if it is not, check the value has not changed
	for _, incomingCP := range incoming {
		// 1) check incoming cp is in list of allowed cps
		var foundAllowedCP bool
		var allowedCP AllowedCollateralParam
		for _, p := range acps {
			if p.Denom != incomingCP.Denom {
				continue
			}
			foundAllowedCP = true
			allowedCP = p
		}
		if !foundAllowedCP {
			// incoming had a CollateralParam that wasn't in the list of allowed ones
			// to add a CollateralParam it must explicitly be in the list of allowed params (with all fields set to true)
			return false
		}

		// 2) Check incoming changes are individually allowed
		// find existing CollateralParam
		var foundCurrentCP bool
		var currentCP cdptypes.CollateralParam
		for _, p := range current {
			if p.Denom != incomingCP.Denom {
				continue
			}
			foundCurrentCP = true
			currentCP = p
		}
		if !foundCurrentCP {
			return false // not allowed to add param to list
		}
		// check changed values are all allowed
		allowed := allowedCP.Allows(incomingCP, currentCP)

		allAllowed = allAllowed && allowed
	}
	return allAllowed
}

type AllowedCollateralParam struct {
	Denom              string `json:"denom" yaml:"denom"`
	LiquidationRatio   bool   `json:"liquidation_ratio" yaml:"liquidation_ratio"`
	DebtLimit          bool   `json:"debt_limit" yaml:"debt_limit"`
	StabilityFee       bool   `json:"stability_fee" yaml:"stability_fee"`
	AuctionSize        bool   `json:"auction_size" yaml:"auction_size"`
	LiquidationPenalty bool   `json:"liquidation_penalty" yaml:"liquidation_penalty"`
	Prefix             bool   `json:"prefix" yaml:"prefix"`
	MarketID           bool   `json:"market_id" yaml:"market_id"`
	ConversionFactor   bool   `json:"conversion_factor" yaml:"conversion_factor"`
}

func (acp AllowedCollateralParam) Allows(incoming, current cdptypes.CollateralParam) bool {
	allowed := (current.LiquidationRatio.Equal(incoming.LiquidationRatio) || acp.LiquidationRatio) &&
		(current.DebtLimit.IsEqual(incoming.DebtLimit) || acp.DebtLimit) &&
		(current.StabilityFee.Equal(incoming.StabilityFee) || acp.StabilityFee) &&
		(current.AuctionSize.Equal(incoming.AuctionSize) || acp.AuctionSize) &&
		(current.LiquidationPenalty.Equal(incoming.LiquidationPenalty) || acp.LiquidationPenalty) &&
		((current.Prefix == incoming.Prefix) || acp.Prefix) &&
		((current.MarketID == incoming.MarketID) || acp.MarketID) &&
		(current.ConversionFactor.Equal(incoming.ConversionFactor) || acp.ConversionFactor)
	return allowed
}
