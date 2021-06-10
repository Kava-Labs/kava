package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/swap/types"
)

var (
	accs []simulation.Account
)

// GenSwapFee generates a random SwapFee in range [0.01, 1.00]
func GenSwapFee(r *rand.Rand) sdk.Dec {
	min := int(1)
	max := int(100)
	percentage := int64(r.Intn(int(max)-min) + min)
	return sdk.NewDec(percentage).Quo(sdk.NewDec(100))
}

// GenSupportedAssets generates random allowed pools
func GenAllowedPools(r *rand.Rand) types.AllowedPools {

	numPools := (r.Intn(10) + 1)
	pools := make(types.AllowedPools, numPools)
	for i := 0; i < numPools; i++ {
		tokenA, tokenB := genTokenDenoms(r)
		for strings.Compare(tokenA, tokenB) == 0 {
			tokenA, tokenB = genTokenDenoms(r)
		}
		newPool := types.NewAllowedPool(tokenA, tokenB)
		pools[i] = newPool
	}
	return pools
}

func genTokenDenoms(r *rand.Rand) (string, string) {
	tokenA := genTokenDenom(r)
	tokenB := genTokenDenom(r)
	for strings.Compare(tokenA, tokenB) == 0 {
		tokenA = genTokenDenom(r)
	}
	tokens := []string{tokenA, tokenB}
	sort.Strings(tokens)
	return tokens[0], tokens[1]
}

func genTokenDenom(r *rand.Rand) string {
	denom := strings.ToLower(simulation.RandStringOfLength(r, 3))
	for err := sdk.ValidateDenom(denom); err != nil; {
		denom = strings.ToLower(simulation.RandStringOfLength(r, 3))
	}
	return denom
}

// RandomizedGenState generates a random GenesisState
func RandomizedGenState(simState *module.SimulationState) {
	accs = simState.Accounts

	swapGenesis := loadRandomSwapGenState(simState)
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, swapGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(swapGenesis)
}

func loadRandomSwapGenState(simState *module.SimulationState) types.GenesisState {
	pools := GenAllowedPools(simState.Rand)
	swapFee := GenSwapFee(simState.Rand)

	swapGenesis := types.GenesisState{
		Params: types.Params{
			AllowedPools: pools,
			SwapFee:      swapFee,
		},
	}

	if err := swapGenesis.Validate(); err != nil {
		panic(err)
	}
	return swapGenesis
}
