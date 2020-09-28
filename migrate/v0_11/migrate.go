package v0_11

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11incentive "github.com/kava-labs/kava/x/incentive"
	v0_9incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_9"
)

// MigrateBep3 migrates from a v0.9 (or v0.10) bep3 genesis state to a v0.11 bep3 genesis state
func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_11bep3.GenesisState {
	var assetParams v0_11bep3.AssetParams
	var assetSupplies v0_11bep3.AssetSupplies
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v10AssetParam := v0_11bep3.AssetParam{
			Active:        asset.Active,
			Denom:         asset.Denom,
			CoinID:        asset.CoinID,
			DeputyAddress: v0_9Params.BnbDeputyAddress,
			FixedFee:      v0_9Params.BnbDeputyFixedFee,
			MinSwapAmount: sdk.OneInt(), // set min swap to one - prevents accounts that hold zero bnb from creating spam txs
			MaxSwapAmount: v0_9Params.MaxAmount,
			MinBlockLock:  v0_9Params.MinBlockLock,
			MaxBlockLock:  v0_9Params.MaxBlockLock,
			SupplyLimit: v0_11bep3.SupplyLimit{
				Limit:          asset.Limit,
				TimeLimited:    false,
				TimePeriod:     time.Duration(0),
				TimeBasedLimit: sdk.ZeroInt(),
			},
		}
		assetParams = append(assetParams, v10AssetParam)
	}
	for _, supply := range oldGenState.AssetSupplies {
		newSupply := v0_11bep3.NewAssetSupply(supply.IncomingSupply, supply.OutgoingSupply, supply.CurrentSupply, sdk.NewCoin(supply.CurrentSupply.Denom, sdk.ZeroInt()), time.Duration(0))
		assetSupplies = append(assetSupplies, newSupply)
	}
	var swaps v0_11bep3.AtomicSwaps
	for _, oldSwap := range oldGenState.AtomicSwaps {
		newSwap := v0_11bep3.AtomicSwap{
			Amount:              oldSwap.Amount,
			RandomNumberHash:    oldSwap.RandomNumberHash,
			ExpireHeight:        oldSwap.ExpireHeight,
			Timestamp:           oldSwap.Timestamp,
			Sender:              oldSwap.Sender,
			Recipient:           oldSwap.Recipient,
			SenderOtherChain:    oldSwap.SenderOtherChain,
			RecipientOtherChain: oldSwap.RecipientOtherChain,
			ClosedBlock:         oldSwap.ClosedBlock,
			Status:              v0_11bep3.SwapStatus(oldSwap.Status),
			CrossChain:          oldSwap.CrossChain,
			Direction:           v0_11bep3.SwapDirection(oldSwap.Direction),
		}
		swaps = append(swaps, newSwap)
	}
	return v0_11bep3.GenesisState{
		Params: v0_11bep3.Params{
			AssetParams: assetParams},
		AtomicSwaps:       swaps,
		Supplies:          assetSupplies,
		PreviousBlockTime: v0_11bep3.DefaultPreviousBlockTime,
	}
}

// MigrateIncentive migrates from a v0.9 (or v0.10) incentive genesis state to a v0.11 incentive genesis state
func MigrateIncentive(oldGenState v0_9incentive.GenesisState) v0_11incentive.GenesisState {
	var newRewards v0_11incentive.Rewards
	var newRewardPeriods v0_11incentive.RewardPeriods
	var newClaimPeriods v0_11incentive.ClaimPeriods
	var newClaims v0_11incentive.Claims
	var newClaimPeriodIds v0_11incentive.GenesisClaimPeriodIDs

	newMultiplier := v0_11incentive.NewMultiplier(v0_11incentive.Large, 12, sdk.OneDec())

	for _, oldReward := range oldGenState.Params.Rewards {
		newReward := v0_11incentive.NewReward(oldReward.Active, oldReward.Denom+"-a", oldReward.AvailableRewards, oldReward.Duration, v0_11incentive.Multipliers{newMultiplier}, oldReward.ClaimDuration)
		newRewards = append(newRewards, newReward)
	}
	newParams := v0_11incentive.NewParams(true, newRewards)

	for _, oldRewardPeriod := range oldGenState.RewardPeriods {

		newRewardPeriod := v0_11incentive.NewRewardPeriod(oldRewardPeriod.Denom+"-a", oldRewardPeriod.Start, oldRewardPeriod.End, oldRewardPeriod.Reward, oldRewardPeriod.ClaimEnd, v0_11incentive.Multipliers{newMultiplier})
		newRewardPeriods = append(newRewardPeriods, newRewardPeriod)
	}

	for _, oldClaimPeriod := range oldGenState.ClaimPeriods {
		newClaimPeriod := v0_11incentive.NewClaimPeriod(oldClaimPeriod.Denom+"-a", oldClaimPeriod.ID, oldClaimPeriod.End, v0_11incentive.Multipliers{newMultiplier})
		newClaimPeriods = append(newClaimPeriods, newClaimPeriod)
	}

	for _, oldClaim := range oldGenState.Claims {
		newClaim := v0_11incentive.NewClaim(oldClaim.Owner, oldClaim.Reward, oldClaim.Denom+"-a", oldClaim.ClaimPeriodID)
		newClaims = append(newClaims, newClaim)
	}

	for _, oldClaimPeriodID := range oldGenState.NextClaimPeriodIDs {
		newClaimPeriodID := v0_11incentive.GenesisClaimPeriodID{
			CollateralType: oldClaimPeriodID.Denom + "-a",
			ID:             oldClaimPeriodID.ID,
		}
		newClaimPeriodIds = append(newClaimPeriodIds, newClaimPeriodID)
	}

	return v0_11incentive.NewGenesisState(newParams, oldGenState.PreviousBlockTime, newRewardPeriods, newClaimPeriods, newClaims, newClaimPeriodIds)
}
