package v0_15

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_15cdp "github.com/kava-labs/kava/x/cdp/types"
	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

// Incentive migrates from a v0.14 incentive genesis state to a v0.15 incentive genesis state
func Incentive(cdc *codec.Codec, incentiveGS v0_14incentive.GenesisState, cdps v0_15cdp.CDPs) v0_15incentive.GenesisState {
	// Migrate params
	claimMultipliers := v0_15incentive.Multipliers{}
	for _, m := range incentiveGS.Params.ClaimMultipliers {
		newMultiplier := v0_15incentive.NewMultiplier(v0_15incentive.MultiplierName(m.Name), m.MonthsLockup, m.Factor)
		claimMultipliers = append(claimMultipliers, newMultiplier)
	}
	newMultipliers := v0_15incentive.MultipliersPerDenom{
		{
			Denom:       "hard",
			Multipliers: claimMultipliers,
		},
		{
			Denom:       "ukava",
			Multipliers: claimMultipliers,
		},
		{
			Denom: "swp",
			Multipliers: v0_15incentive.Multipliers{
				{
					Name:         v0_15incentive.Small,
					MonthsLockup: 1,
					Factor:       sdk.MustNewDecFromStr("0.1"),
				},
				{
					Name:         v0_15incentive.Large,
					MonthsLockup: 12,
					Factor:       sdk.OneDec(),
				},
			},
		},
	}

	usdxMintingRewardPeriods := v0_15incentive.RewardPeriods{}
	for _, rp := range incentiveGS.Params.USDXMintingRewardPeriods {
		usdxMintingRewardPeriod := v0_15incentive.NewRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond)
		usdxMintingRewardPeriods = append(usdxMintingRewardPeriods, usdxMintingRewardPeriod)
	}

	delegatorRewardPeriods := v0_15incentive.MultiRewardPeriods{}
	for _, rp := range incentiveGS.Params.HardDelegatorRewardPeriods {
		rewardsPerSecond := sdk.NewCoins(rp.RewardsPerSecond, SwpDelegatorRewardsPerSecond)
		delegatorRewardPeriod := v0_15incentive.NewMultiRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rewardsPerSecond)
		delegatorRewardPeriods = append(delegatorRewardPeriods, delegatorRewardPeriod)
	}

	// TODO: finalize swap reward pool IDs, rewards per second, start/end times. Should swap rewards start active?
	swapRewardPeriods := v0_15incentive.MultiRewardPeriods{}

	// Build new params from migrated values
	params := v0_15incentive.NewParams(
		usdxMintingRewardPeriods,
		migrateMultiRewardPeriods(incentiveGS.Params.HardSupplyRewardPeriods),
		migrateMultiRewardPeriods(incentiveGS.Params.HardBorrowRewardPeriods),
		delegatorRewardPeriods,
		swapRewardPeriods,
		newMultipliers,
		incentiveGS.Params.ClaimEnd,
	)

	// Migrate accumulation times and reward indexes
	usdxGenesisRewardState := migrateGenesisRewardState(incentiveGS.USDXAccumulationTimes, incentiveGS.USDXRewardIndexes)
	hardSupplyGenesisRewardState := migrateGenesisRewardState(incentiveGS.HardSupplyAccumulationTimes, incentiveGS.HardSupplyRewardIndexes)
	hardBorrowGenesisRewardState := migrateGenesisRewardState(incentiveGS.HardBorrowAccumulationTimes, incentiveGS.HardBorrowRewardIndexes)
	delegatorGenesisRewardState := migrateGenesisRewardState(incentiveGS.HardDelegatorAccumulationTimes, incentiveGS.HardDelegatorRewardIndexes)
	swapGenesisRewardState := v0_15incentive.DefaultGenesisRewardState // There is no previous swap rewards so accumulation starts at genesis time.

	// Migrate USDX minting claims
	usdxMintingClaims := migrateUSDXMintingClaims(incentiveGS.USDXMintingClaims)
	usdxMintingFormattedIndexes := convertRewardIndexesToUSDXMintingIndexes(usdxGenesisRewardState.MultiRewardIndexes)
	usdxMintingClaims = replaceUSDXClaimIndexes(usdxMintingClaims, usdxMintingFormattedIndexes)
	usdxMintingClaims = ensureAllCDPsHaveClaims(usdxMintingClaims, cdps, usdxMintingFormattedIndexes)
	var missedRewards map[string]sdk.Coin
	cdc.MustUnmarshalJSON([]byte(missedUSDXMintingRewards), &missedRewards)
	usdxMintingClaims = addRewards(usdxMintingClaims, missedRewards)

	// Migrate Hard protocol claims (includes creating new Delegator claims)
	hardClaims := v0_15incentive.HardLiquidityProviderClaims{}
	delegatorClaims := v0_15incentive.DelegatorClaims{}
	for _, claim := range incentiveGS.HardLiquidityProviderClaims {
		// Migrate supply multi reward indexes
		supplyMultiRewardIndexes := migrateMultiRewardIndexes(claim.SupplyRewardIndexes)

		// Migrate borrow multi reward indexes
		borrowMultiRewardIndexes := migrateMultiRewardIndexes(claim.BorrowRewardIndexes)

		// Migrate delegator reward indexes to multi reward indexes inside DelegatorClaims
		delegatorMultiRewardIndexes := v0_15incentive.MultiRewardIndexes{}
		delegatorRewardIndexes := v0_15incentive.RewardIndexes{}
		for _, ri := range claim.DelegatorRewardIndexes {
			// TODO add checks to ensure old reward indexes are as expected
			delegatorRewardIndex := v0_15incentive.NewRewardIndex(v0_14incentive.HardLiquidityRewardDenom, ri.RewardFactor)
			delegatorRewardIndexes = append(delegatorRewardIndexes, delegatorRewardIndex)
		}
		// TODO should this include indexes if none exist on the old claim?
		delegatorMultiRewardIndex := v0_15incentive.NewMultiRewardIndex(v0_14incentive.BondDenom, delegatorRewardIndexes)
		delegatorMultiRewardIndexes = append(delegatorMultiRewardIndexes, delegatorMultiRewardIndex)

		// TODO: It's impossible to distinguish between rewards from delegation vs. liquidity providing
		//		 as they're all combined inside claim.Reward, so I'm just putting them all inside
		// 		 the hard claim to avoid duplicating rewards.
		delegatorClaim := v0_15incentive.NewDelegatorClaim(claim.Owner, sdk.NewCoins(), delegatorMultiRewardIndexes)
		delegatorClaims = append(delegatorClaims, delegatorClaim)

		hardClaim := v0_15incentive.NewHardLiquidityProviderClaim(claim.Owner, claim.Reward,
			supplyMultiRewardIndexes, borrowMultiRewardIndexes)
		hardClaims = append(hardClaims, hardClaim)
	}

	// Add Swap Claims
	swapClaims := v0_15incentive.DefaultSwapClaims

	return v0_15incentive.NewGenesisState(
		params,
		usdxGenesisRewardState,
		hardSupplyGenesisRewardState,
		hardBorrowGenesisRewardState,
		delegatorGenesisRewardState,
		swapGenesisRewardState,
		usdxMintingClaims,
		hardClaims,
		delegatorClaims,
		swapClaims,
	)
}

// migrateUSDXMintingClaims converts the a slice of v0.14 USDX minting claims into v0.15 USDX minting claims.
// As both types are the same underneath, this just converts types and does no other modification.
func migrateUSDXMintingClaims(oldClaims v0_14incentive.USDXMintingClaims) v0_15incentive.USDXMintingClaims {
	newClaims := v0_15incentive.USDXMintingClaims{}
	for _, oldClaim := range oldClaims {
		rewardIndexes := migrateRewardIndexes(oldClaim.RewardIndexes)
		usdxMintingClaim := v0_15incentive.NewUSDXMintingClaim(oldClaim.Owner, oldClaim.Reward, rewardIndexes)
		newClaims = append(newClaims, usdxMintingClaim)
	}
	return newClaims
}

// replaceUSDXClaimIndexes overwrites the reward indexes in all the claims with the current global indexes.
func replaceUSDXClaimIndexes(claims v0_15incentive.USDXMintingClaims, globalIndexes v0_15incentive.RewardIndexes) v0_15incentive.USDXMintingClaims {
	var amendedClaims v0_15incentive.USDXMintingClaims
	for _, claim := range claims {
		claim.RewardIndexes = globalIndexes
		amendedClaims = append(amendedClaims, claim)
	}
	return amendedClaims
}

// convertRewardIndexesToUSDXMintingIndexes converts a genesis reward indexes into the format used within usdx minting claims.
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

// ensureAllCDPsHaveClaims ensures that there is a claim for every cdp in the provided list.
// It uses the provided global indexes as the indexes for any added claim.
func ensureAllCDPsHaveClaims(newClaims v0_15incentive.USDXMintingClaims, cdps v0_15cdp.CDPs, globalIndexes v0_15incentive.RewardIndexes) v0_15incentive.USDXMintingClaims {
	for _, cdp := range cdps {

		claimFound := false
		for _, claim := range newClaims {
			if claim.Owner.Equals(cdp.Owner) {
				claimFound = true
				break
			}
		}

		if !claimFound {

			claim := v0_15incentive.NewUSDXMintingClaim(
				cdp.Owner,
				sdk.NewCoin(v0_15incentive.USDXMintingRewardDenom, sdk.ZeroInt()),
				globalIndexes,
			)
			newClaims = append(newClaims, claim)
		}

	}
	return newClaims
}

// addRewards adds some coins to a list of claims according to a map of address: coin.
// It panics if any coin denom doesn't match the denom in the claim.
func addRewards(newClaims v0_15incentive.USDXMintingClaims, rewards map[string]sdk.Coin) v0_15incentive.USDXMintingClaims {

	var amendedClaims v0_15incentive.USDXMintingClaims
	for _, claim := range newClaims {
		r, found := rewards[claim.Owner.String()]
		if found {
			claim.Reward = claim.Reward.Add(r)
		}
		amendedClaims = append(amendedClaims, claim)
	}
	return amendedClaims
}

func migrateMultiRewardPeriods(oldPeriods v0_14incentive.MultiRewardPeriods) v0_15incentive.MultiRewardPeriods {
	newPeriods := v0_15incentive.MultiRewardPeriods{}
	for _, rp := range oldPeriods {
		newPeriod := v0_15incentive.NewMultiRewardPeriod(
			rp.Active,
			rp.CollateralType,
			rp.Start,
			rp.End,
			rp.RewardsPerSecond,
		)
		newPeriods = append(newPeriods, newPeriod)
	}
	return newPeriods
}

func migrateGenesisRewardState(oldAccumulationTimes v0_14incentive.GenesisAccumulationTimes, oldIndexes v0_14incentive.GenesisRewardIndexesSlice) v0_15incentive.GenesisRewardState {
	accumulationTimes := v0_15incentive.AccumulationTimes{}
	for _, t := range oldAccumulationTimes {
		newAccumulationTime := v0_15incentive.NewAccumulationTime(t.CollateralType, t.PreviousAccumulationTime)
		accumulationTimes = append(accumulationTimes, newAccumulationTime)
	}
	multiRewardIndexes := v0_15incentive.MultiRewardIndexes{}
	for _, gri := range oldIndexes {
		multiRewardIndex := v0_15incentive.NewMultiRewardIndex(gri.CollateralType, migrateRewardIndexes(gri.RewardIndexes))
		multiRewardIndexes = append(multiRewardIndexes, multiRewardIndex)
	}
	return v0_15incentive.NewGenesisRewardState(
		accumulationTimes,
		multiRewardIndexes,
	)
}

func migrateMultiRewardIndexes(oldIndexes v0_14incentive.MultiRewardIndexes) v0_15incentive.MultiRewardIndexes {
	newIndexes := v0_15incentive.MultiRewardIndexes{}
	for _, mri := range oldIndexes {
		multiRewardIndex := v0_15incentive.NewMultiRewardIndex(
			mri.CollateralType,
			migrateRewardIndexes(mri.RewardIndexes),
		)
		newIndexes = append(newIndexes, multiRewardIndex)
	}
	return newIndexes
}

func migrateRewardIndexes(oldIndexes v0_14incentive.RewardIndexes) v0_15incentive.RewardIndexes {
	newIndexes := v0_15incentive.RewardIndexes{}
	for _, ri := range oldIndexes {
		rewardIndex := v0_15incentive.NewRewardIndex(ri.CollateralType, ri.RewardFactor)
		newIndexes = append(newIndexes, rewardIndex)
	}
	return newIndexes
}
