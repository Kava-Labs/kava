package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Claim stores the rewards that can be claimed by owner
type Claim struct {
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward         sdk.Coin       `json:"reward" yaml:"reward"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	RewardIndex    RewardIndex    `json:"reward_index" yaml:"reward_index"`
}

// NewClaim returns a new Claim
func NewClaim(owner sdk.AccAddress, reward sdk.Coin, collateralType string, rewardIndex RewardIndex) Claim {
	return Claim{
		Owner:          owner,
		Reward:         reward,
		CollateralType: collateralType,
		RewardIndex:    rewardIndex,
	}
}

// Validate performs a basic check of a Claim fields.
func (c Claim) Validate() error {
	if c.Owner.Empty() {
		return errors.New("claim owner cannot be empty")
	}
	if !c.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", c.Reward)
	}
	if strings.TrimSpace(c.CollateralType) == "" {
		return fmt.Errorf("claim collateral type cannot be blank: %s", c)
	}
	return c.RewardIndex.Validate()
}

// String implements fmt.Stringer
func (c Claim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Collateral Type: %s,
	Reward: %s,
	Factor: %s,
	`, c.Owner, c.CollateralType, c.Reward, c.RewardIndex)
}

// Claims array of Claim
type Claims []Claim

// AugmentedClaim claim type with information about whether the claim is currently claimable
type AugmentedClaim struct {
	Claim Claim `json:"claim" yaml:"claim"`

	Claimable bool `json:"claimable" yaml:"claimable"`
}

func (ac AugmentedClaim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Collateral Type: %s,
	Reward: %s,
	Claimable: %t,
	`, ac.Claim.Owner, ac.Claim.CollateralType, ac.Claim.Reward, ac.Claimable,
	)
}

// NewAugmentedClaim returns a new augmented claim struct
func NewAugmentedClaim(claim Claim, claimable bool) AugmentedClaim {
	return AugmentedClaim{
		Claim:     claim,
		Claimable: claimable,
	}
}

// AugmentedClaims array of AugmentedClaim
type AugmentedClaims []AugmentedClaim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs Claims) Validate() error {
	for _, c := range cs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// RewardIndex stores reward accumulation information
type RewardIndex struct {
	Denom string  `json:"denom" yaml:"denom"`
	Value sdk.Dec `json:"value" yaml:"value"`
}

func NewRewardIndex(denom string, value sdk.Dec) RewardIndex {
	return RewardIndex{
		Denom: denom,
		Value: value,
	}
}

// Validate validates reward index
func (ri RewardIndex) Validate() error {
	if ri.Value.IsNegative() {
		return fmt.Errorf("reward index value should be positive, is %s for %s", ri.Value, ri.Denom)
	}
	return sdk.ValidateDenom(ri.Denom)
}
