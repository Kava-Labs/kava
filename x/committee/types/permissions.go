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
	govtypes.RegisterProposalTypeCodec(SimpleParamChangePermission{}, "kava/SimpleParamChangePermission")
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
//				SimpleParamChangePermission
// ------------------------------------------

// SimpleParamChangePermission only allows changes to certain params
type SimpleParamChangePermission struct {
	AllowedParams AllowedParams `json:"allowed_params" yaml:"allowed_params"`
}

var _ Permission = SimpleParamChangePermission{}

func (perm SimpleParamChangePermission) Allows(_ sdk.Context, _ *codec.Codec, _ ParamKeeper, p PubProposal) bool {
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

func (perm SimpleParamChangePermission) MarshalYAML() (interface{}, error) {
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
//				SubParamChangePermission
// ------------------------------------------

// ParamChangeProposal only allows changes to certain params
type SubParamChangePermission struct {
	AllowedParams           AllowedParams           `json:"allowed_params" yaml:"allowed_params"`
	AllowedCollateralParams AllowedCollateralParams `json:"allowed_collateral_params" yaml:"allowed_collateral_params"`
	AllowedDebtParam        AllowedDebtParam        `json:"allowed_debt_param" yaml:"allowed_debt_param"`
	//AllowedSupportedAsset   AllowedSupportedAssets  `json:"allowed_supported_assets" yaml:"allowed_supported_assets"`
	//AllowedMarkets          AllowedMarkets          `json:"allowed_markets" yaml:"allowed_markets"`
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

	// Check any DebtParam changes are allowed

	// Get the incoming DebtParam value
	var foundIncomingDP bool
	var incomingDP cdptypes.DebtParam
	for _, change := range proposal.Changes {
		if !(change.Subspace == cdptypes.ModuleName && change.Key == string(cdptypes.KeyDebtParam)) {
			continue
		}
		// note: in case of duplicates take the last value
		foundIncomingDP = true
		if err := appCdc.UnmarshalJSON([]byte(change.Value), &incomingDP); err != nil {
			return false // invalid json value, so just disallow
		}
	}
	// only check if there was a proposed change
	if foundIncomingDP {
		// Get the current value of the DebtParams
		cdpSubspace, found := pk.GetSubspace(cdptypes.ModuleName)
		if !found {
			panic(fmt.Sprintf("subspace doesn't exist: %s", cdptypes.ModuleName)) // TODO return false?
		}
		var currentDP cdptypes.DebtParam
		cdpSubspace.Get(ctx, cdptypes.KeyDebtParam, &currentDP) // panics if something goes wrong

		// Check the incoming changes in the DebtParam are allowed
		debtParamChangeAllowed := perm.AllowedDebtParam.Allows(currentDP, incomingDP)
		if !debtParamChangeAllowed {
			return false
		}
	}
	// TODO these could be abstracted into one function - the types could be passed in.

	// TODO repeat for SupportedAsset, Markets

	return true
}

type AllowedCollateralParams []AllowedCollateralParam

func (acps AllowedCollateralParams) Allows(current, incoming cdptypes.CollateralParams) bool {
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
		allowed := allowedCP.Allows(currentCP, incomingCP)

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

func (acp AllowedCollateralParam) Allows(current, incoming cdptypes.CollateralParam) bool {
	allowed := ((acp.Denom == current.Denom) && (acp.Denom == incoming.Denom)) && // require denoms to be all equal
		(current.LiquidationRatio.Equal(incoming.LiquidationRatio) || acp.LiquidationRatio) &&
		(current.DebtLimit.IsEqual(incoming.DebtLimit) || acp.DebtLimit) &&
		(current.StabilityFee.Equal(incoming.StabilityFee) || acp.StabilityFee) &&
		(current.AuctionSize.Equal(incoming.AuctionSize) || acp.AuctionSize) &&
		(current.LiquidationPenalty.Equal(incoming.LiquidationPenalty) || acp.LiquidationPenalty) &&
		((current.Prefix == incoming.Prefix) || acp.Prefix) &&
		((current.MarketID == incoming.MarketID) || acp.MarketID) &&
		(current.ConversionFactor.Equal(incoming.ConversionFactor) || acp.ConversionFactor)
	return allowed
}

type AllowedDebtParam struct {
	Denom            bool `json:"denom" yaml:"denom"`
	ReferenceAsset   bool `json:"reference_asset" yaml:"reference_asset"`
	ConversionFactor bool `json:"conversion_factor" yaml:"conversion_factor"`
	DebtFloor        bool `json:"debt_floor" yaml:"debt_floor"`
	SavingsRate      bool `json:"savings_rate" yaml:"savings_rate"`
}

func (adp AllowedDebtParam) Allows(current, incoming cdptypes.DebtParam) bool {
	allowed := ((current.Denom == incoming.Denom) || adp.Denom) &&
		((current.ReferenceAsset == incoming.ReferenceAsset) || adp.ReferenceAsset) &&
		(current.ConversionFactor.Equal(incoming.ConversionFactor) || adp.ConversionFactor) &&
		(current.DebtFloor.Equal(incoming.DebtFloor) || adp.DebtFloor) &&
		(current.SavingsRate.Equal(incoming.SavingsRate) || adp.SavingsRate)
	return allowed
}
