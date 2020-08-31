package v0_9

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CDP is the state of a single collateralized debt position.
type CDP struct {
	ID              uint64         `json:"id" yaml:"id"`                 // unique id for cdp
	Owner           sdk.AccAddress `json:"owner" yaml:"owner"`           // Account that authorizes changes to the CDP
	Collateral      sdk.Coin       `json:"collateral" yaml:"collateral"` // Amount of collateral stored in this CDP
	Principal       sdk.Coin       `json:"principal" yaml:"principal"`
	AccumulatedFees sdk.Coin       `json:"accumulated_fees" yaml:"accumulated_fees"`
	FeesUpdated     time.Time      `json:"fees_updated" yaml:"fees_updated"` // Amount of stable coin drawn from this CDP
}

// NewCDP creates a new CDP object
func NewCDP(id uint64, owner sdk.AccAddress, collateral sdk.Coin, principal sdk.Coin, time time.Time) CDP {
	fees := sdk.NewCoin(principal.Denom, sdk.ZeroInt())
	return CDP{
		ID:              id,
		Owner:           owner,
		Collateral:      collateral,
		Principal:       principal,
		AccumulatedFees: fees,
		FeesUpdated:     time,
	}
}

// CDPs a collection of CDP objects
type CDPs []CDP

// Deposit defines an amount of coins deposited by an account to a cdp
type Deposit struct {
	CdpID     uint64         `json:"cdp_id" yaml:"cdp_id"`       //  cdpID of the cdp
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"` //  Address of the depositor
	Amount    sdk.Coin       `json:"amount" yaml:"amount"`       //  Deposit amount
}

// NewDeposit creates a new Deposit object
func NewDeposit(cdpID uint64, depositor sdk.AccAddress, amount sdk.Coin) Deposit {
	return Deposit{cdpID, depositor, amount}
}

// Deposits a collection of Deposit objects
type Deposits []Deposit

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                   Params    `json:"params" yaml:"params"`
	CDPs                     CDPs      `json:"cdps" yaml:"cdps"`
	Deposits                 Deposits  `json:"deposits" yaml:"deposits"`
	StartingCdpID            uint64    `json:"starting_cdp_id" yaml:"starting_cdp_id"`
	DebtDenom                string    `json:"debt_denom" yaml:"debt_denom"`
	GovDenom                 string    `json:"gov_denom" yaml:"gov_denom"`
	PreviousDistributionTime time.Time `json:"previous_distribution_time" yaml:"previous_distribution_time"`
}

// Params governance parameters for cdp module
type Params struct {
	CollateralParams             CollateralParams `json:"collateral_params" yaml:"collateral_params"`
	DebtParam                    DebtParam        `json:"debt_param" yaml:"debt_param"`
	GlobalDebtLimit              sdk.Coin         `json:"global_debt_limit" yaml:"global_debt_limit"`
	SurplusAuctionThreshold      sdk.Int          `json:"surplus_auction_threshold" yaml:"surplus_auction_threshold"`
	SurplusAuctionLot            sdk.Int          `json:"surplus_auction_lot" yaml:"surplus_auction_lot"`
	DebtAuctionThreshold         sdk.Int          `json:"debt_auction_threshold" yaml:"debt_auction_threshold"`
	DebtAuctionLot               sdk.Int          `json:"debt_auction_lot" yaml:"debt_auction_lot"`
	SavingsDistributionFrequency time.Duration    `json:"savings_distribution_frequency" yaml:"savings_distribution_frequency"`
	CircuitBreaker               bool             `json:"circuit_breaker" yaml:"circuit_breaker"`
}

// NewParams returns a new params object
func NewParams(
	debtLimit sdk.Coin, collateralParams CollateralParams, debtParam DebtParam, surplusThreshold,
	surplusLot, debtThreshold, debtLot sdk.Int, distributionFreq time.Duration, breaker bool,
) Params {
	return Params{
		GlobalDebtLimit:              debtLimit,
		CollateralParams:             collateralParams,
		DebtParam:                    debtParam,
		SurplusAuctionThreshold:      surplusThreshold,
		SurplusAuctionLot:            surplusLot,
		DebtAuctionThreshold:         debtThreshold,
		DebtAuctionLot:               debtLot,
		SavingsDistributionFrequency: distributionFreq,
		CircuitBreaker:               breaker,
	}
}

// CollateralParam governance parameters for each collateral type within the cdp module
type CollateralParam struct {
	Denom               string   `json:"denom" yaml:"denom"`                             // Coin name of collateral type
	LiquidationRatio    sdk.Dec  `json:"liquidation_ratio" yaml:"liquidation_ratio"`     // The ratio (Collateral (priced in stable coin) / Debt) under which a CDP will be liquidated
	DebtLimit           sdk.Coin `json:"debt_limit" yaml:"debt_limit"`                   // Maximum amount of debt allowed to be drawn from this collateral type
	StabilityFee        sdk.Dec  `json:"stability_fee" yaml:"stability_fee"`             // per second stability fee for loans opened using this collateral
	AuctionSize         sdk.Int  `json:"auction_size" yaml:"auction_size"`               // Max amount of collateral to sell off in any one auction.
	LiquidationPenalty  sdk.Dec  `json:"liquidation_penalty" yaml:"liquidation_penalty"` // percentage penalty (between [0, 1]) applied to a cdp if it is liquidated
	Prefix              byte     `json:"prefix" yaml:"prefix"`
	SpotMarketID        string   `json:"spot_market_id" yaml:"spot_market_id"`               // marketID of the spot price of the asset from the pricefeed - used for opening CDPs, depositing, withdrawing
	LiquidationMarketID string   `json:"liquidation_market_id" yaml:"liquidation_market_id"` // marketID of the pricefeed used for liquidation
	ConversionFactor    sdk.Int  `json:"conversion_factor" yaml:"conversion_factor"`         // factor for converting internal units to one base unit of collateral
}

// CollateralParams array of CollateralParam
type CollateralParams []CollateralParam

// DebtParam governance params for debt assets
type DebtParam struct {
	Denom            string  `json:"denom" yaml:"denom"`
	ReferenceAsset   string  `json:"reference_asset" yaml:"reference_asset"`
	ConversionFactor sdk.Int `json:"conversion_factor" yaml:"conversion_factor"`
	DebtFloor        sdk.Int `json:"debt_floor" yaml:"debt_floor"`     // minimum active loan size, used to prevent dust
	SavingsRate      sdk.Dec `json:"savings_rate" yaml:"savings_rate"` // the percentage of stability fees that are redirected to savings rate
}
