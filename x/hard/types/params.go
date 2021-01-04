package types

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// Parameter keys and default values
var (
	KeyActive                 = []byte("Active")
	KeyLPSchedules            = []byte("LPSchedules")
	KeyDelegatorSchedule      = []byte("DelegatorSchedule")
	KeyMoneyMarkets           = []byte("MoneyMarkets")
	KeyCheckLtvIndexCount     = []byte("CheckLtvIndexCount")
	DefaultActive             = true
	DefaultGovSchedules       = DistributionSchedules{}
	DefaultLPSchedules        = DistributionSchedules{}
	DefaultDelegatorSchedules = DelegatorDistributionSchedules{}
	DefaultMoneyMarkets       = MoneyMarkets{}
	DefaultCheckLtvIndexCount = 10
	GovDenom                  = cdptypes.DefaultGovDenom
)

// Params governance parameters for hard module
type Params struct {
	Active                         bool                           `json:"active" yaml:"active"`
	LiquidityProviderSchedules     DistributionSchedules          `json:"liquidity_provider_schedules" yaml:"liquidity_provider_schedules"`
	DelegatorDistributionSchedules DelegatorDistributionSchedules `json:"delegator_distribution_schedules" yaml:"delegator_distribution_schedules"`
	MoneyMarkets                   MoneyMarkets                   `json:"money_markets" yaml:"money_markets"`
	CheckLtvIndexCount             int                            `json:"check_ltv_index_count" yaml:"check_ltv_index_count"`
}

// DistributionSchedule distribution schedule for liquidity providers
type DistributionSchedule struct {
	Active           bool        `json:"active" yaml:"active"`
	DepositDenom     string      `json:"deposit_denom" yaml:"deposit_denom"`
	Start            time.Time   `json:"start" yaml:"start"`
	End              time.Time   `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coin    `json:"rewards_per_second" yaml:"rewards_per_second"`
	ClaimEnd         time.Time   `json:"claim_end" yaml:"claim_end"`
	ClaimMultipliers Multipliers `json:"claim_multipliers" yaml:"claim_multipliers"`
}

// NewDistributionSchedule returns a new DistributionSchedule
func NewDistributionSchedule(active bool, denom string, start, end time.Time, reward sdk.Coin, claimEnd time.Time, multipliers Multipliers) DistributionSchedule {
	return DistributionSchedule{
		Active:           active,
		DepositDenom:     denom,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
		ClaimEnd:         claimEnd,
		ClaimMultipliers: multipliers,
	}
}

// String implements fmt.Stringer
func (ds DistributionSchedule) String() string {
	return fmt.Sprintf(`Liquidity Provider Distribution Schedule:
	Deposit Denom: %s,
	Start: %s,
	End: %s,
	Rewards Per Second: %s,
	Claim End: %s,
	Active: %t
	`, ds.DepositDenom, ds.Start, ds.End, ds.RewardsPerSecond, ds.ClaimEnd, ds.Active)
}

// Validate performs a basic check of a distribution schedule.
func (ds DistributionSchedule) Validate() error {
	if !ds.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward coins %s for %s", ds.RewardsPerSecond, ds.DepositDenom)
	}
	if !ds.RewardsPerSecond.IsPositive() {
		return fmt.Errorf("reward amount must be positive, is %s for %s", ds.RewardsPerSecond, ds.DepositDenom)
	}
	if ds.RewardsPerSecond.Denom != "hard" {
		return fmt.Errorf("reward denom should be hard, is %s", ds.RewardsPerSecond.Denom)
	}
	if ds.Start.IsZero() {
		return errors.New("reward period start time cannot be 0")
	}
	if ds.End.IsZero() {
		return errors.New("reward period end time cannot be 0")
	}
	if ds.Start.After(ds.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", ds.End, ds.Start)
	}
	if ds.ClaimEnd.Before(ds.End) {
		return fmt.Errorf("claim end time %s cannot be before end time %s", ds.ClaimEnd, ds.End)
	}
	for _, multiplier := range ds.ClaimMultipliers {
		if err := multiplier.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DistributionSchedules slice of DistributionSchedule
type DistributionSchedules []DistributionSchedule

// Validate checks if all the LiquidityProviderSchedules are valid and there are no duplicated
// entries.
func (dss DistributionSchedules) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, ds := range dss {
		if seenPeriods[ds.DepositDenom] {
			return fmt.Errorf("duplicated distribution provider schedule with deposit denom %s", ds.DepositDenom)
		}

		if err := ds.Validate(); err != nil {
			return err
		}
		seenPeriods[ds.DepositDenom] = true
	}

	return nil
}

// String implements fmt.Stringer
func (dss DistributionSchedules) String() string {
	out := "Distribution Schedules\n"
	for _, ds := range dss {
		out += fmt.Sprintf("%s\n", ds)
	}
	return out
}

// DelegatorDistributionSchedule distribution schedule for delegators
type DelegatorDistributionSchedule struct {
	DistributionSchedule DistributionSchedule `json:"distribution_schedule" yaml:"distribution_schedule"`

	DistributionFrequency time.Duration `json:"distribution_frequency" yaml:"distribution_frequency"`
}

// NewDelegatorDistributionSchedule returns a new DelegatorDistributionSchedule
func NewDelegatorDistributionSchedule(ds DistributionSchedule, frequency time.Duration) DelegatorDistributionSchedule {
	return DelegatorDistributionSchedule{
		DistributionSchedule:  ds,
		DistributionFrequency: frequency,
	}
}

// Validate performs a basic check of a reward fields.
func (dds DelegatorDistributionSchedule) Validate() error {
	if err := dds.DistributionSchedule.Validate(); err != nil {
		return err
	}
	if dds.DistributionFrequency <= 0 {
		return fmt.Errorf("distribution frequency should be positive, got %d", dds.DistributionFrequency)
	}
	return nil
}

// DelegatorDistributionSchedules slice of DelegatorDistributionSchedule
type DelegatorDistributionSchedules []DelegatorDistributionSchedule

// Validate checks if all the LiquidityProviderSchedules are valid and there are no duplicated
// entries.
func (dds DelegatorDistributionSchedules) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, ds := range dds {
		if seenPeriods[ds.DistributionSchedule.DepositDenom] {
			return fmt.Errorf("duplicated liquidity provider schedule with deposit denom %s", ds.DistributionSchedule.DepositDenom)
		}

		if err := ds.Validate(); err != nil {
			return err
		}
		seenPeriods[ds.DistributionSchedule.DepositDenom] = true
	}

	return nil
}

// Multiplier amount the claim rewards get increased by, along with how long the claim rewards are locked
type Multiplier struct {
	Name         MultiplierName `json:"name" yaml:"name"`
	MonthsLockup int64          `json:"months_lockup" yaml:"months_lockup"`
	Factor       sdk.Dec        `json:"factor" yaml:"factor"`
}

// NewMultiplier returns a new Multiplier
func NewMultiplier(name MultiplierName, lockup int64, factor sdk.Dec) Multiplier {
	return Multiplier{
		Name:         name,
		MonthsLockup: lockup,
		Factor:       factor,
	}
}

// Validate multiplier param
func (m Multiplier) Validate() error {
	if err := m.Name.IsValid(); err != nil {
		return err
	}
	if m.MonthsLockup < 0 {
		return fmt.Errorf("expected non-negative lockup, got %d", m.MonthsLockup)
	}
	if m.Factor.IsNegative() {
		return fmt.Errorf("expected non-negative factor, got %s", m.Factor.String())
	}

	return nil
}

// GetMultiplier returns the named multiplier from the input distribution schedule
func (ds DistributionSchedule) GetMultiplier(name MultiplierName) (Multiplier, bool) {
	for _, multiplier := range ds.ClaimMultipliers {
		if multiplier.Name == name {
			return multiplier, true
		}
	}
	return Multiplier{}, false
}

// Multipliers slice of Multiplier
type Multipliers []Multiplier

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
	AuctionSize            sdk.Int           `json:"auction_size" yaml:"auction_size"`
	KeeperRewardPercentage sdk.Dec           `json:"keeper_reward_percentage" yaml:"keeper_reward_percentages"`
}

// NewMoneyMarket returns a new MoneyMarket
func NewMoneyMarket(denom string, borrowLimit BorrowLimit, spotMarketID string, conversionFactor,
	auctionSize sdk.Int, interestRateModel InterestRateModel, reserveFactor, keeperRewardPercentage sdk.Dec) MoneyMarket {
	return MoneyMarket{
		Denom:                  denom,
		BorrowLimit:            borrowLimit,
		SpotMarketID:           spotMarketID,
		ConversionFactor:       conversionFactor,
		AuctionSize:            auctionSize,
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

	if !mm.AuctionSize.IsPositive() {
		return fmt.Errorf("Auction size must be a positive integer")
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
	if !mm.AuctionSize.Equal(mmCompareTo.AuctionSize) {
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
func NewParams(active bool, lps DistributionSchedules, dds DelegatorDistributionSchedules,
	moneyMarkets MoneyMarkets, checkLtvIndexCount int) Params {
	return Params{
		Active:                         active,
		LiquidityProviderSchedules:     lps,
		DelegatorDistributionSchedules: dds,
		MoneyMarkets:                   moneyMarkets,
		CheckLtvIndexCount:             checkLtvIndexCount,
	}
}

// DefaultParams returns default params for hard module
func DefaultParams() Params {
	return NewParams(DefaultActive, DefaultLPSchedules, DefaultDelegatorSchedules,
		DefaultMoneyMarkets, DefaultCheckLtvIndexCount)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Active: %t
	Liquidity Provider Distribution Schedules %s
	Delegator Distribution Schedule %s
	Money Markets %v
	Check LTV Index Count: %v`,
		p.Active, p.LiquidityProviderSchedules, p.DelegatorDistributionSchedules,
		p.MoneyMarkets, p.CheckLtvIndexCount)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyActive, &p.Active, validateActiveParam),
		params.NewParamSetPair(KeyLPSchedules, &p.LiquidityProviderSchedules, validateLPParams),
		params.NewParamSetPair(KeyDelegatorSchedule, &p.DelegatorDistributionSchedules, validateDelegatorParams),
		params.NewParamSetPair(KeyMoneyMarkets, &p.MoneyMarkets, validateMoneyMarketParams),
		params.NewParamSetPair(KeyCheckLtvIndexCount, &p.CheckLtvIndexCount, validateCheckLtvIndexCount),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateActiveParam(p.Active); err != nil {
		return err
	}

	if err := validateDelegatorParams(p.DelegatorDistributionSchedules); err != nil {
		return err
	}

	if err := validateLPParams(p.LiquidityProviderSchedules); err != nil {
		return err
	}

	if err := validateMoneyMarketParams(p.MoneyMarkets); err != nil {
		return err
	}

	return validateCheckLtvIndexCount(p.CheckLtvIndexCount)
}

func validateActiveParam(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateLPParams(i interface{}) error {
	dss, ok := i.(DistributionSchedules)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, ds := range dss {
		err := ds.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func validateDelegatorParams(i interface{}) error {
	dds, ok := i.(DelegatorDistributionSchedules)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return dds.Validate()
}

func validateMoneyMarketParams(i interface{}) error {
	mm, ok := i.(MoneyMarkets)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return mm.Validate()
}

func validateCheckLtvIndexCount(i interface{}) error {
	ltvCheckCount, ok := i.(int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if ltvCheckCount < 0 {
		return fmt.Errorf("CheckLtvIndexCount param must be positive, got: %d", ltvCheckCount)
	}

	return nil
}
