package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/incentive/types"
)

var (
	CollateralDenoms    = []string{}
	RewardDenom         = "ukava"
	MaxTotalAssetReward = sdk.NewInt(1000000000)
)

// RandomizedGenState generates a random GenesisState for incentive module
func RandomizedGenState(simState *module.SimulationState) {

	// Get collateral asset denoms from existing CDP genesis state and pass to incentive params
	var cdpGenesis cdp.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[cdp.ModuleName], &cdpGenesis)
	if len(CollateralDenoms) == 0 {
		for _, collateral := range cdpGenesis.Params.CollateralParams {
			CollateralDenoms = append(CollateralDenoms, collateral.Type)
		}
	}

	params := genParams(simState.Rand)

	// Generate random reward and claim periods
	rewardPeriods := genRewardPeriods(simState.Rand, simState.GenTimestamp, params.Rewards)
	claimPeriods := genClaimPeriods(rewardPeriods)
	claimPeriodIDs := genNextClaimPeriodIds(claimPeriods)

	// New genesis state holds valid, linked reward periods, claim periods, and claim period IDs
	incentiveGenesis := types.NewGenesisState(params, types.DefaultPreviousBlockTime,
		rewardPeriods, claimPeriods, types.Claims{}, claimPeriodIDs)
	if err := incentiveGenesis.Validate(); err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, incentiveGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(incentiveGenesis)
}

// genParams generates random rewards and is active by default
func genParams(r *rand.Rand) types.Params {
	params := types.NewParams(true, genRewards(r))
	if err := params.Validate(); err != nil {
		panic(err)
	}
	return params
}

// genRewards generates rewards for each specified collateral type
func genRewards(r *rand.Rand) types.Rewards {
	rewards := make(types.Rewards, len(CollateralDenoms))
	for i, denom := range CollateralDenoms {
		active := true
		// total reward is in range (half max total reward, max total reward)
		amount := simulation.RandIntBetween(r, int(MaxTotalAssetReward.Int64()/2), int(MaxTotalAssetReward.Int64()))
		totalRewards := sdk.NewInt64Coin(RewardDenom, int64(amount))
		// generate a random number of hours between 6-48 to use for reward's times
		numbHours := simulation.RandIntBetween(r, 6, 48)
		duration := time.Duration(time.Hour * time.Duration(numbHours))
		timeLock := time.Duration(time.Hour * time.Duration(numbHours/2)) // half as long as duration
		claimDuration := time.Hour * time.Duration(numbHours*2)           // twice as long as duration
		rewards[i] = types.NewReward(active, denom, totalRewards, duration, timeLock, claimDuration)
	}
	return rewards
}

// genRewardPeriods generates chronological reward periods for each given reward type
func genRewardPeriods(r *rand.Rand, timestamp time.Time, rewards types.Rewards) types.RewardPeriods {
	rewardPeriods := make(types.RewardPeriods, len(rewards))
	rewardPeriodStart := timestamp

	for i, reward := range rewards {
		// Set up reward period parameters
		start := rewardPeriodStart
		end := start.Add(reward.Duration).UTC()
		baseRewardAmount := reward.AvailableRewards.Amount.Quo(sdk.NewInt(100)) // base period reward is 1/100 total reward
		// Earlier periods have larger rewards
		amount := sdk.NewCoin("ukava", baseRewardAmount.Mul(sdk.NewInt(int64(i))))
		claimEnd := end.Add(reward.ClaimDuration)
		claimTimeLock := reward.TimeLock
		// Create reward period and append to array
		rewardPeriods[i] = types.NewRewardPeriod(reward.CollateralType, start, end, amount, claimEnd, claimTimeLock)
		// Update start time of next reward period
		rewardPeriodStart = end
	}
	return rewardPeriods
}

// genClaimPeriods loads valid claim periods for an array of reward periods
func genClaimPeriods(rewardPeriods types.RewardPeriods) types.ClaimPeriods {
	denomRewardPeriodsCount := make(map[string]uint64)
	claimPeriods := make(types.ClaimPeriods, len(rewardPeriods))
	for i, rewardPeriod := range rewardPeriods {
		// Increment reward period count for this denom (this is our claim period's ID)
		denom := rewardPeriod.CollateralType
		numbRewardPeriods := denomRewardPeriodsCount[denom] + 1
		denomRewardPeriodsCount[denom] = numbRewardPeriods
		// Set end and timelock from the associated reward period
		end := rewardPeriod.ClaimEnd
		claimTimeLock := rewardPeriod.ClaimTimeLock
		// Create the new claim period for this reward period
		claimPeriods[i] = types.NewClaimPeriod(denom, numbRewardPeriods, end, claimTimeLock)
	}
	return claimPeriods
}

// genNextClaimPeriodIds returns an array of the most recent claim period IDs for each denom
func genNextClaimPeriodIds(cps types.ClaimPeriods) types.GenesisClaimPeriodIDs {
	// Build a map of the most recent claim periods by denom
	mostRecentClaimPeriodByDenom := make(map[string]uint64)
	var claimPeriodIDs types.GenesisClaimPeriodIDs
	for _, cp := range cps {
		if cp.ID <= mostRecentClaimPeriodByDenom[cp.CollateralType] {
			continue
		}
		claimPeriodIDs = append(claimPeriodIDs, types.GenesisClaimPeriodID{CollateralType: cp.CollateralType, ID: cp.ID})
		mostRecentClaimPeriodByDenom[cp.CollateralType] = cp.ID

	}
	return claimPeriodIDs
}
