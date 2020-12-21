package v0_13

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tmtime "github.com/tendermint/tendermint/types/time"
)

// Parameter keys
var (
	KeyGlobalDebtLimit        = []byte("GlobalDebtLimit")
	KeyCollateralParams       = []byte("CollateralParams")
	KeyDebtParam              = []byte("DebtParam")
	KeyDistributionFrequency  = []byte("DistributionFrequency")
	KeyCircuitBreaker         = []byte("CircuitBreaker")
	KeyDebtThreshold          = []byte("DebtThreshold")
	KeyDebtLot                = []byte("DebtLot")
	KeySurplusThreshold       = []byte("SurplusThreshold")
	KeySurplusLot             = []byte("SurplusLot")
	KeySavingsRateDistributed = []byte("SavingsRateDistributed")
	DefaultGlobalDebt         = sdk.NewCoin(DefaultStableDenom, sdk.ZeroInt())
	DefaultCircuitBreaker     = false
	DefaultCollateralParams   = CollateralParams{}
	DefaultDebtParam          = DebtParam{
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
	DefaultSurplusThreshold             = sdk.NewInt(500000000000)
	DefaultDebtThreshold                = sdk.NewInt(100000000000)
	DefaultSurplusLot                   = sdk.NewInt(10000000000)
	DefaultDebtLot                      = sdk.NewInt(10000000000)
	DefaultPreviousDistributionTime     = tmtime.Canonical(time.Unix(0, 0))
	DefaultSavingsDistributionFrequency = time.Hour * 12
	DefaultSavingsRateDistributed       = sdk.NewInt(0)
	minCollateralPrefix                 = 0
	maxCollateralPrefix                 = 255
	stabilityFeeMax                     = sdk.MustNewDecFromStr("1.000000051034942716") // 500% APR
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                    Params                   `json:"params" yaml:"params"`
	CDPs                      CDPs                     `json:"cdps" yaml:"cdps"`
	Deposits                  Deposits                 `json:"deposits" yaml:"deposits"`
	StartingCdpID             uint64                   `json:"starting_cdp_id" yaml:"starting_cdp_id"`
	DebtDenom                 string                   `json:"debt_denom" yaml:"debt_denom"`
	GovDenom                  string                   `json:"gov_denom" yaml:"gov_denom"`
	PreviousDistributionTime  time.Time                `json:"previous_distribution_time" yaml:"previous_distribution_time"`
	SavingsRateDistributed    sdk.Int                  `json:"savings_rate_distributed" yaml:"savings_rate_distributed"`
	PreviousAccumulationTimes GenesisAccumulationTimes `json:"previous_accumulation_times" yaml:"previous_accumulation_times"`
	TotalPrincipals           GenesisTotalPrincipals   `json:"total_principals" yaml:"total_principals"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, cdps CDPs, deposits Deposits, startingCdpID uint64,
	debtDenom, govDenom string, previousDistTime time.Time, savingsRateDist sdk.Int,
	prevAccumTimes GenesisAccumulationTimes, totalPrincipals GenesisTotalPrincipals) GenesisState {
	return GenesisState{
		Params:                    params,
		CDPs:                      cdps,
		Deposits:                  deposits,
		StartingCdpID:             startingCdpID,
		DebtDenom:                 debtDenom,
		GovDenom:                  govDenom,
		PreviousDistributionTime:  previousDistTime,
		SavingsRateDistributed:    savingsRateDist,
		PreviousAccumulationTimes: prevAccumTimes,
		TotalPrincipals:           totalPrincipals,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if err := gs.CDPs.Validate(); err != nil {
		return err
	}

	if err := gs.Deposits.Validate(); err != nil {
		return err
	}

	if err := gs.PreviousAccumulationTimes.Validate(); err != nil {
		return err
	}

	if err := gs.TotalPrincipals.Validate(); err != nil {
		return err
	}

	if gs.PreviousDistributionTime.IsZero() {
		return fmt.Errorf("previous distribution time not set")
	}

	if err := validateSavingsRateDistributed(gs.SavingsRateDistributed); err != nil {
		return err
	}

	if err := sdk.ValidateDenom(gs.DebtDenom); err != nil {
		return fmt.Errorf(fmt.Sprintf("debt denom invalid: %v", err))
	}

	if err := sdk.ValidateDenom(gs.GovDenom); err != nil {
		return fmt.Errorf(fmt.Sprintf("gov denom invalid: %v", err))
	}

	return nil
}

func validateSavingsRateDistributed(i interface{}) error {
	savingsRateDist, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if savingsRateDist.IsNegative() {
		return fmt.Errorf("savings rate distributed should not be negative: %s", savingsRateDist)
	}

	return nil
}

// GenesisAccumulationTime stores the previous distribution time and its corresponding denom
type GenesisAccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
	InterestFactor           sdk.Dec   `json:"interest_factor" yaml:"interest_factor"`
}

// NewGenesisAccumulationTime returns a new GenesisAccumulationTime
func NewGenesisAccumulationTime(ctype string, prevTime time.Time, factor sdk.Dec) GenesisAccumulationTime {
	return GenesisAccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
		InterestFactor:           factor,
	}
}

// GenesisAccumulationTimes slice of GenesisAccumulationTime
type GenesisAccumulationTimes []GenesisAccumulationTime

// Validate performs validation of GenesisAccumulationTimes
func (gats GenesisAccumulationTimes) Validate() error {
	for _, gat := range gats {
		if err := gat.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate performs validation of GenesisAccumulationTime
func (gat GenesisAccumulationTime) Validate() error {
	if gat.InterestFactor.LT(sdk.OneDec()) {
		return fmt.Errorf("interest factor should be ≥ 1.0, is %s for %s", gat.InterestFactor, gat.CollateralType)
	}
	return nil
}

// GenesisTotalPrincipal stores the total principal and its corresponding collateral type
type GenesisTotalPrincipal struct {
	CollateralType string  `json:"collateral_type" yaml:"collateral_type"`
	TotalPrincipal sdk.Int `json:"total_principal" yaml:"total_principal"`
}

// NewGenesisTotalPrincipal returns a new GenesisTotalPrincipal
func NewGenesisTotalPrincipal(ctype string, principal sdk.Int) GenesisTotalPrincipal {
	return GenesisTotalPrincipal{
		CollateralType: ctype,
		TotalPrincipal: principal,
	}
}

// GenesisTotalPrincipals slice of GenesisTotalPrincipal
type GenesisTotalPrincipals []GenesisTotalPrincipal

// Validate performs validation of GenesisTotalPrincipal
func (gtp GenesisTotalPrincipal) Validate() error {
	if gtp.TotalPrincipal.IsNegative() {
		return fmt.Errorf("total principal should be positive, is %s for %s", gtp.TotalPrincipal, gtp.CollateralType)
	}
	return nil
}

// Validate performs validation of GenesisTotalPrincipals
func (gtps GenesisTotalPrincipals) Validate() error {
	for _, gtp := range gtps {
		if err := gtp.Validate(); err != nil {
			return err
		}
	}
	return nil
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
	Denom               string   `json:"denom" yaml:"denom"` // Coin name of collateral type
	Type                string   `json:"type" yaml:"type"`
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

// NewCollateralParam returns a new CollateralParam
func NewCollateralParam(denom, ctype string, liqRatio sdk.Dec, debtLimit sdk.Coin, stabilityFee sdk.Dec, auctionSize sdk.Int, liqPenalty sdk.Dec, prefix byte, spotMarketID, liquidationMarketID string, conversionFactor sdk.Int) CollateralParam {
	return CollateralParam{
		Denom:               denom,
		Type:                ctype,
		LiquidationRatio:    liqRatio,
		DebtLimit:           debtLimit,
		StabilityFee:        stabilityFee,
		AuctionSize:         auctionSize,
		LiquidationPenalty:  liqPenalty,
		Prefix:              prefix,
		SpotMarketID:        spotMarketID,
		LiquidationMarketID: liquidationMarketID,
		ConversionFactor:    conversionFactor,
	}
}

// String implements fmt.Stringer
func (cp CollateralParam) String() string {
	return fmt.Sprintf(`Collateral:
	Denom: %s
	Type: %s
	Liquidation Ratio: %s
	Stability Fee: %s
	Liquidation Penalty: %s
	Debt Limit: %s
	Auction Size: %s
	Prefix: %b
	Spot Market ID: %s
	Liquidation Market ID: %s
	Conversion Factor: %s`,
		cp.Denom, cp.Type, cp.LiquidationRatio, cp.StabilityFee, cp.LiquidationPenalty, cp.DebtLimit, cp.AuctionSize, cp.Prefix, cp.SpotMarketID, cp.LiquidationMarketID, cp.ConversionFactor)
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

// NewDebtParam returns a new DebtParam
func NewDebtParam(denom, refAsset string, conversionFactor, debtFloor sdk.Int, savingsRate sdk.Dec) DebtParam {
	return DebtParam{
		Denom:            denom,
		ReferenceAsset:   refAsset,
		ConversionFactor: conversionFactor,
		DebtFloor:        debtFloor,
		SavingsRate:      savingsRate,
	}
}

// DebtParams array of DebtParam
type DebtParams []DebtParam

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

	if err := validateSurplusAuctionLotParam(p.SurplusAuctionLot); err != nil {
		return err
	}

	if err := validateDebtAuctionThresholdParam(p.DebtAuctionThreshold); err != nil {
		return err
	}

	if err := validateDebtAuctionLotParam(p.DebtAuctionLot); err != nil {
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

	prefixDupMap := make(map[int]bool)
	typeDupMap := make(map[string]bool)
	for _, cp := range collateralParams {
		if err := sdk.ValidateDenom(cp.Denom); err != nil {
			return fmt.Errorf("collateral denom invalid %s", cp.Denom)
		}

		if strings.TrimSpace(cp.SpotMarketID) == "" {
			return fmt.Errorf("spot market id cannot be blank %s", cp)
		}

		if strings.TrimSpace(cp.Type) == "" {
			return fmt.Errorf("collateral type cannot be blank %s", cp)
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

		_, found = typeDupMap[cp.Type]
		if found {
			return fmt.Errorf("duplicate cdp collateral type: %s", cp.Type)
		}
		typeDupMap[cp.Type] = true

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

func validateSurplusAuctionLotParam(i interface{}) error {
	sal, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !sal.IsPositive() {
		return fmt.Errorf("surplus auction lot should be positive: %s", sal)
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

func validateDebtAuctionLotParam(i interface{}) error {
	dal, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !dal.IsPositive() {
		return fmt.Errorf("debt auction lot should be positive: %s", dal)
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

// NewCDP creates a new CDP object
func NewCDP(id uint64, owner sdk.AccAddress, collateral sdk.Coin, collateralType string, principal sdk.Coin, time time.Time, interestFactor sdk.Dec) CDP {
	fees := sdk.NewCoin(principal.Denom, sdk.ZeroInt())
	return CDP{
		ID:              id,
		Owner:           owner,
		Type:            collateralType,
		Collateral:      collateral,
		Principal:       principal,
		AccumulatedFees: fees,
		FeesUpdated:     time,
		InterestFactor:  interestFactor,
	}
}

// NewCDPWithFees creates a new CDP object, for use during migration
func NewCDPWithFees(id uint64, owner sdk.AccAddress, collateral sdk.Coin, collateralType string, principal, fees sdk.Coin, time time.Time, interestFactor sdk.Dec) CDP {
	return CDP{
		ID:              id,
		Owner:           owner,
		Type:            collateralType,
		Collateral:      collateral,
		Principal:       principal,
		AccumulatedFees: fees,
		FeesUpdated:     time,
		InterestFactor:  interestFactor,
	}
}

// Validate performs a basic validation of the CDP fields.
func (cdp CDP) Validate() error {
	if cdp.ID == 0 {
		return errors.New("cdp id cannot be 0")
	}
	if cdp.Owner.Empty() {
		return errors.New("cdp owner cannot be empty")
	}
	if !cdp.Collateral.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "collateral %s", cdp.Collateral)
	}
	if !cdp.Principal.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "principal %s", cdp.Principal)
	}
	if !cdp.AccumulatedFees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "accumulated fees %s", cdp.AccumulatedFees)
	}
	if cdp.FeesUpdated.IsZero() {
		return errors.New("cdp updated fee time cannot be zero")
	}
	if strings.TrimSpace(cdp.Type) == "" {
		return fmt.Errorf("cdp type cannot be empty")
	}
	return nil
}

// GetTotalPrincipal returns the total principle for the cdp
func (cdp CDP) GetTotalPrincipal() sdk.Coin {
	return cdp.Principal.Add(cdp.AccumulatedFees)
}

// CDPs a collection of CDP objects
type CDPs []CDP

// Validate validates each CDP
func (cdps CDPs) Validate() error {
	for _, cdp := range cdps {
		if err := cdp.Validate(); err != nil {
			return err
		}
	}
	return nil
}

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

// Validate performs a basic validation of the deposit fields.
func (d Deposit) Validate() error {
	if d.CdpID == 0 {
		return errors.New("deposit's cdp id cannot be 0")
	}
	if d.Depositor.Empty() {
		return errors.New("depositor cannot be empty")
	}
	if !d.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "deposit %s", d.Amount)
	}
	return nil
}

// Deposits a collection of Deposit objects
type Deposits []Deposit

// Validate validates each deposit
func (ds Deposits) Validate() error {
	for _, d := range ds {
		if err := d.Validate(); err != nil {
			return err
		}
	}
	return nil
}
