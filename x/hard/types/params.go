package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// Parameter keys and default values
var (
	KeyMoneyMarkets              = []byte("MoneyMarkets")
	KeyMinimumBorrowUSDValue     = []byte("MinimumBorrowUSDValue")
	DefaultMoneyMarkets          = MoneyMarkets{}
	DefaultMinimumBorrowUSDValue = sdk.NewDec(10) // $10 USD minimum borrow value
	GovDenom                     = cdptypes.DefaultGovDenom
	DefaultAccumulationTimes     = GenesisAccumulationTimes{}
	DefaultTotalSupplied         = sdk.Coins{}
	DefaultTotalBorrowed         = sdk.Coins{}
	DefaultTotalReserves         = sdk.Coins{}
	DefaultDeposits              = Deposits{}
	DefaultBorrows               = Borrows{}
)

// Params governance parameters for hard module
type Params struct {
	MoneyMarkets          MoneyMarkets `json:"money_markets" yaml:"money_markets"`
	MinimumBorrowUSDValue sdk.Dec      `json:"minimum_borrow_usd_value" yaml:"minimum_borrow_usd_value"`
}

// BorrowLimit enforces restrictions on a money market
type BorrowLimit struct {
	HasMaxLimit  bool    `json:"has_max_limit" yaml:"has_max_limit"`
	MaximumLimit sdk.Dec `json:"maximum_limit" yaml:"maximum_limit"`
	LoanToValue  sdk.Dec `json:"loan_to_value" yaml:"loan_to_value"`
}

// NewBorrowLimit returns a new BorrowLimit
func NewBorrowLimit(hasMaxLimit bool, maximumLimit, loanToValue sdk.Dec) BorrowLimit {
	return BorrowLimit{
		HasMaxLimit:  hasMaxLimit,
		MaximumLimit: maximumLimit,
		LoanToValue:  loanToValue,
	}
}

// Validate BorrowLimit
func (bl BorrowLimit) Validate() error {
	if bl.MaximumLimit.IsNegative() {
		return fmt.Errorf("maximum limit USD cannot be negative: %s", bl.MaximumLimit)
	}
	if !bl.LoanToValue.IsPositive() {
		return fmt.Errorf("loan-to-value must be a positive integer: %s", bl.LoanToValue)
	}
	if bl.LoanToValue.GT(sdk.OneDec()) {
		return fmt.Errorf("loan-to-value cannot be greater than 1.0: %s", bl.LoanToValue)
	}
	return nil
}

// Equal returns a boolean indicating if an BorrowLimit is equal to another BorrowLimit
func (bl BorrowLimit) Equal(blCompareTo BorrowLimit) bool {
	if bl.HasMaxLimit != blCompareTo.HasMaxLimit {
		return false
	}
	if !bl.MaximumLimit.Equal(blCompareTo.MaximumLimit) {
		return false
	}
	if !bl.LoanToValue.Equal(blCompareTo.LoanToValue) {
		return false
	}
	return true
}

// MoneyMarket is a money market for an individual asset
type MoneyMarket struct {
	Denom                  string            `json:"denom" yaml:"denom"`
	BorrowLimit            BorrowLimit       `json:"borrow_limit" yaml:"borrow_limit"`
	SpotMarketID           string            `json:"spot_market_id" yaml:"spot_market_id"`
	ConversionFactor       sdk.Int           `json:"conversion_factor" yaml:"conversion_factor"`
	InterestRateModel      InterestRateModel `json:"interest_rate_model" yaml:"interest_rate_model"`
	ReserveFactor          sdk.Dec           `json:"reserve_factor" yaml:"reserve_factor"`
	KeeperRewardPercentage sdk.Dec           `json:"keeper_reward_percentage" yaml:"keeper_reward_percentages"`
}

// NewMoneyMarket returns a new MoneyMarket
func NewMoneyMarket(denom string, borrowLimit BorrowLimit, spotMarketID string, conversionFactor sdk.Int,
	interestRateModel InterestRateModel, reserveFactor, keeperRewardPercentage sdk.Dec) MoneyMarket {
	return MoneyMarket{
		Denom:                  denom,
		BorrowLimit:            borrowLimit,
		SpotMarketID:           spotMarketID,
		ConversionFactor:       conversionFactor,
		InterestRateModel:      interestRateModel,
		ReserveFactor:          reserveFactor,
		KeeperRewardPercentage: keeperRewardPercentage,
	}
}

// Validate MoneyMarket param
func (mm MoneyMarket) Validate() error {
	if err := sdk.ValidateDenom(mm.Denom); err != nil {
		return err
	}

	if err := mm.BorrowLimit.Validate(); err != nil {
		return err
	}

	if err := mm.InterestRateModel.Validate(); err != nil {
		return err
	}

	if mm.ReserveFactor.IsNegative() || mm.ReserveFactor.GT(sdk.OneDec()) {
		return fmt.Errorf("Reserve factor must be between 0.0-1.0")
	}

	if mm.KeeperRewardPercentage.IsNegative() || mm.KeeperRewardPercentage.GT(sdk.OneDec()) {
		return fmt.Errorf("Keeper reward percentage must be between 0.0-1.0")
	}

	return nil
}

// Equal returns a boolean indicating if a MoneyMarket is equal to another MoneyMarket
func (mm MoneyMarket) Equal(mmCompareTo MoneyMarket) bool {
	if mm.Denom != mmCompareTo.Denom {
		return false
	}
	if !mm.BorrowLimit.Equal(mmCompareTo.BorrowLimit) {
		return false
	}
	if mm.SpotMarketID != mmCompareTo.SpotMarketID {
		return false
	}
	if !mm.ConversionFactor.Equal(mmCompareTo.ConversionFactor) {
		return false
	}
	if !mm.InterestRateModel.Equal(mmCompareTo.InterestRateModel) {
		return false
	}
	if !mm.ReserveFactor.Equal(mmCompareTo.ReserveFactor) {
		return false
	}
	if !mm.KeeperRewardPercentage.Equal(mmCompareTo.KeeperRewardPercentage) {
		return false
	}
	return true
}

// MoneyMarkets slice of MoneyMarket
type MoneyMarkets []MoneyMarket

// Validate borrow limits
func (mms MoneyMarkets) Validate() error {
	for _, moneyMarket := range mms {
		if err := moneyMarket.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// InterestRateModel contains information about an asset's interest rate
type InterestRateModel struct {
	BaseRateAPY    sdk.Dec `json:"base_rate_apy" yaml:"base_rate_apy"`
	BaseMultiplier sdk.Dec `json:"base_multiplier" yaml:"base_multiplier"`
	Kink           sdk.Dec `json:"kink" yaml:"kink"`
	JumpMultiplier sdk.Dec `json:"jump_multiplier" yaml:"jump_multiplier"`
}

// NewInterestRateModel returns a new InterestRateModel
func NewInterestRateModel(baseRateAPY, baseMultiplier, kink, jumpMultiplier sdk.Dec) InterestRateModel {
	return InterestRateModel{
		BaseRateAPY:    baseRateAPY,
		BaseMultiplier: baseMultiplier,
		Kink:           kink,
		JumpMultiplier: jumpMultiplier,
	}
}

// Validate InterestRateModel param
func (irm InterestRateModel) Validate() error {
	if irm.BaseRateAPY.IsNegative() || irm.BaseRateAPY.GT(sdk.OneDec()) {
		return fmt.Errorf("Base rate APY must be between 0.0-1.0")
	}

	if irm.BaseMultiplier.IsNegative() {
		return fmt.Errorf("Base multiplier must be positive")
	}

	if irm.Kink.IsNegative() || irm.Kink.GT(sdk.OneDec()) {
		return fmt.Errorf("Kink must be between 0.0-1.0")
	}

	if irm.JumpMultiplier.IsNegative() {
		return fmt.Errorf("Jump multiplier must be positive")
	}

	return nil
}

// Equal returns a boolean indicating if an InterestRateModel is equal to another InterestRateModel
func (irm InterestRateModel) Equal(irmCompareTo InterestRateModel) bool {
	if !irm.BaseRateAPY.Equal(irmCompareTo.BaseRateAPY) {
		return false
	}
	if !irm.BaseMultiplier.Equal(irmCompareTo.BaseMultiplier) {
		return false
	}
	if !irm.Kink.Equal(irmCompareTo.Kink) {
		return false
	}
	if !irm.JumpMultiplier.Equal(irmCompareTo.JumpMultiplier) {
		return false
	}
	return true
}

// InterestRateModels slice of InterestRateModel
type InterestRateModels []InterestRateModel

// NewParams returns a new params object
func NewParams(moneyMarkets MoneyMarkets, minimumBorrowUSDValue sdk.Dec) Params {
	return Params{
		MoneyMarkets:          moneyMarkets,
		MinimumBorrowUSDValue: minimumBorrowUSDValue,
	}
}

// DefaultParams returns default params for hard module
func DefaultParams() Params {
	return NewParams(DefaultMoneyMarkets, DefaultMinimumBorrowUSDValue)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Minimum Borrow USD Value: %v
	Money Markets: %v`,
		p.MinimumBorrowUSDValue, p.MoneyMarkets)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMoneyMarkets, &p.MoneyMarkets, validateMoneyMarketParams),
		params.NewParamSetPair(KeyMinimumBorrowUSDValue, &p.MinimumBorrowUSDValue, validateMinimumBorrowUSDValue),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateMinimumBorrowUSDValue(p.MinimumBorrowUSDValue); err != nil {
		return err
	}

	return validateMoneyMarketParams(p.MoneyMarkets)
}

func validateMinimumBorrowUSDValue(i interface{}) error {
	minBorrowVal, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if minBorrowVal.IsNegative() {
		return fmt.Errorf("Minimum borrow USD value cannot be negative")
	}

	return nil
}

func validateMoneyMarketParams(i interface{}) error {
	mm, ok := i.(MoneyMarkets)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return mm.Validate()
}
