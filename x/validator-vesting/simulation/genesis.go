package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/x/validator-vesting/internal/types"
)

// RandomizedGenState generates a random GenesisState for validator-vesting
func RandomizedGenState(simState *module.SimulationState) {
	var authGenState authtypes.GenesisState
	authSimState := simState.GenState[authtypes.ModuleName]
	simState.Cdc.MustUnmarshalJSON(authSimState, &authGenState)
	var newGenesisAccs authexported.GenesisAccounts
	for _, acc := range authGenState.Accounts {
		va, ok := acc.(vestexported.VestingAccount)
		if ok {
			// 50% of the time convert the vesting account

			if simState.Rand.Intn(100) < 50 {
				bacc := authtypes.NewBaseAccountWithAddress(va.GetAddress())
				err := bacc.SetCoins(va.GetCoins())
				if err != nil {
					panic(err)
				}

				duration := va.GetEndTime() - va.GetStartTime()
				vestingPeriods := getRandomVestingPeriods(duration, simState.Rand, va.GetCoins())
				vestingCoins := getVestingCoins(vestingPeriods)
				bva := vestingtypes.NewBaseVestingAccount(&bacc, vestingCoins, va.GetEndTime())
				var gacc authexported.GenesisAccount
				if simState.Rand.Intn(100) < 50 {
					// convert to periodic vesting account 50%
					gacc = vestingtypes.NewPeriodicVestingAccountRaw(bva, va.GetStartTime(), vestingPeriods)
					err = gacc.Validate()
					if err != nil {
						panic(err)
					}
				} else {
					consAdd := getRandomValidatorConsAddr(simState, simulation.RandIntBetween(simState.Rand, 0, int(simState.NumBonded)-1))
					// convert to validator vesting account 50%
					// set signing threshold to be anywhere between 1 and 100
					gacc = types.NewValidatorVestingAccountRaw(
						bva, va.GetStartTime(), vestingPeriods, consAdd, nil,
						int64(simulation.RandIntBetween(simState.Rand, 1, 100)),
					)
					err = gacc.Validate()
					if err != nil {
						panic(err)
					}
				}
				newGenesisAccs = append(newGenesisAccs, gacc)
			} else {
				newGenesisAccs = append(newGenesisAccs, acc)
			}
		} else {
			newGenesisAccs = append(newGenesisAccs, acc)
		}
	}
	newAuthGenesis := authtypes.NewGenesisState(authGenState.Params, newGenesisAccs)
	simState.GenState[authtypes.ModuleName] = simState.Cdc.MustMarshalJSON(newAuthGenesis)
}

func getRandomValidatorConsAddr(simState *module.SimulationState, rint int) sdk.ConsAddress {
	acc := simState.Accounts[rint]
	return sdk.ConsAddress(acc.PubKey.Address())
}

func getRandomVestingPeriods(duration int64, r *rand.Rand, origCoins sdk.Coins) vestingtypes.Periods {
	maxPeriods := int64(50)
	if duration < maxPeriods {
		maxPeriods = duration
	}
	numPeriods := simulation.RandIntBetween(r, 1, int(maxPeriods))
	lenPeriod := duration / int64(numPeriods)
	periodLengths := make([]int64, numPeriods)
	totalLength := int64(0)
	for i := 0; i < numPeriods; i++ {
		periodLengths[i] = lenPeriod
		totalLength += lenPeriod
	}
	if duration-totalLength != 0 {
		periodLengths[len(periodLengths)-1] += (duration - totalLength)
	}

	coinFraction := simulation.RandIntBetween(r, 1, 100)
	vestingCoins := sdk.NewCoins()
	for _, ic := range origCoins {
		amountVesting := ic.Amount.Int64() / int64(coinFraction)
		vestingCoins = vestingCoins.Add(sdk.NewCoins(sdk.NewInt64Coin(ic.Denom, amountVesting)))
	}
	periodCoins := sdk.NewCoins()
	for _, c := range vestingCoins {
		amountPeriod := c.Amount.Int64() / int64(numPeriods)
		periodCoins = periodCoins.Add(sdk.NewCoins(sdk.NewInt64Coin(c.Denom, amountPeriod)))
	}

	vestingPeriods := make([]vestingtypes.Period, numPeriods)
	for i := 0; i < numPeriods; i++ {
		vestingPeriods[i] = vestingtypes.Period{Length: int64(periodLengths[i]), Amount: periodCoins}
	}

	return vestingPeriods

}

func getVestingCoins(periods vestingtypes.Periods) sdk.Coins {
	vestingCoins := sdk.NewCoins()
	for _, p := range periods {
		vestingCoins = vestingCoins.Add(p.Amount)
	}
	return vestingCoins
}
