package v0_15

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface{}

// ------------------------------------------
//				GodPermission
// ------------------------------------------

// GodPermission allows any governance proposal. It is used mainly for testing.
type GodPermission struct{}

var _ Permission = GodPermission{}

// ------------------------------------------
//				SimpleParamChangePermission
// ------------------------------------------

// SimpleParamChangePermission only allows changes to certain params
type SimpleParamChangePermission struct {
	AllowedParams AllowedParams `json:"allowed_params" yaml:"allowed_params"`
}

var _ Permission = SimpleParamChangePermission{}

// AllowedParam permission type for module parameter keys
type AllowedParam struct {
	Subspace string `json:"subspace" yaml:"subspace"`
	Key      string `json:"key" yaml:"key"`
}

// AllowedParams slice of AllowedParam
type AllowedParams []AllowedParam

// ------------------------------------------
//				TextPermission
// ------------------------------------------

// TextPermission allows any text governance proposal.
type TextPermission struct{}

var _ Permission = TextPermission{}

// ------------------------------------------
//				SoftwareUpgradePermission
// ------------------------------------------

// SoftwareUpgradePermission permission type for software upgrade proposals
type SoftwareUpgradePermission struct{}

var _ Permission = SoftwareUpgradePermission{}

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

// AllowedCollateralParams slice of AllowedCollateralParam
type AllowedCollateralParams []AllowedCollateralParam

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
	prefix, spotMarket, liquidationMarket, conversionFactor, keeperReward, ltvIndexCount bool,
) AllowedCollateralParam {
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

// AllowedDebtParam permission struct for changes to debt parameter keys (cdp module)
type AllowedDebtParam struct {
	Denom            bool `json:"denom" yaml:"denom"`
	ReferenceAsset   bool `json:"reference_asset" yaml:"reference_asset"`
	ConversionFactor bool `json:"conversion_factor" yaml:"conversion_factor"`
	DebtFloor        bool `json:"debt_floor" yaml:"debt_floor"`
}

// AllowedAssetParams slice of AllowedAssetParam
type AllowedAssetParams []AllowedAssetParam

// AllowedAssetParam bep3 asset parameters that can be changed by committee
type AllowedAssetParam struct {
	Denom         string `json:"denom" yaml:"denom"`
	CoinID        bool   `json:"coin_id" yaml:"coin_id"`
	Limit         bool   `json:"limit" yaml:"limit"`
	Active        bool   `json:"active" yaml:"active"`
	MaxSwapAmount bool   `json:"max_swap_amount" yaml:"max_swap_amount"`
	MinBlockLock  bool   `json:"min_block_lock" yaml:"min_block_lock"`
}

// AllowedMarkets slice of AllowedMarket
type AllowedMarkets []AllowedMarket

// AllowedMarket permission struct for market parameters (pricefeed module)
type AllowedMarket struct {
	MarketID   string `json:"market_id" yaml:"market_id"`
	BaseAsset  bool   `json:"base_asset" yaml:"base_asset"`
	QuoteAsset bool   `json:"quote_asset" yaml:"quote_asset"`
	Oracles    bool   `json:"oracles" yaml:"oracles"`
	Active     bool   `json:"active" yaml:"active"`
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

// AllowedMoneyMarkets slice of AllowedMoneyMarket
type AllowedMoneyMarkets []AllowedMoneyMarket
