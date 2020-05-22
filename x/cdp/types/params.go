package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"

	tmtime "github.com/tendermint/tendermint/types/time"
)

// Parameter keys
var (
	KeyGlobalDebtLimit       = []byte("GlobalDebtLimit")
	KeyCollateralParams      = []byte("CollateralParams")
	KeyDebtParam             = []byte("DebtParam")
	KeyDistributionFrequency = []byte("DistributionFrequency")
	KeyCircuitBreaker        = []byte("CircuitBreaker")
	KeyDebtThreshold         = []byte("DebtThreshold")
	KeySurplusThreshold      = []byte("SurplusThreshold")
	DefaultGlobalDebt        = sdk.NewCoin(DefaultStableDenom, sdk.ZeroInt())
	DefaultCircuitBreaker    = false
	DefaultCollateralParams  = CollateralParams{}
	DefaultDebtParam         = DebtParam{
		Denom:            "usdx",
		ReferenceAsset:   "usd",
		ConversionFactor: sdk.NewInt(6),
		DebtFloor:        sdk.NewInt(10000000),
		SavingsRate:      sdk.MustNewDecFromStr("0.95"),
	}
	DefaultCdpStartingID                = uint64(1)
	DefaultDebtDenom                    = "debt"
	DefaultGovDenom                     = "ukava"
	DefaultStableDenom                  = "usdx"
	DefaultSurplusThreshold             = sdk.NewInt(1000000000)
	DefaultDebtThreshold                = sdk.NewInt(1000000000)
	DefaultPreviousDistributionTime     = tmtime.Canonical(time.Unix(0, 0))
	DefaultSavingsDistributionFrequency = time.Hour * 12
	minCollateralPrefix                 = 0
	maxCollateralPrefix                 = 255
	stabilityFeeMax                     = sdk.MustNewDecFromStr("1.000000051034942716") // 500% APR
)

// Params governance parameters for cdp module
type Params struct {
	CollateralParams             CollateralParams `json:"collateral_params" yaml:"collateral_params"`
	DebtParam                    DebtParam        `json:"debt_param" yaml:"debt_param"`
	GlobalDebtLimit              sdk.Coin         `json:"global_debt_limit" yaml:"global_debt_limit"`
	SurplusAuctionThreshold      sdk.Int          `json:"surplus_auction_threshold" yaml:"surplus_auction_threshold"`
	DebtAuctionThreshold         sdk.Int          `json:"debt_auction_threshold" yaml:"debt_auction_threshold"`
	SavingsDistributionFrequency time.Duration    `json:"savings_distribution_frequency" yaml:"savings_distribution_frequency"`
	CircuitBreaker               bool             `json:"circuit_breaker" yaml:"circuit_breaker"`
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Global Debt Limit: %s
	Collateral Params: %s
	Debt Params: %s
	Surplus Auction Threshold: %s
	Debt Auction Threshold: %s
	Savings Distribution Frequency: %s
	Circuit Breaker: %t`,
		p.GlobalDebtLimit, p.CollateralParams, p.DebtParam, p.SurplusAuctionThreshold, p.DebtAuctionThreshold, p.SavingsDistributionFrequency, p.CircuitBreaker,
	)
}

// NewParams returns a new params object
func NewParams(debtLimit sdk.Coin, collateralParams CollateralParams, debtParam DebtParam, surplusThreshold sdk.Int, debtThreshold sdk.Int, distributionFreq time.Duration, breaker bool) Params {
	return Params{
		GlobalDebtLimit:              debtLimit,
		CollateralParams:             collateralParams,
		DebtParam:                    debtParam,
		DebtAuctionThreshold:         debtThreshold,
		SurplusAuctionThreshold:      surplusThreshold,
		SavingsDistributionFrequency: distributionFreq,
		CircuitBreaker:               breaker,
	}
}

// DefaultParams returns default params for cdp module
func DefaultParams() Params {
	return NewParams(DefaultGlobalDebt, DefaultCollateralParams, DefaultDebtParam, DefaultSurplusThreshold, DefaultDebtThreshold, DefaultSavingsDistributionFrequency, DefaultCircuitBreaker)
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
	SpotMarketID        string   `json:"spot_market_id" yaml:"spot_market_id"`              // marketID of the spot price of the asset from the pricefeed - used for opening CDPs, depositing, withdrawing
	LiquidationMarketID string   `json:"liquidation_market_id" yaml:"liquidation_market_id` // marketID of the pricefeed used for liquidation
	ConversionFactor    sdk.Int  `json:"conversion_factor" yaml:"conversion_factor"`        // factor for converting internal units to one base unit of collateral
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
	Spot Market ID: %s
	Liquidation Market ID: %s
	Conversion Factor: %s`,
		cp.Denom, cp.LiquidationRatio, cp.StabilityFee, cp.LiquidationPenalty, cp.DebtLimit, cp.AuctionSize, cp.Prefix, cp.SpotMarketID, cp.LiquidationMarketID, cp.ConversionFactor)
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
	DebtFloor        sdk.Int `json:"debt_floor" yaml:"debt_floor"`     // minimum active loan size, used to prevent dust
	SavingsRate      sdk.Dec `json:"savings_rate" yaml:"savings_rate"` // the percentage of stability fees that are redirected to savings rate
}

func (dp DebtParam) String() string {
	return fmt.Sprintf(`Debt:
	Denom: %s
	Reference Asset: %s
	Conversion Factor: %s
	Debt Floor %s
	Savings  Rate %s
	`, dp.Denom, dp.ReferenceAsset, dp.ConversionFactor, dp.DebtFloor, dp.SavingsRate)
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
		params.NewParamSetPair(KeyGlobalDebtLimit, &p.GlobalDebtLimit, validateGlobalDebtLimitParam),
		params.NewParamSetPair(KeyCollateralParams, &p.CollateralParams, validateCollateralParams),
		params.NewParamSetPair(KeyDebtParam, &p.DebtParam, validateDebtParam),
		params.NewParamSetPair(KeyCircuitBreaker, &p.CircuitBreaker, validateCircuitBreakerParam),
		params.NewParamSetPair(KeySurplusThreshold, &p.SurplusAuctionThreshold, validateSurplusAuctionThresholdParam),
		params.NewParamSetPair(KeyDebtThreshold, &p.DebtAuctionThreshold, validateDebtAuctionThresholdParam),
		params.NewParamSetPair(KeyDistributionFrequency, &p.SavingsDistributionFrequency, validateSavingsDistributionFrequencyParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateGlobalDebtLimitParam(p.GlobalDebtLimit); err != nil {
		return err
	}

	if err := validateCollateralParams(p.CollateralParams); err != nil {
		return err
	}

	if err := validateDebtParam(p.DebtParam); err != nil {
		return err
	}

	if err := validateCircuitBreakerParam(p.CircuitBreaker); err != nil {
		return err
	}

	if err := validateSurplusAuctionThresholdParam(p.SurplusAuctionThreshold); err != nil {
		return err
	}

	if err := validateDebtAuctionThresholdParam(p.DebtAuctionThreshold); err != nil {
		return err
	}

	if err := validateSavingsDistributionFrequencyParam(p.SavingsDistributionFrequency); err != nil {
		return err
	}

	if len(p.CollateralParams) == 0 { // default value OK
		return nil
	}

	if (DebtParam{}) != p.DebtParam {
		if p.DebtParam.Denom != p.GlobalDebtLimit.Denom {
			return fmt.Errorf("debt denom %s does not match global debt denom %s",
				p.DebtParam.Denom, p.GlobalDebtLimit.Denom)
		}
	}

	// validate collateral params
	collateralDupMap := make(map[string]int)
	prefixDupMap := make(map[int]int)
	collateralParamsDebtLimit := sdk.ZeroInt()

	for _, cp := range p.CollateralParams {

		prefix := int(cp.Prefix)
		prefixDupMap[prefix] = 1
		collateralDupMap[cp.Denom] = 1

		if cp.DebtLimit.Denom != p.GlobalDebtLimit.Denom {
			return fmt.Errorf("collateral debt limit denom %s does not match global debt limit denom %s",
				cp.DebtLimit.Denom, p.GlobalDebtLimit.Denom)
		}

		collateralParamsDebtLimit = collateralParamsDebtLimit.Add(cp.DebtLimit.Amount)

		if cp.DebtLimit.Amount.GT(p.GlobalDebtLimit.Amount) {
			return fmt.Errorf("collateral debt limit %s exceeds global debt limit: %s", cp.DebtLimit, p.GlobalDebtLimit)
		}
	}

	if collateralParamsDebtLimit.GT(p.GlobalDebtLimit.Amount) {
		return fmt.Errorf("sum of collateral debt limits %s exceeds global debt limit %s",
			collateralParamsDebtLimit, p.GlobalDebtLimit)
	}

	return nil
}

func validateGlobalDebtLimitParam(i interface{}) error {
	globalDebtLimit, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !globalDebtLimit.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "global debt limit %s", globalDebtLimit.String())
	}

	return nil
}

func validateCollateralParams(i interface{}) error {
	collateralParams, ok := i.(CollateralParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	collateralDupMap := make(map[string]bool)
	prefixDupMap := make(map[int]bool)
	for _, cp := range collateralParams {
		if err := sdk.ValidateDenom(cp.Denom); err != nil {
			return fmt.Errorf("collateral denom invalid %s", cp.Denom)
		}

		if strings.TrimSpace(cp.SpotMarketID) == "" {
			return fmt.Errorf("spot market id cannot be blank %s", cp)
		}

		if strings.TrimSpace(cp.LiquidationMarketID) == "" {
			return fmt.Errorf("liquidation market id cannot be blank %s", cp)
		}

		prefix := int(cp.Prefix)
		if prefix < minCollateralPrefix || prefix > maxCollateralPrefix {
			return fmt.Errorf("invalid prefix for collateral denom %s: %b", cp.Denom, cp.Prefix)
		}

		_, found := prefixDupMap[prefix]
		if found {
			return fmt.Errorf("duplicate prefix for collateral denom %s: %v", cp.Denom, []byte{cp.Prefix})
		}

		prefixDupMap[prefix] = true

		_, found = collateralDupMap[cp.Denom]
		if found {
			return fmt.Errorf("duplicate collateral denom: %s", cp.Denom)
		}

		collateralDupMap[cp.Denom] = true

		if !cp.DebtLimit.IsValid() {
			return fmt.Errorf("debt limit for all collaterals should be positive, is %s for %s", cp.DebtLimit, cp.Denom)
		}

		if cp.LiquidationPenalty.LT(sdk.ZeroDec()) || cp.LiquidationPenalty.GT(sdk.OneDec()) {
			return fmt.Errorf("liquidation penalty should be between 0 and 1, is %s for %s", cp.LiquidationPenalty, cp.Denom)
		}
		if !cp.AuctionSize.IsPositive() {
			return fmt.Errorf("auction size should be positive, is %s for %s", cp.AuctionSize, cp.Denom)
		}
		if cp.StabilityFee.LT(sdk.OneDec()) || cp.StabilityFee.GT(stabilityFeeMax) {
			return fmt.Errorf("stability fee must be ≥ 1.0, ≤ %s, is %s for %s", stabilityFeeMax, cp.StabilityFee, cp.Denom)
		}
	}

	return nil
}

func validateDebtParam(i interface{}) error {
	debtParam, ok := i.(DebtParam)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := sdk.ValidateDenom(debtParam.Denom); err != nil {
		return fmt.Errorf("debt denom invalid %s", debtParam.Denom)
	}

	if debtParam.SavingsRate.LT(sdk.ZeroDec()) || debtParam.SavingsRate.GT(sdk.OneDec()) {
		return fmt.Errorf("savings rate should be between 0 and 1, is %s for %s", debtParam.SavingsRate, debtParam.Denom)
	}
	return nil
}

func validateCircuitBreakerParam(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateSurplusAuctionThresholdParam(i interface{}) error {
	sat, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !sat.IsPositive() {
		return fmt.Errorf("surplus auction threshold should be positive: %s", sat)
	}

	return nil
}

func validateDebtAuctionThresholdParam(i interface{}) error {
	dat, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !dat.IsPositive() {
		return fmt.Errorf("debt auction threshold should be positive: %s", dat)
	}

	return nil
}

func validateSavingsDistributionFrequencyParam(i interface{}) error {
	sdf, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if sdf.Seconds() <= float64(0) {
		return fmt.Errorf("savings distribution frequency should be positive: %s", sdf)
	}

	return nil
}
