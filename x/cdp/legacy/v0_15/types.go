package v0_15

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "cdp"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                    Params                   `json:"params" yaml:"params"`
	CDPs                      CDPs                     `json:"cdps" yaml:"cdps"`
	Deposits                  Deposits                 `json:"deposits" yaml:"deposits"`
	StartingCdpID             uint64                   `json:"starting_cdp_id" yaml:"starting_cdp_id"`
	DebtDenom                 string                   `json:"debt_denom" yaml:"debt_denom"`
	GovDenom                  string                   `json:"gov_denom" yaml:"gov_denom"`
	PreviousAccumulationTimes GenesisAccumulationTimes `json:"previous_accumulation_times" yaml:"previous_accumulation_times"`
	TotalPrincipals           GenesisTotalPrincipals   `json:"total_principals" yaml:"total_principals"`
}

// Params governance parameters for cdp module
type Params struct {
	CollateralParams        CollateralParams `json:"collateral_params" yaml:"collateral_params"`
	DebtParam               DebtParam        `json:"debt_param" yaml:"debt_param"`
	GlobalDebtLimit         sdk.Coin         `json:"global_debt_limit" yaml:"global_debt_limit"`
	SurplusAuctionThreshold sdkmath.Int      `json:"surplus_auction_threshold" yaml:"surplus_auction_threshold"`
	SurplusAuctionLot       sdkmath.Int      `json:"surplus_auction_lot" yaml:"surplus_auction_lot"`
	DebtAuctionThreshold    sdkmath.Int      `json:"debt_auction_threshold" yaml:"debt_auction_threshold"`
	DebtAuctionLot          sdkmath.Int      `json:"debt_auction_lot" yaml:"debt_auction_lot"`
	CircuitBreaker          bool             `json:"circuit_breaker" yaml:"circuit_breaker"`
}

// CollateralParams array of CollateralParam
type CollateralParams []CollateralParam

// CollateralParam governance parameters for each collateral type within the cdp module
type CollateralParam struct {
	Denom                            string      `json:"denom" yaml:"denom"` // Coin name of collateral type
	Type                             string      `json:"type" yaml:"type"`
	LiquidationRatio                 sdk.Dec     `json:"liquidation_ratio" yaml:"liquidation_ratio"`     // The ratio (Collateral (priced in stable coin) / Debt) under which a CDP will be liquidated
	DebtLimit                        sdk.Coin    `json:"debt_limit" yaml:"debt_limit"`                   // Maximum amount of debt allowed to be drawn from this collateral type
	StabilityFee                     sdk.Dec     `json:"stability_fee" yaml:"stability_fee"`             // per second stability fee for loans opened using this collateral
	AuctionSize                      sdkmath.Int `json:"auction_size" yaml:"auction_size"`               // Max amount of collateral to sell off in any one auction.
	LiquidationPenalty               sdk.Dec     `json:"liquidation_penalty" yaml:"liquidation_penalty"` // percentage penalty (between [0, 1]) applied to a cdp if it is liquidated
	Prefix                           byte        `json:"prefix" yaml:"prefix"`
	SpotMarketID                     string      `json:"spot_market_id" yaml:"spot_market_id"`                                           // marketID of the spot price of the asset from the pricefeed - used for opening CDPs, depositing, withdrawing
	LiquidationMarketID              string      `json:"liquidation_market_id" yaml:"liquidation_market_id"`                             // marketID of the pricefeed used for liquidation
	KeeperRewardPercentage           sdk.Dec     `json:"keeper_reward_percentage" yaml:"keeper_reward_percentage"`                       // the percentage of a CDPs collateral that gets rewarded to a keeper that liquidates the position
	CheckCollateralizationIndexCount sdkmath.Int `json:"check_collateralization_index_count" yaml:"check_collateralization_index_count"` // the number of cdps that will be checked for liquidation in the begin blocker
	ConversionFactor                 sdkmath.Int `json:"conversion_factor" yaml:"conversion_factor"`                                     // factor for converting internal units to one base unit of collateral
}

// CDPs a collection of CDP objects
type CDPs []CDP

// CDP is the state of a single collateralized debt position.
type CDP struct {
	ID              uint64         `json:"id" yaml:"id"`                             // unique id for cdp
	Owner           sdk.AccAddress `json:"owner" yaml:"owner"`                       // Account that authorizes changes to the CDP
	Type            string         `json:"type" yaml:"type"`                         // string representing the unique collateral type of the CDP
	Collateral      sdk.Coin       `json:"collateral" yaml:"collateral"`             // Amount of collateral stored in this CDP
	Principal       sdk.Coin       `json:"principal" yaml:"principal"`               // Amount of debt drawn using the CDP
	AccumulatedFees sdk.Coin       `json:"accumulated_fees" yaml:"accumulated_fees"` // Fees accumulated since the CDP was opened or debt was last repaid
	FeesUpdated     time.Time      `json:"fees_updated" yaml:"fees_updated"`         // The time when fees were last updated
	InterestFactor  sdk.Dec        `json:"interest_factor" yaml:"interest_factor"`   // the interest factor when fees were last calculated for this CDP
}

// Deposits a collection of Deposit objects
type Deposits []Deposit

// Deposit defines an amount of coins deposited by an account to a cdp
type Deposit struct {
	CdpID     uint64         `json:"cdp_id" yaml:"cdp_id"`       //  cdpID of the cdp
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"` //  Address of the depositor
	Amount    sdk.Coin       `json:"amount" yaml:"amount"`       //  Deposit amount
}

// GenesisAccumulationTimes slice of GenesisAccumulationTime
type GenesisAccumulationTimes []GenesisAccumulationTime

// GenesisAccumulationTime stores the previous distribution time and its corresponding denom
type GenesisAccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
	InterestFactor           sdk.Dec   `json:"interest_factor" yaml:"interest_factor"`
}

// GenesisTotalPrincipals slice of GenesisTotalPrincipal
type GenesisTotalPrincipals []GenesisTotalPrincipal

// GenesisTotalPrincipal stores the total principal and its corresponding collateral type
type GenesisTotalPrincipal struct {
	CollateralType string      `json:"collateral_type" yaml:"collateral_type"`
	TotalPrincipal sdkmath.Int `json:"total_principal" yaml:"total_principal"`
}

// DebtParam governance params for debt assets
type DebtParam struct {
	Denom            string      `json:"denom" yaml:"denom"`
	ReferenceAsset   string      `json:"reference_asset" yaml:"reference_asset"`
	ConversionFactor sdkmath.Int `json:"conversion_factor" yaml:"conversion_factor"`
	DebtFloor        sdkmath.Int `json:"debt_floor" yaml:"debt_floor"` // minimum active loan size, used to prevent dust
}
