package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Claim stores the rewards that can be claimed by owner
type USDXMintingClaim struct {
	Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward        sdk.Coin       `json:"reward" yaml:"reward"`
	RewardIndexes RewardIndexes  `json:"reward_indexes" yaml:"reward_indexes"`
}

// NewUSDXMintingClaim returns a new USDXMintingClaim
func NewUSDXMintingClaim(owner sdk.AccAddress, reward sdk.Coin, rewardIndexes RewardIndexes) USDXMintingClaim {
	return USDXMintingClaim{
		Owner:         owner,
		Reward:        reward,
		RewardIndexes: rewardIndexes,
	}
}

// Validate performs a basic check of a Claim fields.
func (c USDXMintingClaim) Validate() error {
	if c.Owner.Empty() {
		return errors.New("claim owner cannot be empty")
	}
	if !c.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", c.Reward)
	}
	return c.RewardIndexes.Validate()
}

// String implements fmt.Stringer
func (c USDXMintingClaim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Reward: %s,
	Reward Indexes: %s,
	`, c.Owner, c.Reward, c.RewardIndexes)
}

// Claims array of Claim
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

// RewardIndex stores reward accumulation information
type RewardIndex struct {
	CollateralType string  `json:"collateral_type" yaml:"collateral_type"`
	RewardFactor   sdk.Dec `json:"reward_factor" yaml:"reward_factor"`
}

func NewRewardIndex(collateralType string, factor sdk.Dec) RewardIndex {
	return RewardIndex{
		CollateralType: collateralType,
		RewardFactor:   factor,
	}
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

// Validate validation for reward indexes
func (ris RewardIndexes) Validate() error {
	for _, ri := range ris {
		if err := ri.Validate(); err != nil {
			return err
		}
	}
	return nil
}
