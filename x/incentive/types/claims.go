package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	USDXMintingClaimType           = "usdx_minting"
	HardLiquidityProviderClaimType = "hard_liquidity_provider"
	BondDenom                      = "ukava"
)

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

// Get fetches a RewardFactor from the indexes by it's denom
func (ris RewardIndexes) Get(denom string) (sdk.Dec, bool) {
	for _, ri := range ris {
		if ri.CollateralType == denom {
			return ri.RewardFactor, true
		}
	}
	return sdk.Dec{}, false
}

// With returns a copy of the indexes with a new reward index added
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

// Get fetches a RewardFactor from the indexes by it's denom
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
	newIndexes := make(MultiRewardIndexes, len(mris))
	copy(newIndexes, mris)

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
func (mris MultiRewardIndexes) RemoveRewardIndex(denom string) MultiRewardIndexes { // TODO modifies original dangerously
	for i, ri := range mris {
		if ri.CollateralType == denom {
			return append(mris[:i], mris[i+1:]...)
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
