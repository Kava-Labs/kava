package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade"

	bep3types "github.com/kava-labs/kava/x/bep3/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/pricefeed"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

func init() {
	// CommitteeChange/Delete proposals are registered on gov's ModuleCdc (see proposal.go).
	// But since these proposals contain Permissions, these types also need registering:
	govtypes.ModuleCdc.RegisterInterface((*Permission)(nil), nil)
	govtypes.RegisterProposalTypeCodec(GodPermission{}, "kava/GodPermission")
	govtypes.RegisterProposalTypeCodec(SimpleParamChangePermission{}, "kava/SimpleParamChangePermission")
	govtypes.RegisterProposalTypeCodec(TextPermission{}, "kava/TextPermission")
	govtypes.RegisterProposalTypeCodec(SoftwareUpgradePermission{}, "kava/SoftwareUpgradePermission")
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

// ------------------------------------------
//				SimpleParamChangePermission
// ------------------------------------------

// SimpleParamChangePermission only allows changes to certain params
type SimpleParamChangePermission struct {
	AllowedParams AllowedParams `json:"allowed_params" yaml:"allowed_params"`
}

var _ Permission = SimpleParamChangePermission{}

// Allows implement permission interface
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

// MarshalYAML implement yaml marshalling
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

// AllowedParam permission type for module parameter keys
type AllowedParam struct {
	Subspace string `json:"subspace" yaml:"subspace"`
	Key      string `json:"key" yaml:"key"`
}

// AllowedParams slice of AllowedParam
type AllowedParams []AllowedParam

// Contains checks if a key is included in param permissions
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

// Allows implement permission interface
func (TextPermission) Allows(_ sdk.Context, _ *codec.Codec, _ ParamKeeper, p PubProposal) bool {
	_, ok := p.(govtypes.TextProposal)
	return ok
}

// MarshalYAML implement yaml marshalling
func (TextPermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type string `yaml:"type"`
	}{
		Type: "text_permission",
	}
	return valueToMarshal, nil
}

// ------------------------------------------
//				SoftwareUpgradePermission
// ------------------------------------------

// SoftwareUpgradePermission permission type for software upgrade proposals
type SoftwareUpgradePermission struct{}

var _ Permission = SoftwareUpgradePermission{}

// Allows implement permission interface
func (SoftwareUpgradePermission) Allows(_ sdk.Context, _ *codec.Codec, _ ParamKeeper, p PubProposal) bool {
	_, ok := p.(upgrade.SoftwareUpgradeProposal)
	return ok
}

// MarshalYAML implement yaml marshalling
func (SoftwareUpgradePermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type string `yaml:"type"`
	}{
		Type: "software_upgrade_permission",
	}
	return valueToMarshal, nil
}

// ------------------------------------------
//				SubParamChangePermission
// ------------------------------------------

// SubParamChangePermission permission type for allowing changes to specific sub-keys within module parameter keys
type SubParamChangePermission struct {
	AllowedParams           AllowedParams           `json:"allowed_params" yaml:"allowed_params"`
	AllowedCollateralParams AllowedCollateralParams `json:"allowed_collateral_params" yaml:"allowed_collateral_params"`
	AllowedDebtParam        AllowedDebtParam        `json:"allowed_debt_param" yaml:"allowed_debt_param"`
	AllowedAssetParams      AllowedAssetParams      `json:"allowed_asset_params" yaml:"allowed_asset_params"`
	AllowedMarkets          AllowedMarkets          `json:"allowed_markets" yaml:"allowed_markets"`
	AllowedMoneyMarkets     AllowedMoneyMarkets     `json:"allowed_money_markets" yaml:"allowed_money_markets"`
}

var _ Permission = SubParamChangePermission{}

// MarshalYAML implement yaml marshalling
func (perm SubParamChangePermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type                    string                  `yaml:"type" json:"type"`
		AllowedParams           AllowedParams           `yaml:"allowed_params" json:"allowed_params"`
		AllowedCollateralParams AllowedCollateralParams `yaml:"allowed_collateral_params" json:"allowed_collateral_params"`
		AllowedDebtParam        AllowedDebtParam        `yaml:"allowed_debt_param" json:"allowed_debt_param"`
		AllowedAssetParams      AllowedAssetParams      `yaml:"allowed_asset_params" json:"allowed_asset_params"`
		AllowedMarkets          AllowedMarkets          `yaml:"allowed_markets" json:"allowed_markets"`
		AllowedMoneyMarkets     AllowedMoneyMarkets     `json:"allowed_money_markets" yaml:"allowed_money_markets"`
	}{
		Type:                    "param_change_permission",
		AllowedParams:           perm.AllowedParams,
		AllowedCollateralParams: perm.AllowedCollateralParams,
		AllowedDebtParam:        perm.AllowedDebtParam,
		AllowedAssetParams:      perm.AllowedAssetParams,
		AllowedMarkets:          perm.AllowedMarkets,
		AllowedMoneyMarkets:     perm.AllowedMoneyMarkets,
	}
	return valueToMarshal, nil
}

// Allows implement permission interface
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
			return false // not using a panic to help avoid begin blocker panics
		}
		var currentCP cdptypes.CollateralParams
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
			return false // not using a panic to help avoid begin blocker panics
		}
		var currentDP cdptypes.DebtParam
		cdpSubspace.Get(ctx, cdptypes.KeyDebtParam, &currentDP) // panics if something goes wrong

		// Check the incoming changes in the DebtParam are allowed
		debtParamChangeAllowed := perm.AllowedDebtParam.Allows(currentDP, incomingDP)
		if !debtParamChangeAllowed {
			return false
		}
	}

	// Check any AssetParams changes are allowed

	// Get the incoming AssetParams value
	var foundIncomingAPs bool
	var incomingAPs bep3types.AssetParams
	for _, change := range proposal.Changes {
		if !(change.Subspace == bep3types.ModuleName && change.Key == string(bep3types.KeyAssetParams)) {
			continue
		}
		// note: in case of duplicates take the last value
		foundIncomingAPs = true
		if err := appCdc.UnmarshalJSON([]byte(change.Value), &incomingAPs); err != nil {
			return false // invalid json value, so just disallow
		}
	}
	// only check if there was a proposed change
	if foundIncomingAPs {
		// Get the current value of the SupportedAssets
		subspace, found := pk.GetSubspace(bep3types.ModuleName)
		if !found {
			return false // not using a panic to help avoid begin blocker panics
		}
		var currentAPs bep3types.AssetParams
		subspace.Get(ctx, bep3types.KeyAssetParams, &currentAPs) // panics if something goes wrong

		// Check all the incoming changes in the CollateralParams are allowed
		assetParamsChangesAllowed := perm.AllowedAssetParams.Allows(currentAPs, incomingAPs)
		if !assetParamsChangesAllowed {
			return false
		}
	}

	// Check any Markets changes are allowed

	// Get the incoming Markets value
	var foundIncomingMs bool
	var incomingMs pricefeedtypes.Markets
	for _, change := range proposal.Changes {
		if !(change.Subspace == pricefeedtypes.ModuleName && change.Key == string(pricefeedtypes.KeyMarkets)) {
			continue
		}
		// note: in case of duplicates take the last value
		foundIncomingMs = true
		if err := appCdc.UnmarshalJSON([]byte(change.Value), &incomingMs); err != nil {
			return false // invalid json value, so just disallow
		}
	}
	// only check if there was a proposed change
	if foundIncomingMs {
		// Get the current value of the Markets
		subspace, found := pk.GetSubspace(pricefeedtypes.ModuleName)
		if !found {
			return false // not using a panic to help avoid begin blocker panics
		}
		var currentMs pricefeedtypes.Markets
		subspace.Get(ctx, pricefeedtypes.KeyMarkets, &currentMs) // panics if something goes wrong

		// Check all the incoming changes in the Markets are allowed
		marketsChangesAllowed := perm.AllowedMarkets.Allows(currentMs, incomingMs)
		if !marketsChangesAllowed {
			return false
		}
	}

	// Check any MoneyMarket changes are alloed

	var foundIncomingMMs bool
	var incomingMMs hard.MoneyMarkets
	for _, change := range proposal.Changes {
		if !(change.Subspace == hard.ModuleName && change.Key == string(hard.KeyMoneyMarkets)) {
			continue
		}
		foundIncomingMMs = true
		if err := appCdc.UnmarshalJSON([]byte(change.Value), &incomingMMs); err != nil {
			return false
		}
	}

	if foundIncomingMMs {
		subspace, found := pk.GetSubspace(hard.ModuleName)
		if !found {
			return false
		}
		var currentMMs hard.MoneyMarkets
		subspace.Get(ctx, hard.KeyMoneyMarkets, &currentMMs)
		mmChangesAllowed := perm.AllowedMoneyMarkets.Allows(currentMMs, incomingMMs)
		if !mmChangesAllowed {
			return false
		}
	}

	return true
}

// AllowedCollateralParams slice of AllowedCollateralParam
type AllowedCollateralParams []AllowedCollateralParam

// Allows determine if collateral params changes are permitted
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
			if p.Type != incomingCP.Type {
				continue
			}
			foundAllowedCP = true
			allowedCP = p
		}
		if !foundAllowedCP {
			// incoming had a CollateralParam that wasn't in the list of allowed ones
			return false
		}

		// 2) Check incoming changes are individually allowed
		// find existing CollateralParam
		var foundCurrentCP bool
		var currentCP cdptypes.CollateralParam
		for _, p := range current {
			if p.Type != incomingCP.Type {
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

// AllowedCollateralParam permission struct for changes to collateral parameter keys (cdp module)
type AllowedCollateralParam struct {
	Type                             string `json:"type" yaml:"type"`
	Denom                            bool   `json:"denom" yaml:"denom"`
	LiquidationRatio                 bool   `json:"liquidation_ratio" yaml:"liquidation_ratio"`
	DebtLimit                        bool   `json:"debt_limit" yaml:"debt_limit"`
	StabilityFee                     bool   `json:"stability_fee" yaml:"stability_fee"`
	AuctionSize                      bool   `json:"auction_size" yaml:"auction_size"`
	LiquidationPenalty               bool   `json:"liquidation_penalty" yaml:"liquidation_penalty"`
	Prefix                           bool   `json:"prefix" yaml:"prefix"`
	SpotMarketID                     bool   `json:"spot_market_id" yaml:"spot_market_id"`
	LiquidationMarketID              bool   `json:"liquidation_market_id" yaml:"liquidation_market_id"`
	ConversionFactor                 bool   `json:"conversion_factor" yaml:"conversion_factor"`
	KeeperRewardPercentage           bool   `json:"keeper_reward_percentage" yaml:"keeper_reward_percentage"`
	CheckCollateralizationIndexCount bool   `json:"check_collateralization_index_count" yaml:"check_collateralization_index_count"`
}

// NewAllowedCollateralParam return a new AllowedCollateralParam
func NewAllowedCollateralParam(
	ctype string, denom, liqRatio, debtLimit,
	stabilityFee, auctionSize, liquidationPenalty,
	prefix, spotMarket, liquidationMarket, conversionFactor, keeperReward, ltvIndexCount bool) AllowedCollateralParam {
	return AllowedCollateralParam{
		Type:                             ctype,
		Denom:                            denom,
		LiquidationRatio:                 liqRatio,
		DebtLimit:                        debtLimit,
		StabilityFee:                     stabilityFee,
		AuctionSize:                      auctionSize,
		LiquidationPenalty:               liquidationPenalty,
		Prefix:                           prefix,
		SpotMarketID:                     spotMarket,
		LiquidationMarketID:              liquidationMarket,
		ConversionFactor:                 conversionFactor,
		KeeperRewardPercentage:           keeperReward,
		CheckCollateralizationIndexCount: ltvIndexCount,
	}
}

// Allows determine if collateral param changes are permitted
func (acp AllowedCollateralParam) Allows(current, incoming cdptypes.CollateralParam) bool {
	allowed := ((acp.Type == current.Type) && (acp.Type == incoming.Type)) && // require collateral types to be all equal
		(current.Denom == incoming.Denom || acp.Denom) &&
		(current.LiquidationRatio.Equal(incoming.LiquidationRatio) || acp.LiquidationRatio) &&
		(current.DebtLimit.IsEqual(incoming.DebtLimit) || acp.DebtLimit) &&
		(current.StabilityFee.Equal(incoming.StabilityFee) || acp.StabilityFee) &&
		(current.AuctionSize.Equal(incoming.AuctionSize) || acp.AuctionSize) &&
		(current.LiquidationPenalty.Equal(incoming.LiquidationPenalty) || acp.LiquidationPenalty) &&
		((current.Prefix == incoming.Prefix) || acp.Prefix) &&
		((current.SpotMarketID == incoming.SpotMarketID) || acp.SpotMarketID) &&
		((current.LiquidationMarketID == incoming.LiquidationMarketID) || acp.LiquidationMarketID) &&
		((current.KeeperRewardPercentage.Equal(incoming.KeeperRewardPercentage)) || acp.KeeperRewardPercentage) &&
		((current.CheckCollateralizationIndexCount.Equal(incoming.CheckCollateralizationIndexCount)) || acp.CheckCollateralizationIndexCount) &&
		(current.ConversionFactor.Equal(incoming.ConversionFactor) || acp.ConversionFactor)
	return allowed
}

// AllowedDebtParam permission struct for changes to debt parameter keys (cdp module)
type AllowedDebtParam struct {
	Denom            bool `json:"denom" yaml:"denom"`
	ReferenceAsset   bool `json:"reference_asset" yaml:"reference_asset"`
	ConversionFactor bool `json:"conversion_factor" yaml:"conversion_factor"`
	DebtFloor        bool `json:"debt_floor" yaml:"debt_floor"`
}

// Allows determines if debt params changes are permitted
func (adp AllowedDebtParam) Allows(current, incoming cdptypes.DebtParam) bool {
	allowed := ((current.Denom == incoming.Denom) || adp.Denom) &&
		((current.ReferenceAsset == incoming.ReferenceAsset) || adp.ReferenceAsset) &&
		(current.ConversionFactor.Equal(incoming.ConversionFactor) || adp.ConversionFactor) &&
		(current.DebtFloor.Equal(incoming.DebtFloor) || adp.DebtFloor)
	return allowed
}

// AllowedAssetParams slice of AllowedAssetParam
type AllowedAssetParams []AllowedAssetParam

// Allows determines if asset params changes are permitted
func (aaps AllowedAssetParams) Allows(current, incoming bep3types.AssetParams) bool {
	allAllowed := true

	// do not allow AssetParams to be added or removed
	// this checks both lists are the same size, then below checks each incoming matches a current
	if len(incoming) != len(current) {
		return false
	}

	// for each asset struct, check it is allowed, and if it is not, check the value has not changed
	for _, incomingAP := range incoming {
		// 1) check incoming ap is in list of allowed aps
		var foundAllowedAP bool
		var allowedAP AllowedAssetParam
		for _, p := range aaps {
			if p.Denom != incomingAP.Denom {
				continue
			}
			foundAllowedAP = true
			allowedAP = p
		}
		if !foundAllowedAP {
			// incoming had a AssetParam that wasn't in the list of allowed ones
			return false
		}

		// 2) Check incoming changes are individually allowed
		// find existing SupportedAsset
		var foundCurrentAP bool
		var currentAP bep3types.AssetParam
		for _, p := range current {
			if p.Denom != incomingAP.Denom {
				continue
			}
			foundCurrentAP = true
			currentAP = p
		}
		if !foundCurrentAP {
			return false // not allowed to add asset to list
		}
		// check changed values are all allowed
		allowed := allowedAP.Allows(currentAP, incomingAP)

		allAllowed = allAllowed && allowed
	}
	return allAllowed
}

// AllowedAssetParam bep3 asset parameters that can be changed by committee
type AllowedAssetParam struct {
	Denom         string `json:"denom" yaml:"denom"`
	CoinID        bool   `json:"coin_id" yaml:"coin_id"`
	Limit         bool   `json:"limit" yaml:"limit"`
	Active        bool   `json:"active" yaml:"active"`
	MaxSwapAmount bool   `json:"max_swap_amount" yaml:"max_swap_amount"`
	MinBlockLock  bool   `json:"min_block_lock" yaml:"min_block_lock"`
}

// Allows bep3 AssetParam parameters than can be changed by committee
func (aap AllowedAssetParam) Allows(current, incoming bep3types.AssetParam) bool {

	allowed := ((aap.Denom == current.Denom) && (aap.Denom == incoming.Denom)) && // require denoms to be all equal
		((current.CoinID == incoming.CoinID) || aap.CoinID) &&
		(current.SupplyLimit.Equals(incoming.SupplyLimit) || aap.Limit) &&
		((current.Active == incoming.Active) || aap.Active) &&
		((current.MaxSwapAmount.Equal(incoming.MaxSwapAmount)) || aap.MaxSwapAmount) &&
		((current.MinBlockLock == incoming.MinBlockLock) || aap.MinBlockLock)
	return allowed
}

// AllowedMarkets slice of AllowedMarket
type AllowedMarkets []AllowedMarket

// Allows determines if markets params changed are permitted
func (ams AllowedMarkets) Allows(current, incoming pricefeedtypes.Markets) bool {
	allAllowed := true

	// do not allow Markets to be added or removed
	// this checks both lists are the same size, then below checks each incoming matches a current
	if len(incoming) != len(current) {
		return false
	}

	// for each market struct, check it is allowed, and if it is not, check the value has not changed
	for _, incomingM := range incoming {
		// 1) check incoming market is in list of allowed markets
		var foundAllowedM bool
		var allowedM AllowedMarket
		for _, p := range ams {
			if p.MarketID != incomingM.MarketID {
				continue
			}
			foundAllowedM = true
			allowedM = p
		}
		if !foundAllowedM {
			// incoming had a Market that wasn't in the list of allowed ones
			return false
		}

		// 2) Check incoming changes are individually allowed
		// find existing SupportedAsset
		var foundCurrentM bool
		var currentM pricefeed.Market
		for _, p := range current {
			if p.MarketID != incomingM.MarketID {
				continue
			}
			foundCurrentM = true
			currentM = p
		}
		if !foundCurrentM {
			return false // not allowed to add market to list
		}
		// check changed values are all allowed
		allowed := allowedM.Allows(currentM, incomingM)

		allAllowed = allAllowed && allowed
	}
	return allAllowed
}

// AllowedMarket permission struct for market parameters (pricefeed module)
type AllowedMarket struct {
	MarketID   string `json:"market_id" yaml:"market_id"`
	BaseAsset  bool   `json:"base_asset" yaml:"base_asset"`
	QuoteAsset bool   `json:"quote_asset" yaml:"quote_asset"`
	Oracles    bool   `json:"oracles" yaml:"oracles"`
	Active     bool   `json:"active" yaml:"active"`
}

// Allows determines if market param changes are permitted
func (am AllowedMarket) Allows(current, incoming pricefeedtypes.Market) bool {
	allowed := ((am.MarketID == current.MarketID) && (am.MarketID == incoming.MarketID)) && // require denoms to be all equal
		((current.BaseAsset == incoming.BaseAsset) || am.BaseAsset) &&
		((current.QuoteAsset == incoming.QuoteAsset) || am.QuoteAsset) &&
		(addressesEqual(current.Oracles, incoming.Oracles) || am.Oracles) &&
		((current.Active == incoming.Active) || am.Active)
	return allowed
}

// addressesEqual check if slices of addresses are equal, the order matters
func addressesEqual(addrs1, addrs2 []sdk.AccAddress) bool {
	if len(addrs1) != len(addrs2) {
		return false
	}
	areEqual := true
	for i := range addrs1 {
		areEqual = areEqual && addrs1[i].Equals(addrs2[i])
	}
	return areEqual
}

// AllowedMoneyMarket permission struct for money market parameters (hard module)
type AllowedMoneyMarket struct {
	Denom                  string `json:"denom" yaml:"denom"`
	BorrowLimit            bool   `json:"borrow_limit" yaml:"borrow_limit"`
	SpotMarketID           bool   `json:"spot_market_id" yaml:"spot_market_id"`
	ConversionFactor       bool   `json:"conversion_factor" yaml:"conversion_factor"`
	InterestRateModel      bool   `json:"interest_rate_model" yaml:"interest_rate_model"`
	ReserveFactor          bool   `json:"reserve_factor" yaml:"reserve_factor"`
	KeeperRewardPercentage bool   `json:"keeper_reward_percentage" yaml:"keeper_reward_percentage"`
}

// NewAllowedMoneyMarket returns a new AllowedMoneyMarket
func NewAllowedMoneyMarket(denom string, bl, sm, cf, irm, rf, kr bool) AllowedMoneyMarket {
	return AllowedMoneyMarket{
		Denom:                  denom,
		BorrowLimit:            bl,
		SpotMarketID:           sm,
		ConversionFactor:       cf,
		InterestRateModel:      irm,
		ReserveFactor:          rf,
		KeeperRewardPercentage: kr,
	}
}

// Allows determines if money market param changes are permitted
func (amm AllowedMoneyMarket) Allows(current, incoming hard.MoneyMarket) bool {
	allowed := ((amm.Denom == current.Denom) && (amm.Denom == incoming.Denom)) &&
		((current.BorrowLimit.Equal(incoming.BorrowLimit)) || amm.BorrowLimit) &&
		((current.SpotMarketID == incoming.SpotMarketID) || amm.SpotMarketID) &&
		((current.ConversionFactor.Equal(incoming.ConversionFactor)) || amm.ConversionFactor) &&
		((current.InterestRateModel.Equal(incoming.InterestRateModel)) || amm.InterestRateModel) &&
		((current.ReserveFactor.Equal(incoming.ReserveFactor)) || amm.ReserveFactor) &&
		((current.KeeperRewardPercentage.Equal(incoming.KeeperRewardPercentage)) || amm.KeeperRewardPercentage)
	return allowed
}

// AllowedMoneyMarkets slice of AllowedMoneyMarket
type AllowedMoneyMarkets []AllowedMoneyMarket

// Allows determins if money market params changes are permitted
func (amms AllowedMoneyMarkets) Allows(current, incoming hard.MoneyMarkets) bool {
	allAllowed := true

	if len(incoming) != len(current) {
		return false
	}

	for _, incomingMM := range incoming {
		var foundAllowedMM bool
		var allowedMM AllowedMoneyMarket

		for _, p := range amms {
			if p.Denom != incomingMM.Denom {
				continue
			}
			foundAllowedMM = true
			allowedMM = p
		}
		if !foundAllowedMM {
			return false
		}

		var foundCurrentMM bool
		var currentMM hard.MoneyMarket

		for _, p := range current {
			if p.Denom != incomingMM.Denom {
				continue
			}
			foundCurrentMM = true
			currentMM = p
		}
		if !foundCurrentMM {
			return false
		}
		allowed := allowedMM.Allows(currentMM, incomingMM)
		allAllowed = allAllowed && allowed
	}

	return allAllowed
}
