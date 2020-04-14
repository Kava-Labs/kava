package simulation

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// RandomizedGenState generates a random GenesisState for cdp
func RandomizedGenState(simState *module.SimulationState) {

	params := genRandomParams(simState)
	genesis := types.NewGenesisState(params, types.DefaultPreviousBlockTime)

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, genesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}

func genRandomParams(simState *module.SimulationState) types.Params {
	periods := genRandomPeriods(simState)
	params := types.NewParams(true, periods)
	return params
}

func genRandomPeriods(simState *module.SimulationState) types.Periods {
	var periods types.Periods
	numPeriods := simulation.RandIntBetween(simState.Rand, 1, 10)
	periodStart := simState.GenTimestamp
	for i := 0; i < numPeriods; i++ {
		// set periods to be between 2 weeks and 2 years
		durationMultiplier := simulation.RandIntBetween(simState.Rand, 7, 104)
		duration := time.Duration(int64(24*durationMultiplier)) * time.Hour
		periodEnd := periodStart.Add(duration)
		inflation := genRandomInflation(simState)
		period := types.NewPeriod(periodStart, periodEnd, inflation)
		periods = append(periods, period)
		periodStart = periodEnd
	}
	return periods
}

func genRandomInflation(simState *module.SimulationState) sdk.Dec {
	aprInflation := sdk.OneDec().Add(simulation.RandomDecAmount(simState.Rand, sdk.MustNewDecFromStr("0.25")))
	// convert APR inflation to SPR (inflation per second)
	inflationSpr, err := approxRoot(aprInflation, uint64(31536000))
	if err != nil {
		panic(fmt.Sprintf("error generating random inflation %v", err))
	}
	return inflationSpr
}
