package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// SecondsPerYear is the number of seconds in a year
const (
	SecondsPerYear = 31536000
	// BaseAprPadding sets the minimum inflation to the calculated SPR inflation rate from being 0.0
	BaseAprPadding = "0.000000003022265980"
)

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
	numPeriods := simtypes.RandIntBetween(r, 1, 10)
	periodStart := timestamp
	for i := 0; i < numPeriods; i++ {
		// set periods to be between 1-3 days
		durationMultiplier := simtypes.RandIntBetween(r, 1, 3)
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
	// If sim.RandomDecAmount is less than base apr padding, add base apr padding
	aprPadding, _ := sdk.NewDecFromStr(BaseAprPadding)
	extraAprInflation := simtypes.RandomDecAmount(r, sdk.MustNewDecFromStr("0.25"))
	for extraAprInflation.LT(aprPadding) {
		extraAprInflation = extraAprInflation.Add(aprPadding)
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
	value := simtypes.RandIntBetween(r, 1, 100)
	return value > threshold
}
