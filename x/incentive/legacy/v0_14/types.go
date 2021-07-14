package v0_14

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	tmtime "github.com/tendermint/tendermint/types/time"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	kavadistTypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Valid reward multipliers
const (
	Small                          MultiplierName = "small"
	Medium                         MultiplierName = "medium"
	Large                          MultiplierName = "large"
	USDXMintingClaimType                          = "usdx_minting"
	HardLiquidityProviderClaimType                = "hard_liquidity_provider"
	BondDenom                                     = "ukava"
	ModuleName                                    = "incentive"
)

// Parameter keys and default values
var (
	KeyUSDXMintingRewardPeriods     = []byte("USDXMintingRewardPeriods")
	KeyHardSupplyRewardPeriods      = []byte("HardSupplyRewardPeriods")
	KeyHardBorrowRewardPeriods      = []byte("HardBorrowRewardPeriods")
	KeyHardDelegatorRewardPeriods   = []byte("HardDelegatorRewardPeriods")
	KeyClaimEnd                     = []byte("ClaimEnd")
	KeyMultipliers                  = []byte("ClaimMultipliers")
	DefaultActive                   = false
	DefaultRewardPeriods            = RewardPeriods{}
	DefaultMultiRewardPeriods       = MultiRewardPeriods{}
	DefaultMultipliers              = Multipliers{}
	DefaultUSDXClaims               = USDXMintingClaims{}
	DefaultHardClaims               = HardLiquidityProviderClaims{}
	DefaultGenesisAccumulationTimes = GenesisAccumulationTimes{}
	DefaultGenesisRewardIndexes     = GenesisRewardIndexesSlice{}
	DefaultClaimEnd                 = tmtime.Canonical(time.Unix(1, 0))
	GovDenom                        = cdptypes.DefaultGovDenom
	PrincipalDenom                  = "usdx"
	IncentiveMacc                   = kavadistTypes.ModuleName
)

// RegisterCodec registers the necessary types for incentive module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Claim)(nil), nil)
	cdc.RegisterConcrete(USDXMintingClaim{}, "incentive/USDXMintingClaim", nil)
	cdc.RegisterConcrete(HardLiquidityProviderClaim{}, "incentive/HardLiquidityProviderClaim", nil)
}

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                         Params                      `json:"params" yaml:"params"`
	USDXAccumulationTimes          GenesisAccumulationTimes    `json:"usdx_accumulation_times" yaml:"usdx_accumulation_times"`
	USDXRewardIndexes              GenesisRewardIndexesSlice   `json:"usdx_reward_indexes" yaml:"usdx_reward_indexes"`
	HardSupplyAccumulationTimes    GenesisAccumulationTimes    `json:"hard_supply_accumulation_times" yaml:"hard_supply_accumulation_times"`
	HardSupplyRewardIndexes        GenesisRewardIndexesSlice   `json:"hard_supply_reward_indexes" yaml:"hard_supply_reward_indexes"`
	HardBorrowAccumulationTimes    GenesisAccumulationTimes    `json:"hard_borrow_accumulation_times" yaml:"hard_borrow_accumulation_times"`
	HardBorrowRewardIndexes        GenesisRewardIndexesSlice   `json:"hard_borrow_reward_indexes" yaml:"hard_borrow_reward_indexes"`
	HardDelegatorAccumulationTimes GenesisAccumulationTimes    `json:"hard_delegator_accumulation_times" yaml:"hard_delegator_accumulation_times"`
	HardDelegatorRewardIndexes     GenesisRewardIndexesSlice   `json:"hard_delegator_reward_indexes" yaml:"hard_delegator_reward_indexes"`
	USDXMintingClaims              USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"`
	HardLiquidityProviderClaims    HardLiquidityProviderClaims `json:"hard_liquidity_provider_claims" yaml:"hard_liquidity_provider_claims"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(
	params Params,
	usdxAccumTimes, hardSupplyAccumTimes, hardBorrowAccumTimes, hardDelegatorAccumTimes GenesisAccumulationTimes,
	usdxIndexes, hardSupplyIndexes, hardBorrowIndexes, hardDelegatorIndexes GenesisRewardIndexesSlice,
	c USDXMintingClaims,
	hc HardLiquidityProviderClaims,
) GenesisState {
	return GenesisState{
		Params:                         params,
		USDXAccumulationTimes:          usdxAccumTimes,
		USDXRewardIndexes:              usdxIndexes,
		HardSupplyAccumulationTimes:    hardSupplyAccumTimes,
		HardSupplyRewardIndexes:        hardSupplyIndexes,
		HardBorrowAccumulationTimes:    hardBorrowAccumTimes,
		HardBorrowRewardIndexes:        hardBorrowIndexes,
		HardDelegatorAccumulationTimes: hardDelegatorAccumTimes,
		HardDelegatorRewardIndexes:     hardDelegatorIndexes,
		USDXMintingClaims:              c,
		HardLiquidityProviderClaims:    hc,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:                         DefaultParams(),
		USDXAccumulationTimes:          DefaultGenesisAccumulationTimes,
		USDXRewardIndexes:              DefaultGenesisRewardIndexes,
		HardSupplyAccumulationTimes:    DefaultGenesisAccumulationTimes,
		HardSupplyRewardIndexes:        DefaultGenesisRewardIndexes,
		HardBorrowAccumulationTimes:    DefaultGenesisAccumulationTimes,
		HardBorrowRewardIndexes:        DefaultGenesisRewardIndexes,
		HardDelegatorAccumulationTimes: DefaultGenesisAccumulationTimes,
		HardDelegatorRewardIndexes:     DefaultGenesisRewardIndexes,
		USDXMintingClaims:              DefaultUSDXClaims,
		HardLiquidityProviderClaims:    DefaultHardClaims,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if err := gs.USDXAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.USDXRewardIndexes.Validate(); err != nil {
		return err
	}
	for _, ri := range gs.USDXRewardIndexes {
		if len(ri.RewardIndexes) > 1 {
			return fmt.Errorf("USDX reward indexes cannot have more than one reward denom, found: %s", ri.RewardIndexes)
		}
	}

	if err := gs.HardSupplyAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardSupplyRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := gs.HardBorrowAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardBorrowRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := gs.HardDelegatorAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardDelegatorRewardIndexes.Validate(); err != nil {
		return err
	}
	for _, ri := range gs.HardDelegatorRewardIndexes {
		if len(ri.RewardIndexes) > 1 {
			return fmt.Errorf("Delegator reward indexes cannot have more than one reward denom, found: %s", ri.RewardIndexes)
		}
	}

	if err := gs.HardLiquidityProviderClaims.Validate(); err != nil {
		return err
	}
	if err := gs.USDXMintingClaims.Validate(); err != nil {
		return err
	}
	return nil
}

// GenesisAccumulationTime stores the previous reward distribution time and its corresponding collateral type
type GenesisAccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
}

// NewGenesisAccumulationTime returns a new GenesisAccumulationTime
func NewGenesisAccumulationTime(ctype string, prevTime time.Time) GenesisAccumulationTime {
	return GenesisAccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
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
	if len(gat.CollateralType) == 0 {
		return fmt.Errorf("genesis accumulation time's collateral type must be defined")
	}
	return nil
}

type GenesisRewardIndexes struct {
	CollateralType string        `json:"collateral_type" yaml:"collateral_type"`
	RewardIndexes  RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// NewGenesisRewardIndexes returns a new GenesisRewardIndexes
func NewGenesisRewardIndexes(ctype string, indexes RewardIndexes) GenesisRewardIndexes {
	return GenesisRewardIndexes{
		CollateralType: ctype,
		RewardIndexes:  indexes,
	}
}

func (gris GenesisRewardIndexes) Validate() error {
	if len(gris.CollateralType) == 0 {
		return fmt.Errorf("genesis reward indexes's collateral type must be defined")
	}
	if err := gris.RewardIndexes.Validate(); err != nil {
		return fmt.Errorf("invalid reward indexes: %v", err)
	}
	return nil
}

type GenesisRewardIndexesSlice []GenesisRewardIndexes

// Validate performs validation of GenesisAccumulationTimes
func (gris GenesisRewardIndexesSlice) Validate() error {
	for _, gri := range gris {
		if err := gri.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Params governance parameters for the incentive module
type Params struct {
	USDXMintingRewardPeriods   RewardPeriods      `json:"usdx_minting_reward_periods" yaml:"usdx_minting_reward_periods"`
	HardSupplyRewardPeriods    MultiRewardPeriods `json:"hard_supply_reward_periods" yaml:"hard_supply_reward_periods"`
	HardBorrowRewardPeriods    MultiRewardPeriods `json:"hard_borrow_reward_periods" yaml:"hard_borrow_reward_periods"`
	HardDelegatorRewardPeriods RewardPeriods      `json:"hard_delegator_reward_periods" yaml:"hard_delegator_reward_periods"`
	ClaimMultipliers           Multipliers        `json:"claim_multipliers" yaml:"claim_multipliers"`
	ClaimEnd                   time.Time          `json:"claim_end" yaml:"claim_end"`
}

// NewParams returns a new params object
func NewParams(usdxMinting RewardPeriods, hardSupply, hardBorrow MultiRewardPeriods,
	hardDelegator RewardPeriods, multipliers Multipliers, claimEnd time.Time) Params {
	return Params{
		USDXMintingRewardPeriods:   usdxMinting,
		HardSupplyRewardPeriods:    hardSupply,
		HardBorrowRewardPeriods:    hardBorrow,
		HardDelegatorRewardPeriods: hardDelegator,
		ClaimMultipliers:           multipliers,
		ClaimEnd:                   claimEnd,
	}
}

// DefaultParams returns default params for incentive module
func DefaultParams() Params {
	return NewParams(DefaultRewardPeriods, DefaultMultiRewardPeriods,
		DefaultMultiRewardPeriods, DefaultRewardPeriods, DefaultMultipliers, DefaultClaimEnd)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	USDX Minting Reward Periods: %s
	Hard Supply Reward Periods: %s
	Hard Borrow Reward Periods: %s
	Hard Delegator Reward Periods: %s
	Claim Multipliers :%s
	Claim End Time: %s
	`, p.USDXMintingRewardPeriods, p.HardSupplyRewardPeriods, p.HardBorrowRewardPeriods,
		p.HardDelegatorRewardPeriods, p.ClaimMultipliers, p.ClaimEnd)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyUSDXMintingRewardPeriods, &p.USDXMintingRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyHardSupplyRewardPeriods, &p.HardSupplyRewardPeriods, validateMultiRewardPeriodsParam),
		params.NewParamSetPair(KeyHardBorrowRewardPeriods, &p.HardBorrowRewardPeriods, validateMultiRewardPeriodsParam),
		params.NewParamSetPair(KeyHardDelegatorRewardPeriods, &p.HardDelegatorRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyClaimEnd, &p.ClaimEnd, validateClaimEndParam),
		params.NewParamSetPair(KeyMultipliers, &p.ClaimMultipliers, validateMultipliersParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {

	if err := validateMultipliersParam(p.ClaimMultipliers); err != nil {
		return err
	}

	if err := validateRewardPeriodsParam(p.USDXMintingRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.HardSupplyRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.HardBorrowRewardPeriods); err != nil {
		return err
	}

	return validateRewardPeriodsParam(p.HardDelegatorRewardPeriods)
}

func validateRewardPeriodsParam(i interface{}) error {
	rewards, ok := i.(RewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

func validateMultiRewardPeriodsParam(i interface{}) error {
	rewards, ok := i.(MultiRewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

func validateMultipliersParam(i interface{}) error {
	multipliers, ok := i.(Multipliers)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return multipliers.Validate()
}

func validateClaimEndParam(i interface{}) error {
	endTime, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if endTime.Unix() <= 0 {
		return fmt.Errorf("end time should not be zero")
	}
	return nil
}

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	Active           bool      `json:"active" yaml:"active"`
	CollateralType   string    `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time `json:"start" yaml:"start"`
	End              time.Time `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coin  `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}

// String implements fmt.Stringer
func (rp RewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Collateral Type: %s,
	Start: %s,
	End: %s,
	Rewards Per Second: %s,
	Active %t,
	`, rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond, rp.Active)
}

// NewRewardPeriod returns a new RewardPeriod
func NewRewardPeriod(active bool, collateralType string, start time.Time, end time.Time, reward sdk.Coin) RewardPeriod {
	return RewardPeriod{
		Active:           active,
		CollateralType:   collateralType,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
	}
}

// Validate performs a basic check of a RewardPeriod fields.
func (rp RewardPeriod) Validate() error {
	if rp.Start.Unix() <= 0 {
		return errors.New("reward period start time cannot be 0")
	}
	if rp.End.Unix() <= 0 {
		return errors.New("reward period end time cannot be 0")
	}
	if rp.Start.After(rp.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", rp.End, rp.Start)
	}
	if !rp.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", rp.RewardsPerSecond)
	}
	if strings.TrimSpace(rp.CollateralType) == "" {
		return fmt.Errorf("reward period collateral type cannot be blank: %s", rp)
	}
	return nil
}

// RewardPeriods array of RewardPeriod
type RewardPeriods []RewardPeriod

// Validate checks if all the RewardPeriods are valid and there are no duplicated
// entries.
func (rps RewardPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, rp := range rps {
		if seenPeriods[rp.CollateralType] {
			return fmt.Errorf("duplicated reward period with collateral type %s", rp.CollateralType)
		}

		if err := rp.Validate(); err != nil {
			return err
		}
		seenPeriods[rp.CollateralType] = true
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

// String implements fmt.Stringer
func (m Multiplier) String() string {
	return fmt.Sprintf(`Claim Multiplier:
	Name: %s
	Months Lockup %d
	Factor %s
	`, m.Name, m.MonthsLockup, m.Factor)
}

// Multipliers slice of Multiplier
type Multipliers []Multiplier

// Validate validates each multiplier
func (ms Multipliers) Validate() error {
	for _, m := range ms {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// String implements fmt.Stringer
func (ms Multipliers) String() string {
	out := "Claim Multipliers\n"
	for _, s := range ms {
		out += fmt.Sprintf("%s\n", s)
	}
	return out
}

// MultiplierName name for valid multiplier
type MultiplierName string

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return fmt.Errorf("invalid multiplier name: %s", mn)
}

// Claim is an interface for handling common claim actions
type Claim interface {
	GetOwner() sdk.AccAddress
	GetReward() sdk.Coin
	GetType() string
}

// Claims is a slice of Claim
type Claims []Claim

// BaseClaim is a common type shared by all Claims
type BaseClaim struct {
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward sdk.Coin       `json:"reward" yaml:"reward"`
}

// GetOwner is a getter for Claim Owner
func (c BaseClaim) GetOwner() sdk.AccAddress { return c.Owner }

// GetReward is a getter for Claim Reward
func (c BaseClaim) GetReward() sdk.Coin { return c.Reward }

// GetType returns the claim type, used to identify auctions in event attributes
func (c BaseClaim) GetType() string { return "base" }

// Validate performs a basic check of a BaseClaim fields
func (c BaseClaim) Validate() error {
	if c.Owner.Empty() {
		return errors.New("claim owner cannot be empty")
	}
	if !c.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", c.Reward)
	}
	return nil
}

// String implements fmt.Stringer
func (c BaseClaim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Reward: %s,
	`, c.Owner, c.Reward)
}

// BaseMultiClaim is a common type shared by all Claims with multiple reward denoms
type BaseMultiClaim struct {
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward sdk.Coins      `json:"reward" yaml:"reward"`
}

// GetOwner is a getter for Claim Owner
func (c BaseMultiClaim) GetOwner() sdk.AccAddress { return c.Owner }

// GetReward is a getter for Claim Reward
func (c BaseMultiClaim) GetReward() sdk.Coins { return c.Reward }

// GetType returns the claim type, used to identify auctions in event attributes
func (c BaseMultiClaim) GetType() string { return "base" }

// Validate performs a basic check of a BaseClaim fields
func (c BaseMultiClaim) Validate() error {
	if c.Owner.Empty() {
		return errors.New("claim owner cannot be empty")
	}
	if !c.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", c.Reward)
	}
	return nil
}

// String implements fmt.Stringer
func (c BaseMultiClaim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Reward: %s,
	`, c.Owner, c.Reward)
}

// -------------- Custom Claim Types --------------

// USDXMintingClaim is for USDX minting rewards
type USDXMintingClaim struct {
	BaseClaim     `json:"base_claim" yaml:"base_claim"`
	RewardIndexes RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// NewUSDXMintingClaim returns a new USDXMintingClaim
func NewUSDXMintingClaim(owner sdk.AccAddress, reward sdk.Coin, rewardIndexes RewardIndexes) USDXMintingClaim {
	return USDXMintingClaim{
		BaseClaim: BaseClaim{
			Owner:  owner,
			Reward: reward,
		},
		RewardIndexes: rewardIndexes,
	}
}

// GetType returns the claim's type
func (c USDXMintingClaim) GetType() string { return USDXMintingClaimType }

// GetReward returns the claim's reward coin
func (c USDXMintingClaim) GetReward() sdk.Coin { return c.Reward }

// GetOwner returns the claim's owner
func (c USDXMintingClaim) GetOwner() sdk.AccAddress { return c.Owner }

// Validate performs a basic check of a Claim fields
func (c USDXMintingClaim) Validate() error {
	if err := c.RewardIndexes.Validate(); err != nil {
		return err
	}

	return c.BaseClaim.Validate()
}

// String implements fmt.Stringer
func (c USDXMintingClaim) String() string {
	return fmt.Sprintf(`%s
	Reward Indexes: %s,
	`, c.BaseClaim, c.RewardIndexes)
}

// HasRewardIndex check if a claim has a reward index for the input collateral type
func (c USDXMintingClaim) HasRewardIndex(collateralType string) (int64, bool) {
	for index, ri := range c.RewardIndexes {
		if ri.CollateralType == collateralType {
			return int64(index), true
		}
	}
	return 0, false
}

// USDXMintingClaims slice of USDXMintingClaim
type USDXMintingClaims []USDXMintingClaim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs USDXMintingClaims) Validate() error {
	for _, c := range cs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HardLiquidityProviderClaim stores the hard liquidity provider rewards that can be claimed by owner
type HardLiquidityProviderClaim struct {
	BaseMultiClaim         `json:"base_claim" yaml:"base_claim"`
	SupplyRewardIndexes    MultiRewardIndexes `json:"supply_reward_indexes" yaml:"supply_reward_indexes"`
	BorrowRewardIndexes    MultiRewardIndexes `json:"borrow_reward_indexes" yaml:"borrow_reward_indexes"`
	DelegatorRewardIndexes RewardIndexes      `json:"delegator_reward_indexes" yaml:"delegator_reward_indexes"`
}

// NewHardLiquidityProviderClaim returns a new HardLiquidityProviderClaim
func NewHardLiquidityProviderClaim(owner sdk.AccAddress, rewards sdk.Coins, supplyRewardIndexes,
	borrowRewardIndexes MultiRewardIndexes, delegatorRewardIndexes RewardIndexes) HardLiquidityProviderClaim {
	return HardLiquidityProviderClaim{
		BaseMultiClaim: BaseMultiClaim{
			Owner:  owner,
			Reward: rewards,
		},
		SupplyRewardIndexes:    supplyRewardIndexes,
		BorrowRewardIndexes:    borrowRewardIndexes,
		DelegatorRewardIndexes: delegatorRewardIndexes,
	}
}

// GetType returns the claim's type
func (c HardLiquidityProviderClaim) GetType() string { return HardLiquidityProviderClaimType }

// GetReward returns the claim's reward coin
func (c HardLiquidityProviderClaim) GetReward() sdk.Coins { return c.Reward }

// GetOwner returns the claim's owner
func (c HardLiquidityProviderClaim) GetOwner() sdk.AccAddress { return c.Owner }

// Validate performs a basic check of a HardLiquidityProviderClaim fields
func (c HardLiquidityProviderClaim) Validate() error {
	if err := c.SupplyRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := c.BorrowRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := c.DelegatorRewardIndexes.Validate(); err != nil {
		return err
	}

	return c.BaseMultiClaim.Validate()
}

// String implements fmt.Stringer
func (c HardLiquidityProviderClaim) String() string {
	return fmt.Sprintf(`%s
	Supply Reward Indexes: %s,
	Borrow Reward Indexes: %s,
	Delegator Reward Indexes: %s,
	`, c.BaseMultiClaim, c.SupplyRewardIndexes, c.BorrowRewardIndexes, c.DelegatorRewardIndexes)
}

// HasSupplyRewardIndex check if a claim has a supply reward index for the input collateral type
func (c HardLiquidityProviderClaim) HasSupplyRewardIndex(denom string) (int64, bool) {
	for index, ri := range c.SupplyRewardIndexes {
		if ri.CollateralType == denom {
			return int64(index), true
		}
	}
	return 0, false
}

// HasBorrowRewardIndex check if a claim has a borrow reward index for the input collateral type
func (c HardLiquidityProviderClaim) HasBorrowRewardIndex(denom string) (int64, bool) {
	for index, ri := range c.BorrowRewardIndexes {
		if ri.CollateralType == denom {
			return int64(index), true
		}
	}
	return 0, false
}

// HasDelegatorRewardIndex check if a claim has a delegator reward index for the input collateral type
func (c HardLiquidityProviderClaim) HasDelegatorRewardIndex(collateralType string) (int64, bool) {
	for index, ri := range c.DelegatorRewardIndexes {
		if ri.CollateralType == collateralType {
			return int64(index), true
		}
	}
	return 0, false
}

// HardLiquidityProviderClaims slice of HardLiquidityProviderClaim
type HardLiquidityProviderClaims []HardLiquidityProviderClaim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs HardLiquidityProviderClaims) Validate() error {
	for _, c := range cs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ---------------------- Reward periods are used by the params ----------------------

// MultiRewardPeriod supports multiple reward types
type MultiRewardPeriod struct {
	Active           bool      `json:"active" yaml:"active"`
	CollateralType   string    `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time `json:"start" yaml:"start"`
	End              time.Time `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coins `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}

// String implements fmt.Stringer
func (mrp MultiRewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Collateral Type: %s,
	Start: %s,
	End: %s,
	Rewards Per Second: %s,
	Active %t,
	`, mrp.CollateralType, mrp.Start, mrp.End, mrp.RewardsPerSecond, mrp.Active)
}

// NewMultiRewardPeriod returns a new MultiRewardPeriod
func NewMultiRewardPeriod(active bool, collateralType string, start time.Time, end time.Time, reward sdk.Coins) MultiRewardPeriod {
	return MultiRewardPeriod{
		Active:           active,
		CollateralType:   collateralType,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
	}
}

// Validate performs a basic check of a MultiRewardPeriod.
func (mrp MultiRewardPeriod) Validate() error {
	if mrp.Start.IsZero() {
		return errors.New("reward period start time cannot be 0")
	}
	if mrp.End.IsZero() {
		return errors.New("reward period end time cannot be 0")
	}
	if mrp.Start.After(mrp.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", mrp.End, mrp.Start)
	}
	if !mrp.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", mrp.RewardsPerSecond)
	}
	if strings.TrimSpace(mrp.CollateralType) == "" {
		return fmt.Errorf("reward period collateral type cannot be blank: %s", mrp)
	}
	return nil
}

// MultiRewardPeriods array of MultiRewardPeriod
type MultiRewardPeriods []MultiRewardPeriod

// GetMultiRewardPeriod fetches a MultiRewardPeriod from an array of MultiRewardPeriods by its denom
func (mrps MultiRewardPeriods) GetMultiRewardPeriod(denom string) (MultiRewardPeriod, bool) {
	for _, rp := range mrps {
		if rp.CollateralType == denom {
			return rp, true
		}
	}
	return MultiRewardPeriod{}, false
}

// GetMultiRewardPeriodIndex returns the index of a MultiRewardPeriod inside array MultiRewardPeriods
func (mrps MultiRewardPeriods) GetMultiRewardPeriodIndex(denom string) (int, bool) {
	for i, rp := range mrps {
		if rp.CollateralType == denom {
			return i, true
		}
	}
	return -1, false
}

// Validate checks if all the RewardPeriods are valid and there are no duplicated
// entries.
func (mrps MultiRewardPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, rp := range mrps {
		if seenPeriods[rp.CollateralType] {
			return fmt.Errorf("duplicated reward period with collateral type %s", rp.CollateralType)
		}

		if err := rp.Validate(); err != nil {
			return err
		}
		seenPeriods[rp.CollateralType] = true
	}

	return nil
}

// ---------------------- Reward indexes are used internally in the store ----------------------

// RewardIndex stores reward accumulation information
type RewardIndex struct {
	CollateralType string  `json:"collateral_type" yaml:"collateral_type"`
	RewardFactor   sdk.Dec `json:"reward_factor" yaml:"reward_factor"`
}

// NewRewardIndex returns a new RewardIndex
func NewRewardIndex(collateralType string, factor sdk.Dec) RewardIndex {
	return RewardIndex{
		CollateralType: collateralType,
		RewardFactor:   factor,
	}
}

func (ri RewardIndex) String() string {
	return fmt.Sprintf(`Collateral Type: %s, RewardFactor: %s`, ri.CollateralType, ri.RewardFactor)
}

// Validate validates reward index
func (ri RewardIndex) Validate() error {
	if ri.RewardFactor.IsNegative() {
		return fmt.Errorf("reward factor value should be positive, is %s for %s", ri.RewardFactor, ri.CollateralType)
	}
	if strings.TrimSpace(ri.CollateralType) == "" {
		return fmt.Errorf("collateral type should not be empty")
	}
	return nil
}

// RewardIndexes slice of RewardIndex
type RewardIndexes []RewardIndex

// GetRewardIndex fetches a RewardIndex by its denom
func (ris RewardIndexes) GetRewardIndex(denom string) (RewardIndex, bool) {
	for _, ri := range ris {
		if ri.CollateralType == denom {
			return ri, true
		}
	}
	return RewardIndex{}, false
}

// Get fetches a RewardFactor by it's denom
func (ris RewardIndexes) Get(denom string) (sdk.Dec, bool) {
	for _, ri := range ris {
		if ri.CollateralType == denom {
			return ri.RewardFactor, true
		}
	}
	return sdk.Dec{}, false
}

// With returns a copy of the indexes with a new reward factor added
func (ris RewardIndexes) With(denom string, factor sdk.Dec) RewardIndexes {
	newIndexes := make(RewardIndexes, len(ris))
	copy(newIndexes, ris)

	for i, ri := range newIndexes {
		if ri.CollateralType == denom {
			newIndexes[i].RewardFactor = factor
			return newIndexes
		}
	}
	return append(newIndexes, NewRewardIndex(denom, factor))
}

// GetFactorIndex gets the index of a specific reward index inside the array by its index
func (ris RewardIndexes) GetFactorIndex(denom string) (int, bool) {
	for i, ri := range ris {
		if ri.CollateralType == denom {
			return i, true
		}
	}
	return -1, false
}

// Validate validation for reward indexes
func (ris RewardIndexes) Validate() error {
	for _, ri := range ris {
		if err := ri.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MultiRewardIndex stores reward accumulation information on multiple reward types
type MultiRewardIndex struct {
	CollateralType string        `json:"collateral_type" yaml:"collateral_type"`
	RewardIndexes  RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// NewMultiRewardIndex returns a new MultiRewardIndex
func NewMultiRewardIndex(collateralType string, indexes RewardIndexes) MultiRewardIndex {
	return MultiRewardIndex{
		CollateralType: collateralType,
		RewardIndexes:  indexes,
	}
}

// GetFactorIndex gets the index of a specific reward index inside the array by its index
func (mri MultiRewardIndex) GetFactorIndex(denom string) (int, bool) {
	for i, ri := range mri.RewardIndexes {
		if ri.CollateralType == denom {
			return i, true
		}
	}
	return -1, false
}

func (mri MultiRewardIndex) String() string {
	return fmt.Sprintf(`Collateral Type: %s, Reward Indexes: %s`, mri.CollateralType, mri.RewardIndexes)
}

// Validate validates multi-reward index
func (mri MultiRewardIndex) Validate() error {
	for _, rf := range mri.RewardIndexes {
		if rf.RewardFactor.IsNegative() {
			return fmt.Errorf("reward index's factor value cannot be negative: %s", rf)
		}
	}
	if strings.TrimSpace(mri.CollateralType) == "" {
		return fmt.Errorf("collateral type should not be empty")
	}
	return nil
}

// MultiRewardIndexes slice of MultiRewardIndex
type MultiRewardIndexes []MultiRewardIndex

// GetRewardIndex fetches a RewardIndex from a MultiRewardIndex by its denom
func (mris MultiRewardIndexes) GetRewardIndex(denom string) (MultiRewardIndex, bool) {
	for _, ri := range mris {
		if ri.CollateralType == denom {
			return ri, true
		}
	}
	return MultiRewardIndex{}, false
}

// Get fetches a RewardIndexes by it's denom
func (mris MultiRewardIndexes) Get(denom string) (RewardIndexes, bool) {
	for _, mri := range mris {
		if mri.CollateralType == denom {
			return mri.RewardIndexes, true
		}
	}
	return nil, false
}

// GetRewardIndexIndex fetches a specific reward index inside the array by its denom
func (mris MultiRewardIndexes) GetRewardIndexIndex(denom string) (int, bool) {
	for i, ri := range mris {
		if ri.CollateralType == denom {
			return i, true
		}
	}
	return -1, false
}

// With returns a copy of the indexes with a new RewardIndexes added
func (mris MultiRewardIndexes) With(denom string, indexes RewardIndexes) MultiRewardIndexes {
	newIndexes := mris.copy()

	for i, mri := range newIndexes {
		if mri.CollateralType == denom {
			newIndexes[i].RewardIndexes = indexes
			return newIndexes
		}
	}
	return append(newIndexes, NewMultiRewardIndex(denom, indexes))
}

// GetCollateralTypes returns a slice of containing all collateral types
func (mris MultiRewardIndexes) GetCollateralTypes() []string {
	var collateralTypes []string
	for _, ri := range mris {
		collateralTypes = append(collateralTypes, ri.CollateralType)
	}
	return collateralTypes
}

// RemoveRewardIndex removes a denom's reward interest factor value
func (mris MultiRewardIndexes) RemoveRewardIndex(denom string) MultiRewardIndexes {
	for i, ri := range mris {
		if ri.CollateralType == denom {
			// copy the slice and underlying array to avoid altering the original
			copy := mris.copy()
			return append(copy[:i], copy[i+1:]...)
		}
	}
	return mris
}

// Validate validation for reward indexes
func (mris MultiRewardIndexes) Validate() error {
	for _, mri := range mris {
		if err := mri.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// copy returns a copy of the slice and underlying array
func (mris MultiRewardIndexes) copy() MultiRewardIndexes {
	newIndexes := make(MultiRewardIndexes, len(mris))
	copy(newIndexes, mris)
	return newIndexes
}
