package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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
	params := genParams(simState.Rand)

	// New genesis state holds valid, linked reward periods, claim periods, and claim period IDs
	incentiveGenesis := types.NewGenesisState(params, types.DefaultPreviousBlockTime,
		types.RewardPeriods{}, types.ClaimPeriods{}, types.Claims{}, types.GenesisClaimPeriodIDs{})
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
