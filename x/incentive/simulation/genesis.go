package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/incentive/types"
)

var (
	CollateralDenoms    = [3]string{"bnb", "xrp", "btc"}
	RewardDenom         = "ukava"
	MaxTotalAssetReward = sdk.NewInt(1000000000)
)

// RandomizedGenState generates a random GenesisState for incentive module
func RandomizedGenState(simState *module.SimulationState) {
	params := GenParams(simState.Rand)
	rewardPeriods := GenRewardPeriods(simState.Rand, simState.GenTimestamp, params.Rewards)
	claimPeriods := GenClaimPeriods(rewardPeriods)
	claimPeriodIDs := GenNextClaimPeriodIds(claimPeriods)

	// New genesis state holds valid, linked reward periods, claim periods, and claim period IDs
	incentiveGenesis := types.NewGenesisState(params, types.DefaultPreviousBlockTime,
		rewardPeriods, claimPeriods, types.Claims{}, claimPeriodIDs)
	if err := incentiveGenesis.Validate(); err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, incentiveGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(incentiveGenesis)
}

// GenParams generates random rewards and is active by default
func GenParams(r *rand.Rand) types.Params {
	params := types.NewParams(true, GenRewards(r))
	if err := params.Validate(); err != nil {
		panic(err)
	}
	return params
}

// GenRewards generates rewards for each specified collateral type
func GenRewards(r *rand.Rand) types.Rewards {
	var rewards types.Rewards
	for _, denom := range CollateralDenoms {
		active := true
		// total reward is in range (half max total reward, max total reward)
		amount := simulation.RandIntBetween(r, int(MaxTotalAssetReward.Int64()/2), int(MaxTotalAssetReward.Int64()))
		totalRewards := sdk.NewInt64Coin(RewardDenom, int64(amount))
		// generate a random number of hours between 6-48 to use for reward's times
		numbHours := simulation.RandIntBetween(r, 6, 48)
		duration := time.Duration(time.Hour * time.Duration(numbHours))
		timeLock := time.Duration(time.Hour * time.Duration(numbHours/2)) // half as long as duration
		claimDuration := time.Hour * time.Duration(numbHours*2)           // twice as long as duration
		reward := types.NewReward(active, denom, totalRewards, duration, timeLock, claimDuration)
		rewards = append(rewards, reward)
	}
	return rewards
}

// GenRewardPeriods generates chronological reward periods for each given reward type
func GenRewardPeriods(r *rand.Rand, timestamp time.Time, rewards types.Rewards) types.RewardPeriods {
	var rewardPeriods types.RewardPeriods
	for _, reward := range rewards {
		rewardPeriodStart := timestamp
		for i := 10; i >= simulation.RandIntBetween(r, 2, 9); i-- {
			// Set up reward period parameters
			start := rewardPeriodStart
			end := start.Add(reward.Duration).UTC()
			baseRewardAmount := reward.AvailableRewards.Amount.Quo(sdk.NewInt(100)) // base period reward is 1/100 total reward
			// Earlier periods have larger rewards
			amount := sdk.NewCoin(reward.Denom, baseRewardAmount.Mul(sdk.NewInt(int64(i))))
			claimEnd := end.Add(reward.ClaimDuration)
			claimTimeLock := reward.TimeLock
			// Create reward period and append to array
			rewardPeriod := types.NewRewardPeriod(reward.Denom, start, end, amount, claimEnd, claimTimeLock)
			rewardPeriods = append(rewardPeriods, rewardPeriod)
			// Update start time of next reward period
			rewardPeriodStart = end
		}
	}
	return rewardPeriods
}

// GenClaimPeriods loads valid claim periods for an array of reward periods
func GenClaimPeriods(rewardPeriods types.RewardPeriods) types.ClaimPeriods {
	denomRewardPeriodsCount := make(map[string]uint64)
	var claimPeriods types.ClaimPeriods
	for _, rewardPeriod := range rewardPeriods {
		// Increment reward period count for this denom (this is our claim period's ID)
		denom := rewardPeriod.Denom
		numbRewardPeriods := denomRewardPeriodsCount[denom] + 1
		denomRewardPeriodsCount[denom] = numbRewardPeriods
		// Set end and timelock from the associated reward period
		end := rewardPeriod.ClaimEnd
		claimTimeLock := rewardPeriod.ClaimTimeLock
		// Create the new claim period for this reward period
		claimPeriod := types.NewClaimPeriod(denom, numbRewardPeriods, end, claimTimeLock)
		claimPeriods = append(claimPeriods, claimPeriod)
	}
	return claimPeriods
}

// GenNextClaimPeriodIds returns an array of the most recent claim period IDs for each denom
func GenNextClaimPeriodIds(cps types.ClaimPeriods) types.GenesisClaimPeriodIDs {
	// Build a map of the most recent claim periods by denom
	mostRecentClaimPeriodByDenom := make(map[string]uint64)
	for _, cp := range cps {
		if cp.ID > mostRecentClaimPeriodByDenom[cp.Denom] {
			mostRecentClaimPeriodByDenom[cp.Denom] = cp.ID
		}
	}
	// Write map contents to an array of GenesisClaimPeriodIDs
	var claimPeriodIDs types.GenesisClaimPeriodIDs
	for key, value := range mostRecentClaimPeriodByDenom {
		claimPeriodID := types.GenesisClaimPeriodID{Denom: key, ID: value}
		claimPeriodIDs = append(claimPeriodIDs, claimPeriodID)
	}
	return claimPeriodIDs
}

// In a list of accounts, replace the first account found with the same address. If not found, append the account.
func replaceOrAppendAccount(accounts []authexported.GenesisAccount, acc authexported.GenesisAccount) []authexported.GenesisAccount {
	newAccounts := accounts
	for i, a := range accounts {
		if a.GetAddress().Equals(acc.GetAddress()) {
			newAccounts[i] = acc
			return newAccounts
		}
	}
	return append(newAccounts, acc)
}

// GenActive generates active bool with 80% chance of true
func GenActive(r *rand.Rand) bool {
	threshold := 80
	value := simulation.RandIntBetween(r, 1, 100)
	if value > threshold {
		return false
	}
	return true
}
