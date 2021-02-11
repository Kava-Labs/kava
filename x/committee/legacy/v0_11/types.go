package v0_11

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade"

	bep3types "github.com/kava-labs/kava/x/bep3/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
	"github.com/kava-labs/kava/x/pricefeed"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

const (
	MaxCommitteeDescriptionLength int = 512
)

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(sdk.Context, *codec.Codec, ParamKeeper, PubProposal) bool
}

type ParamKeeper interface {
	GetSubspace(string) (params.Subspace, bool)
}

// A Committee is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Committee struct {
	ID               uint64           `json:"id" yaml:"id"`
	Description      string           `json:"description" yaml:"description"`
	Members          []sdk.AccAddress `json:"members" yaml:"members"`
	Permissions      []Permission     `json:"permissions" yaml:"permissions"`
	VoteThreshold    sdk.Dec          `json:"vote_threshold" yaml:"vote_threshold"`       // Smallest percentage of members that must vote for a proposal to pass.
	ProposalDuration time.Duration    `json:"proposal_duration" yaml:"proposal_duration"` // The length of time a proposal remains active for. Proposals will close earlier if they get enough votes.
}

func NewCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission, threshold sdk.Dec, duration time.Duration) Committee {
	return Committee{
		ID:               id,
		Description:      description,
		Members:          members,
		Permissions:      permissions,
		VoteThreshold:    threshold,
		ProposalDuration: duration,
	}
}

func (c Committee) HasMember(addr sdk.AccAddress) bool {
	for _, m := range c.Members {
		if m.Equals(addr) {
			return true
		}
	}
	return false
}

// HasPermissionsFor returns whether the committee is authorized to enact a proposal.
// As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
func (c Committee) HasPermissionsFor(ctx sdk.Context, appCdc *codec.Codec, pk ParamKeeper, proposal PubProposal) bool {
	for _, p := range c.Permissions {
		if p.Allows(ctx, appCdc, pk, proposal) {
			return true
		}
	}
	return false
}

func (c Committee) Validate() error {

	addressMap := make(map[string]bool, len(c.Members))
	for _, m := range c.Members {
		// check there are no duplicate members
		if _, ok := addressMap[m.String()]; ok {
			return fmt.Errorf("committe cannot have duplicate members, %s", m)
		}
		// check for valid addresses
		if m.Empty() {
			return fmt.Errorf("committee cannot have empty member address")
		}
		addressMap[m.String()] = true
	}

	if len(c.Members) == 0 {
		return fmt.Errorf("committee cannot have zero members")
	}

	if len(c.Description) > MaxCommitteeDescriptionLength {
		return fmt.Errorf("description length %d longer than max allowed %d", len(c.Description), MaxCommitteeDescriptionLength)
	}

	for _, p := range c.Permissions {
		if p == nil {
			return fmt.Errorf("committee cannot have a nil permission")
		}
	}

	// threshold must be in the range (0,1]
	if c.VoteThreshold.IsNil() || c.VoteThreshold.LTE(sdk.ZeroDec()) || c.VoteThreshold.GT(sdk.NewDec(1)) {
		return fmt.Errorf("invalid threshold: %s", c.VoteThreshold)
	}

	if c.ProposalDuration < 0 {
		return fmt.Errorf("invalid proposal duration: %s", c.ProposalDuration)
	}

	return nil
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

// PubProposal is the interface that all proposals must fulfill to be submitted to a committee.
// Proposal types can be created external to this module. For example a ParamChangeProposal, or CommunityPoolSpendProposal.
// It is pinned to the equivalent type in the gov module to create compatibility between proposal types.
type PubProposal govtypes.Content

// Proposal is an internal record of a governance proposal submitted to a committee.
type Proposal struct {
	PubProposal `json:"pub_proposal" yaml:"pub_proposal"`
	ID          uint64    `json:"id" yaml:"id"`
	CommitteeID uint64    `json:"committee_id" yaml:"committee_id"`
	Deadline    time.Time `json:"deadline" yaml:"deadline"`
}

func NewProposal(pubProposal PubProposal, id uint64, committeeID uint64, deadline time.Time) Proposal {
	return Proposal{
		PubProposal: pubProposal,
		ID:          id,
		CommitteeID: committeeID,
		Deadline:    deadline,
	}
}

// HasExpiredBy calculates if the proposal will have expired by a certain time.
// All votes must be cast before deadline, those cast at time == deadline are not valid
func (p Proposal) HasExpiredBy(time time.Time) bool {
	return !time.Before(p.Deadline)
}

// String implements the fmt.Stringer interface, and importantly overrides the String methods inherited from the embedded PubProposal type.
func (p Proposal) String() string {
	bz, _ := yaml.Marshal(p)
	return string(bz)
}

// ------------------------------------------
//				Votes
// ------------------------------------------

type Vote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
}

func NewVote(proposalID uint64, voter sdk.AccAddress) Vote {
	return Vote{
		ProposalID: proposalID,
		Voter:      voter,
	}
}

func (v Vote) Validate() error {
	if v.Voter.Empty() {
		return fmt.Errorf("voter address cannot be empty")
	}
	return nil
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
//				SoftwareUpgradePermission
// ------------------------------------------

type SoftwareUpgradePermission struct{}

var _ Permission = SoftwareUpgradePermission{}

func (SoftwareUpgradePermission) Allows(_ sdk.Context, _ *codec.Codec, _ ParamKeeper, p PubProposal) bool {
	_, ok := p.(upgrade.SoftwareUpgradeProposal)
	return ok
}

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

// ParamChangeProposal only allows changes to certain params
type SubParamChangePermission struct {
	AllowedParams           AllowedParams           `json:"allowed_params" yaml:"allowed_params"`
	AllowedCollateralParams AllowedCollateralParams `json:"allowed_collateral_params" yaml:"allowed_collateral_params"`
	AllowedDebtParam        AllowedDebtParam        `json:"allowed_debt_param" yaml:"allowed_debt_param"`
	AllowedAssetParams      AllowedAssetParams      `json:"allowed_asset_params" yaml:"allowed_asset_params"`
	AllowedMarkets          AllowedMarkets          `json:"allowed_markets" yaml:"allowed_markets"`
}

var _ Permission = SubParamChangePermission{}

func (perm SubParamChangePermission) MarshalYAML() (interface{}, error) {
	valueToMarshal := struct {
		Type                    string                  `yaml:"type"`
		AllowedParams           AllowedParams           `yaml:"allowed_params"`
		AllowedCollateralParams AllowedCollateralParams `yaml:"allowed_collateral_params"`
		AllowedDebtParam        AllowedDebtParam        `yaml:"allowed_debt_param"`
		AllowedAssetParams      AllowedAssetParams      `yaml:"allowed_asset_params"`
		AllowedMarkets          AllowedMarkets          `yaml:"allowed_markets"`
	}{
		Type:                    "param_change_permission",
		AllowedParams:           perm.AllowedParams,
		AllowedCollateralParams: perm.AllowedCollateralParams,
		AllowedDebtParam:        perm.AllowedDebtParam,
		AllowedAssetParams:      perm.AllowedAssetParams,
		AllowedMarkets:          perm.AllowedMarkets,
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

	return true
}

// AllowedCollateralParam cdp.CollateralParam fields that can be subject to committee governance
type AllowedCollateralParam struct {
	Type                string `json:"type" yaml:"type"`
	Denom               bool   `json:"denom" yaml:"denom"`
	LiquidationRatio    bool   `json:"liquidation_ratio" yaml:"liquidation_ratio"`
	DebtLimit           bool   `json:"debt_limit" yaml:"debt_limit"`
	StabilityFee        bool   `json:"stability_fee" yaml:"stability_fee"`
	AuctionSize         bool   `json:"auction_size" yaml:"auction_size"`
	LiquidationPenalty  bool   `json:"liquidation_penalty" yaml:"liquidation_penalty"`
	Prefix              bool   `json:"prefix" yaml:"prefix"`
	SpotMarketID        bool   `json:"spot_market_id" yaml:"spot_market_id"`
	LiquidationMarketID bool   `json:"liquidation_market_id" yaml:"liquidation_market_id"`
	ConversionFactor    bool   `json:"conversion_factor" yaml:"conversion_factor"`
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

func (acp AllowedCollateralParam) Allows(current, incoming cdptypes.CollateralParam) bool {
	allowed := ((acp.Type == current.Type) && (acp.Type == incoming.Type)) && // require collatreral types to be all equal
		(current.Denom == incoming.Denom || acp.Denom) &&
		(current.LiquidationRatio.Equal(incoming.LiquidationRatio) || acp.LiquidationRatio) &&
		(current.DebtLimit.IsEqual(incoming.DebtLimit) || acp.DebtLimit) &&
		(current.StabilityFee.Equal(incoming.StabilityFee) || acp.StabilityFee) &&
		(current.AuctionSize.Equal(incoming.AuctionSize) || acp.AuctionSize) &&
		(current.LiquidationPenalty.Equal(incoming.LiquidationPenalty) || acp.LiquidationPenalty) &&
		((current.Prefix == incoming.Prefix) || acp.Prefix) &&
		((current.SpotMarketID == incoming.SpotMarketID) || acp.SpotMarketID) &&
		((current.LiquidationMarketID == incoming.LiquidationMarketID) || acp.LiquidationMarketID) &&
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

type AllowedAssetParams []AllowedAssetParam

// Allows implement permission interface
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

type AllowedMarkets []AllowedMarket

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

type AllowedMarket struct {
	MarketID   string `json:"market_id" yaml:"market_id"`
	BaseAsset  bool   `json:"base_asset" yaml:"base_asset"`
	QuoteAsset bool   `json:"quote_asset" yaml:"quote_asset"`
	Oracles    bool   `json:"oracles" yaml:"oracles"`
	Active     bool   `json:"active" yaml:"active"`
}

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

// DefaultNextProposalID is the starting poiint for proposal IDs.
const DefaultNextProposalID uint64 = 1

// GenesisState is state that must be provided at chain genesis.
type GenesisState struct {
	NextProposalID uint64      `json:"next_proposal_id" yaml:"next_proposal_id"`
	Committees     []Committee `json:"committees" yaml:"committees"`
	Proposals      []Proposal  `json:"proposals" yaml:"proposals"`
	Votes          []Vote      `json:"votes" yaml:"votes"`
}

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(nextProposalID uint64, committees []Committee, proposals []Proposal, votes []Vote) GenesisState {
	return GenesisState{
		NextProposalID: nextProposalID,
		Committees:     committees,
		Proposals:      proposals,
		Votes:          votes,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultNextProposalID,
		[]Committee{},
		[]Proposal{},
		[]Vote{},
	)
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	// validate committees
	committeeMap := make(map[uint64]bool, len(gs.Committees))
	for _, com := range gs.Committees {
		// check there are no duplicate IDs
		if _, ok := committeeMap[com.ID]; ok {
			return fmt.Errorf("duplicate committee ID found in genesis state; id: %d", com.ID)
		}
		committeeMap[com.ID] = true

		// validate committee
		if err := com.Validate(); err != nil {
			return err
		}
	}

	// validate proposals
	proposalMap := make(map[uint64]bool, len(gs.Proposals))
	for _, p := range gs.Proposals {
		// check there are no duplicate IDs
		if _, ok := proposalMap[p.ID]; ok {
			return fmt.Errorf("duplicate proposal ID found in genesis state; id: %d", p.ID)
		}
		proposalMap[p.ID] = true

		// validate next proposal ID
		if p.ID >= gs.NextProposalID {
			return fmt.Errorf("NextProposalID is not greater than all proposal IDs; id: %d", p.ID)
		}

		// check committee exists
		if !committeeMap[p.CommitteeID] {
			return fmt.Errorf("proposal refers to non existent committee; proposal: %+v", p)
		}

		// validate pubProposal
		if err := p.PubProposal.ValidateBasic(); err != nil {
			return fmt.Errorf("proposal %d invalid: %w", p.ID, err)
		}
	}

	// validate votes
	for _, v := range gs.Votes {
		// validate committee
		if err := v.Validate(); err != nil {
			return err
		}

		// check proposal exists
		if !proposalMap[v.ProposalID] {
			return fmt.Errorf("vote refers to non existent proposal; vote: %+v", v)
		}
	}
	return nil
}

func RegisterCodec(cdc *codec.Codec) {

	// Proposals
	cdc.RegisterInterface((*PubProposal)(nil), nil)

	// Permissions
	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "kava/GodPermission", nil)
	cdc.RegisterConcrete(SimpleParamChangePermission{}, "kava/SimpleParamChangePermission", nil)
	cdc.RegisterConcrete(TextPermission{}, "kava/TextPermission", nil)
	cdc.RegisterConcrete(SoftwareUpgradePermission{}, "kava/SoftwareUpgradePermission", nil)
	cdc.RegisterConcrete(SubParamChangePermission{}, "kava/SubParamChangePermission", nil)
}
