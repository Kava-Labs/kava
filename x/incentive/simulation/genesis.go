package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/incentive/types"
)

var (
	CollateralDenoms                 = [3]string{"bnb", "xrp", "btc"}
	RewardDenom                      = "ukava"
	MaxTotalAssetReward              = sdk.NewInt(10000000000000000)
	BaseRewardPeriodReward           = sdk.NewInt(10000000)
	ClaimerStartingCollateralBalance = sdk.NewInt(10000000000)
)

// RandomizedGenState generates a random GenesisState for incentive module
func RandomizedGenState(simState *module.SimulationState) {
	params := GenParams(simState.Rand)
	rewardPeriods := GenRewardPeriods(simState.Rand, simState.GenTimestamp, params.Rewards)
	claimPeriods := GenClaimPeriods(rewardPeriods)
	claimPeriodIDs := GenNextClaimPeriodIds(claimPeriods)

	// New genesis state holds valid, linked reward periods, claim periods, and claim period IDs
	incentiveGenesis := types.NewGenesisState(
		params, types.DefaultPreviousBlockTime, rewardPeriods,
		claimPeriods, types.Claims{}, claimPeriodIDs)
	if err := incentiveGenesis.Validate(); err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, incentiveGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(incentiveGenesis)

	authGenesis, totalCoins := loadAuthGenState(simState, incentiveGenesis)
	simState.GenState[auth.ModuleName] = simState.Cdc.MustMarshalJSON(authGenesis)

	// Update supply to match amount of coins in auth
	var supplyGenesis supply.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[supply.ModuleName], &supplyGenesis)
	for _, totalAssetClaimerBalances := range totalCoins {
		supplyGenesis.Supply = supplyGenesis.Supply.Add(totalAssetClaimerBalances)
	}
	simState.GenState[supply.ModuleName] = simState.Cdc.MustMarshalJSON(supplyGenesis)
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
		// generate a random number of days between 7-30 to use for reward's times
		numbDays := simulation.RandIntBetween(r, 7, 30) * 24
		duration := time.Duration(time.Hour * time.Duration(numbDays))
		timeLock := time.Duration(time.Hour * time.Duration(numbDays*10)) // 10 times as long as duration
		claimDuration := time.Hour * time.Duration(numbDays*2)            // twice as long as duration
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
			baseRewardAmount := reward.Reward.Amount.Quo(BaseRewardPeriodReward)
			// Earlier periods have larger rewards
			amount := sdk.NewCoin(reward.Denom, baseRewardAmount.Mul(sdk.NewInt(int64(i*2))))
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
		timeLock := rewardPeriod.ClaimTimeLock
		// Create the new claim period for this reward period
		claimPeriod := types.NewClaimPeriod(denom, numbRewardPeriods, end, timeLock)
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

func loadAuthGenState(simState *module.SimulationState, incentiveGenesis types.GenesisState) (
	auth.GenesisState, []sdk.Coins) {
	var authGenesis auth.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[auth.ModuleName], &authGenesis)

	// Load the first 10 accounts
	claimers := LoadClaimers(authGenesis.Accounts)
	if len(claimers) != 10 {
		panic("not enough claimer accounts were loaded from auth's genesis accounts")
	}

	// Load starting balance of each collateral type to each claimer's account
	var totalCoins []sdk.Coins
	for _, claimer := range claimers {
		for _, reward := range incentiveGenesis.Params.Rewards {
			startingCoinBalance := sdk.NewCoins(sdk.NewCoin(reward.Denom, ClaimerStartingCollateralBalance))
			if err := claimer.SetCoins(claimer.GetCoins().Add(startingCoinBalance)); err != nil {
				panic(err)
			}
			totalCoins = append(totalCoins, startingCoinBalance)
		}
		// Update claimer's account in auth genesis
		authGenesis.Accounts = replaceOrAppendAccount(authGenesis.Accounts, claimer)
	}
	return authGenesis, totalCoins
}

// LoadClaimers loads the first 10 accounts from auth
func LoadClaimers(accounts []authexported.GenesisAccount) []authexported.GenesisAccount {
	var claimers []authexported.GenesisAccount
	for i, acc := range accounts {
		if i < 10 {
			claimers = append(claimers, acc)
		} else {
			break
		}
	}
	return claimers
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
