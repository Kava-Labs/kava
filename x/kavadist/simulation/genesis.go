package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// SecondsPerYear is the number of seconds in a year
const SecondsPerYear = 31536000

// BaseAprPadding prevents the calculated SPR inflation rate from being 0.0
const BaseAprPadding = "0.000000000100000000"

// RandomizedGenState generates a random GenesisState for kavadist module
func RandomizedGenState(simState *module.SimulationState) {
	params := genRandomParams(simState)
	if err := params.Validate(); err != nil {
		panic(err)
	}

	kavadistGenesis := types.NewGenesisState(params, types.DefaultPreviousBlockTime)
	if err := kavadistGenesis.Validate(); err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, kavadistGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(kavadistGenesis)
}

func genRandomParams(simState *module.SimulationState) types.Params {
	periods := genRandomPeriods(simState.Rand, simState.GenTimestamp)
	params := types.NewParams(true, periods)
	return params
}

func genRandomPeriods(r *rand.Rand, timestamp time.Time) types.Periods {
	var periods types.Periods
	numPeriods := simulation.RandIntBetween(r, 1, 10)
	periodStart := timestamp
	for i := 0; i < numPeriods; i++ {
		// set periods to be between 1-3 days
		durationMultiplier := simulation.RandIntBetween(r, 1, 3)
		duration := time.Duration(int64(24*durationMultiplier)) * time.Hour
		periodEnd := periodStart.Add(duration)
		inflation := genRandomInflation(r)
		period := types.NewPeriod(periodStart, periodEnd, inflation)
		periods = append(periods, period)
		periodStart = periodEnd
	}
	return periods
}

func genRandomInflation(r *rand.Rand) sdk.Dec {
	// If sim.RandomDecAmount returns 0 (happens frequently by design), add BaseAprPadding
	extraAprInflation := simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("0.25"))
	for extraAprInflation.Equal(sdk.ZeroDec()) {
		extraAprInflation = extraAprInflation.Add(sdk.MustNewDecFromStr(BaseAprPadding))
	}

	aprInflation := sdk.OneDec().Add(extraAprInflation)
	// convert APR inflation to SPR (inflation per second)
	inflationSpr, err := approxRoot(aprInflation, uint64(SecondsPerYear))
	if err != nil {
		panic(fmt.Sprintf("error generating random inflation %v", err))
	}
	return inflationSpr
}

func genRandomActive(r *rand.Rand) bool {
	threshold := 50
	value := simulation.RandIntBetween(r, 1, 100)
	if value > threshold {
		return true
	}
	return false
}
