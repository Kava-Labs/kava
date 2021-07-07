package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	USDXMintingClaimType           = "usdx_minting"
	HardLiquidityProviderClaimType = "hard_liquidity_provider"
	DelegatorClaimType             = "delegator_claim"
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
	BaseMultiClaim      `json:"base_claim" yaml:"base_claim"`
	SupplyRewardIndexes MultiRewardIndexes `json:"supply_reward_indexes" yaml:"supply_reward_indexes"`
	BorrowRewardIndexes MultiRewardIndexes `json:"borrow_reward_indexes" yaml:"borrow_reward_indexes"`
}

// NewHardLiquidityProviderClaim returns a new HardLiquidityProviderClaim
func NewHardLiquidityProviderClaim(owner sdk.AccAddress, rewards sdk.Coins,
	supplyRewardIndexes, borrowRewardIndexes MultiRewardIndexes) HardLiquidityProviderClaim {
	return HardLiquidityProviderClaim{
		BaseMultiClaim: BaseMultiClaim{
			Owner:  owner,
			Reward: rewards,
		},
		SupplyRewardIndexes: supplyRewardIndexes,
		BorrowRewardIndexes: borrowRewardIndexes,
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

	return c.BaseMultiClaim.Validate()
}

// String implements fmt.Stringer
func (c HardLiquidityProviderClaim) String() string {
	return fmt.Sprintf(`%s
	Supply Reward Indexes: %s,
	Borrow Reward Indexes: %s,
	`, c.BaseMultiClaim, c.SupplyRewardIndexes, c.BorrowRewardIndexes)
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

// DelegatorClaim stores delegation rewards that can be claimed by owner
type DelegatorClaim struct {
	BaseMultiClaim `json:"base_claim" yaml:"base_claim"`
	RewardIndexes  MultiRewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// NewDelegatorClaim returns a new DelegatorClaim
func NewDelegatorClaim(owner sdk.AccAddress, rewards sdk.Coins, rewardIndexes MultiRewardIndexes) DelegatorClaim {
	return DelegatorClaim{
		BaseMultiClaim: BaseMultiClaim{
			Owner:  owner,
			Reward: rewards,
		},
		RewardIndexes: rewardIndexes,
	}
}

// GetType returns the claim's type
func (c DelegatorClaim) GetType() string { return DelegatorClaimType }

// GetReward returns the claim's reward coin
func (c DelegatorClaim) GetReward() sdk.Coins { return c.Reward }

// GetOwner returns the claim's owner
func (c DelegatorClaim) GetOwner() sdk.AccAddress { return c.Owner }

// Validate performs a basic check of a DelegatorClaim fields
func (c DelegatorClaim) Validate() error {
	if err := c.RewardIndexes.Validate(); err != nil {
		return err
	}

	return c.BaseMultiClaim.Validate()
}

// String implements fmt.Stringer
func (c DelegatorClaim) String() string {
	return fmt.Sprintf(`%s
	Reward Indexes: %s,
	`, c.BaseMultiClaim, c.RewardIndexes)
}

// HasRewardIndex checks if a DelegatorClaim has a reward index for the input collateral type
func (c DelegatorClaim) HasRewardIndex(collateralType string) (int64, bool) {
	for index, ri := range c.RewardIndexes {
		if ri.CollateralType == collateralType {
			return int64(index), true
		}
	}
	return 0, false
}

// DelegatorClaim slice of DelegatorClaim
type DelegatorClaims []DelegatorClaim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs DelegatorClaims) Validate() error {
	for _, c := range cs {
		if err := c.Validate(); err != nil {
			return err
		}
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
	newIndexes := ris.copy()

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

// Mul returns a copy of RewardIndexes with all factors multiplied by a single value.
func (ris RewardIndexes) Mul(multiplier sdk.Dec) RewardIndexes {
	newIndexes := ris.copy()

	for i := range newIndexes {
		newIndexes[i].RewardFactor = newIndexes[i].RewardFactor.Mul(multiplier)
	}
	return newIndexes
}

// Quo returns a copy of RewardIndexes with all factors divided by a single value.
// It uses sdk.Dec.Quo for the division.
func (ris RewardIndexes) Quo(divisor sdk.Dec) RewardIndexes {
	newIndexes := ris.copy()

	for i := range newIndexes {
		newIndexes[i].RewardFactor = newIndexes[i].RewardFactor.Quo(divisor)
	}
	return newIndexes
}

// Add combines two reward indexes by adding together factors with the same CollateralType.
// Any CollateralTypes unique to either reward indexes are included in the output as is.
func (ris RewardIndexes) Add(addend RewardIndexes) RewardIndexes {
	newIndexes := ris.copy()

	for _, addRi := range addend {
		found := false
		for i, origRi := range newIndexes {
			if origRi.CollateralType == addRi.CollateralType {
				found = true
				newIndexes[i].RewardFactor = newIndexes[i].RewardFactor.Add(addRi.RewardFactor)
			}
		}
		if !found {
			newIndexes = append(newIndexes, addRi)
		}
	}
	return newIndexes
}

// copy returns a copy of the reward indexes slice and underlying array
func (ris RewardIndexes) copy() RewardIndexes {
	if ris == nil { // return nil rather than empty slice when ris is nil
		return nil
	}
	newIndexes := make(RewardIndexes, len(ris))
	copy(newIndexes, ris)
	return newIndexes
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
