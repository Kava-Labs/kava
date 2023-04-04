package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var (
	KeyMoneyMarkets              = []byte("MoneyMarkets")
	KeyMinimumBorrowUSDValue     = []byte("MinimumBorrowUSDValue")
	DefaultMoneyMarkets          = MoneyMarkets{}
	DefaultMinimumBorrowUSDValue = sdk.NewDec(10) // $10 USD minimum borrow value
	DefaultAccumulationTimes     = GenesisAccumulationTimes{}
	DefaultTotalSupplied         = sdk.Coins{}
	DefaultTotalBorrowed         = sdk.Coins{}
	DefaultTotalReserves         = sdk.Coins{}
	DefaultDeposits              = Deposits{}
	DefaultBorrows               = Borrows{}
)

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
	if bl.LoanToValue.IsNegative() {
		return fmt.Errorf("loan-to-value must be a non-negative decimal: %s", bl.LoanToValue)
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

// NewMoneyMarket returns a new MoneyMarket
func NewMoneyMarket(denom string, borrowLimit BorrowLimit, spotMarketID string, conversionFactor sdkmath.Int,
	interestRateModel InterestRateModel, reserveFactor, keeperRewardPercentage sdk.Dec,
) MoneyMarket {
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

	if mm.ConversionFactor.IsNil() || mm.ConversionFactor.LT(sdk.OneInt()) {
		return fmt.Errorf("conversion '%s' factor must be ≥ one", mm.ConversionFactor)
	}

	if err := mm.InterestRateModel.Validate(); err != nil {
		return err
	}

	if mm.ReserveFactor.IsNegative() || mm.ReserveFactor.GT(sdk.OneDec()) {
		return fmt.Errorf("reserve factor must be between 0.0-1.0")
	}

	if mm.KeeperRewardPercentage.IsNegative() || mm.KeeperRewardPercentage.GT(sdk.OneDec()) {
		return fmt.Errorf("keeper reward percentage must be between 0.0-1.0")
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
		return fmt.Errorf("base rate APY must be in the inclusive range 0.0-1.0")
	}

	if irm.BaseMultiplier.IsNegative() {
		return fmt.Errorf("base multiplier must not be negative")
	}

	if irm.Kink.IsNegative() || irm.Kink.GT(sdk.OneDec()) {
		return fmt.Errorf("kink must be in the inclusive range 0.0-1.0")
	}

	if irm.JumpMultiplier.IsNegative() {
		return fmt.Errorf("jump multiplier must not be negative")
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

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMoneyMarkets, &p.MoneyMarkets, validateMoneyMarketParams),
		paramtypes.NewParamSetPair(KeyMinimumBorrowUSDValue, &p.MinimumBorrowUSDValue, validateMinimumBorrowUSDValue),
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
		return fmt.Errorf("minimum borrow USD value cannot be negative")
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
