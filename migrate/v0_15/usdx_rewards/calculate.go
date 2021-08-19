package main

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	v0_15cdp "github.com/kava-labs/kava/x/cdp/types"
	v0_15incentive "github.com/kava-labs/kava/x/incentive"
)

type rewards map[string]sdk.Coin

type rewardsCalculator struct {
	claims        v0_15incentive.USDXMintingClaims
	globalIndexes v0_15incentive.RewardIndexes
	cdps          v0_15cdp.CDPs
}

func NewRewardsCalculator(claims v0_15incentive.USDXMintingClaims, globalIndexes v0_15incentive.RewardIndexes, cdps v0_15cdp.CDPs) rewardsCalculator {
	return rewardsCalculator{
		claims:        claims,
		globalIndexes: globalIndexes,
		cdps:          cdps,
	}
}

// Calculate synchronizes all of the claims and returns the new rewards to be added by owner address.
func (rc rewardsCalculator) Calculate() (rewards, error) {
	rewards := rewards{}

	for _, claim := range rc.claims {

		reward, err := rc.calculateRewardsForClaim(claim)
		if err != nil {
			return nil, err
		}

		if _, found := rewards[claim.Owner.String()]; found {
			return nil, fmt.Errorf("duplicate claim found: %s", claim.Owner)
		}
		if !reward.IsZero() {
			rewards[claim.Owner.String()] = reward
		}
	}
	return rewards, nil
}

// calculateRewardsForClaim synchronizes the claim to and returns the new rewards that should be added.
// It syncs against the stored global indexes, looking up cdps for the source share amounts.
func (rc rewardsCalculator) calculateRewardsForClaim(claim v0_15incentive.USDXMintingClaim) (sdk.Coin, error) {
	reward := sdk.ZeroInt()

	for _, index := range rc.globalIndexes {
		oldFactor, found := claim.RewardIndexes.Get(index.CollateralType)
		if !found {
			oldFactor = sdk.ZeroDec()
		}

		sourceShares := sdk.ZeroDec()
		cdp, found := rc.getCDP(claim.Owner, index.CollateralType)
		if found {
			sourceShares = cdp.GetTotalPrincipal().Amount.ToDec()
		}

		amount, err := rc.calculateSingleReward(oldFactor, index.RewardFactor, sourceShares)
		if err != nil {
			return sdk.Coin{}, err
		}

		reward = reward.Add(amount)
	}
	return sdk.NewCoin(v0_15incentive.USDXMintingRewardDenom, reward), nil
}

// calculateSingleReward computes how much rewards should have accrued to a reward source (eg a user's btcb-a cdp principal)
// between two index values.
func (rc rewardsCalculator) calculateSingleReward(oldIndex, newIndex, sourceShares sdk.Dec) (sdk.Int, error) {
	increase := newIndex.Sub(oldIndex)
	if increase.IsNegative() {
		return sdk.Int{}, sdkerrors.Wrapf(v0_15incentive.ErrDecreasingRewardFactor, "old: %v, new: %v", oldIndex, newIndex)
	}
	reward := increase.Mul(sourceShares).RoundInt()
	return reward, nil
}

// getCDP looks up a cdp by owner and collateral type.
func (rc rewardsCalculator) getCDP(owner sdk.AccAddress, collateralType string) (v0_15cdp.CDP, bool) {
	for _, cdp := range rc.cdps {
		if cdp.Owner.Equals(owner) && cdp.Type == collateralType {
			return cdp, true
		}
	}
	return v0_15cdp.CDP{}, false
}

func convertRewardIndexesToUSDXMintingIndexes(mris v0_15incentive.MultiRewardIndexes) v0_15incentive.RewardIndexes {
	var newIndexes v0_15incentive.RewardIndexes
	for _, mri := range mris {
		factor, found := mri.RewardIndexes.Get(v0_15incentive.USDXMintingRewardDenom)
		if !found {
			panic(fmt.Sprintf("found global usdx minting reward index without denom '%s': %s", v0_15incentive.USDXMintingRewardDenom, mri))
		}
		newIndexes = newIndexes.With(mri.CollateralType, factor)
	}
	return newIndexes
}
