package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// Parameter keys
var (
	KeyGlobalDebtLimit       = []byte("GlobalDebtLimit")
	KeyCollateralParams      = []byte("CollateralParams")
	KeyDebtParams            = []byte("DebtParams")
	KeyCircuitBreaker        = []byte("CircuitBreaker")
	KeyDebtThreshold         = []byte("DebtThreshold")
	KeySurplusThreshold      = []byte("SurplusThreshold")
	DefaultGlobalDebt        = sdk.Coins{}
	DefaultCircuitBreaker    = false
	DefaultCollateralParams  = CollateralParams{}
	DefaultDebtParams        = DebtParams{}
	DefaultCdpStartingID     = uint64(1)
	DefaultDebtDenom         = "debt"
	DefaultGovDenom          = "ukava"
	DefaultSurplusThreshold  = sdk.NewInt(1000)
	DefaultDebtThreshold     = sdk.NewInt(1000)
	DefaultPreviousBlockTime = tmtime.Canonical(time.Unix(0, 0))
	minCollateralPrefix      = 0
	maxCollateralPrefix      = 255
)

// Params governance parameters for cdp module
type Params struct {
	CollateralParams        CollateralParams `json:"collateral_params" yaml:"collateral_params"`
	DebtParams              DebtParams       `json:"debt_params" yaml:"debt_params"`
	GlobalDebtLimit         sdk.Coins        `json:"global_debt_limit" yaml:"global_debt_limit"`
	SurplusAuctionThreshold sdk.Int          `json:"surplus_auction_threshold" yaml:"surplus_auction_threshold"`
	DebtAuctionThreshold    sdk.Int          `json:"debt_auction_threshold" yaml:"debt_auction_threshold"`
	CircuitBreaker          bool             `json:"circuit_breaker" yaml:"circuit_breaker"`
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Global Debt Limit: %s
	Collateral Params: %s
	Debt Params: %s
	Surplus Auction Threshold: %s
	Debt Auction Threshold: %s
	Circuit Breaker: %t`,
		p.GlobalDebtLimit, p.CollateralParams, p.DebtParams, p.SurplusAuctionThreshold, p.DebtAuctionThreshold, p.CircuitBreaker,
	)
}

// NewParams returns a new params object
func NewParams(debtLimit sdk.Coins, collateralParams CollateralParams, debtParams DebtParams, surplusThreshold sdk.Int, debtThreshold sdk.Int, breaker bool) Params {
	return Params{
		GlobalDebtLimit:         debtLimit,
		CollateralParams:        collateralParams,
		DebtParams:              debtParams,
		DebtAuctionThreshold:    debtThreshold,
		SurplusAuctionThreshold: surplusThreshold,
		CircuitBreaker:          breaker,
	}
}

// DefaultParams returns default params for cdp module
func DefaultParams() Params {
	return NewParams(DefaultGlobalDebt, DefaultCollateralParams, DefaultDebtParams, DefaultSurplusThreshold, DefaultDebtThreshold, DefaultCircuitBreaker)
}

// CollateralParam governance parameters for each collateral type within the cdp module
type CollateralParam struct {
	Denom              string    `json:"denom" yaml:"denom"`                             // Coin name of collateral type
	LiquidationRatio   sdk.Dec   `json:"liquidation_ratio" yaml:"liquidation_ratio"`     // The ratio (Collateral (priced in stable coin) / Debt) under which a CDP will be liquidated
	DebtLimit          sdk.Coins `json:"debt_limit" yaml:"debt_limit"`                   // Maximum amount of debt allowed to be drawn from this collateral type
	StabilityFee       sdk.Dec   `json:"stability_fee" yaml:"stability_fee"`             // per second stability fee for loans opened using this collateral
	AuctionSize        sdk.Int   `json:"auction_size" yaml:"auction_size"`               // Max amount of collateral to sell off in any one auction.
	LiquidationPenalty sdk.Dec   `json:"liquidation_penalty" yaml:"liquidation_penalty"` // percentage penalty (between [0, 1]) applied to a cdp if it is liquidated
	Prefix             byte      `json:"prefix" yaml:"prefix"`
	MarketID           string    `json:"market_id" yaml:"market_id"`                 // marketID for fetching price of the asset from the pricefeed
	ConversionFactor   sdk.Int   `json:"conversion_factor" yaml:"conversion_factor"` // factor for converting internal units to one base unit of collateral
}

// String implements fmt.Stringer
func (cp CollateralParam) String() string {
	return fmt.Sprintf(`Collateral:
	Denom: %s
	Liquidation Ratio: %s
	Stability Fee: %s
	Liquidation Penalty: %s
	Debt Limit: %s
	Auction Size: %s
	Prefix: %b
	Market ID: %s
	Conversion Factor: %s`,
		cp.Denom, cp.LiquidationRatio, cp.StabilityFee, cp.LiquidationPenalty, cp.DebtLimit, cp.AuctionSize, cp.Prefix, cp.MarketID, cp.ConversionFactor)
}

// CollateralParams array of CollateralParam
type CollateralParams []CollateralParam

// String implements fmt.Stringer
func (cps CollateralParams) String() string {
	out := "Collateral Params\n"
	for _, cp := range cps {
		out += fmt.Sprintf("%s\n", cp)
	}
	return out
}

// DebtParam governance params for debt assets
type DebtParam struct {
	Denom            string  `json:"denom" yaml:"denom"`
	ReferenceAsset   string  `json:"reference_asset" yaml:"reference_asset"`
	ConversionFactor sdk.Int `json:"conversion_factor" yaml:"conversion_factor"`
	DebtFloor        sdk.Int `json:"debt_floor" yaml:"debt_floor"` // minimum active loan size, used to prevent dust
}

func (dp DebtParam) String() string {
	return fmt.Sprintf(`Debt:
	Denom: %s
	Reference Asset: %s
	Conversion Factor: %s
	Debt Floor %s`, dp.Denom, dp.ReferenceAsset, dp.ConversionFactor, dp.DebtFloor)
}

// DebtParams array of DebtParam
type DebtParams []DebtParam

// String implements fmt.Stringer
func (dps DebtParams) String() string {
	out := "Debt Params\n"
	for _, dp := range dps {
		out += fmt.Sprintf("%s\n", dp)
	}
	return out
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyGlobalDebtLimit, Value: &p.GlobalDebtLimit},
		{Key: KeyCollateralParams, Value: &p.CollateralParams},
		{Key: KeyDebtParams, Value: &p.DebtParams},
		{Key: KeyCircuitBreaker, Value: &p.CircuitBreaker},
		{Key: KeySurplusThreshold, Value: &p.SurplusAuctionThreshold},
		{Key: KeyDebtThreshold, Value: &p.DebtAuctionThreshold},
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	// validate debt params
	debtDenoms := make(map[string]int)
	for _, dp := range p.DebtParams {
		_, found := debtDenoms[dp.Denom]
		if found {
			return fmt.Errorf("duplicate debt denom: %s", dp.Denom)
		}
		debtDenoms[dp.Denom] = 1

	}

	// validate collateral params
	collateralDupMap := make(map[string]int)
	prefixDupMap := make(map[int]int)
	collateralParamsDebtLimit := sdk.Coins{}
	for _, cp := range p.CollateralParams {
		prefix := int(cp.Prefix)
		if prefix < minCollateralPrefix || prefix > maxCollateralPrefix {
			return fmt.Errorf("invalid prefix for collateral denom %s: %b", cp.Denom, cp.Prefix)
		}
		_, found := prefixDupMap[prefix]
		if found {
			return fmt.Errorf("duplicate prefix for collateral denom %s: %v", cp.Denom, []byte{cp.Prefix})
		}

		prefixDupMap[prefix] = 1
		_, found = collateralDupMap[cp.Denom]

		if found {
			return fmt.Errorf("duplicate collateral denom: %s", cp.Denom)
		}
		collateralDupMap[cp.Denom] = 1

		if cp.DebtLimit.IsAnyNegative() {
			return fmt.Errorf("debt limit for all collaterals should be positive, is %s for %s", cp.DebtLimit, cp.Denom)
		}
		collateralParamsDebtLimit = collateralParamsDebtLimit.Add(cp.DebtLimit)

		for _, dc := range cp.DebtLimit {
			_, found := debtDenoms[dc.Denom]
			if !found {
				return fmt.Errorf("debt limit for collateral %s contains invalid debt denom %s", cp.Denom, dc.Denom)
			}
		}
		if cp.DebtLimit.IsAnyGT(p.GlobalDebtLimit) {
			return fmt.Errorf("collateral debt limit for %s exceeds global debt limit: \n\tglobal debt limit: %s\n\tcollateral debt limits: %s",
				cp.Denom, p.GlobalDebtLimit, cp.DebtLimit)
		}
		if cp.LiquidationPenalty.LT(sdk.ZeroDec()) || cp.LiquidationPenalty.GT(sdk.OneDec()) {
			return fmt.Errorf("liquidation penalty should be between 0 and 1, is %s for %s", cp.LiquidationPenalty, cp.Denom)
		}
		if !cp.AuctionSize.IsPositive() {
			return fmt.Errorf("auction size should be positive, is %s for %s", cp.AuctionSize, cp.Denom)
		}
		if cp.StabilityFee.LT(sdk.OneDec()) {
			return fmt.Errorf("stability fee must be ≥ 1.0, is %s for %s", cp.StabilityFee, cp.Denom)
		}
	}
	if collateralParamsDebtLimit.IsAnyGT(p.GlobalDebtLimit) {
		return fmt.Errorf("collateral debt limit exceeds global debt limit:\n\tglobal debt limit: %s\n\tcollateral debt limits: %s",
			p.GlobalDebtLimit, collateralParamsDebtLimit)
	}

	// validate global params
	if p.GlobalDebtLimit.IsAnyNegative() {
		return fmt.Errorf("global debt limit should be positive for all debt tokens, is %s", p.GlobalDebtLimit)
	}
	if !p.SurplusAuctionThreshold.IsPositive() {
		return fmt.Errorf("surplus auction threshold should be positive, is %s", p.SurplusAuctionThreshold)
	}
	if !p.DebtAuctionThreshold.IsPositive() {
		return fmt.Errorf("debt auction threshold should be positive, is %s", p.DebtAuctionThreshold)
	}
	return nil
}
