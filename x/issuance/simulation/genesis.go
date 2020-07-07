package simulation

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/issuance/types"
)

var (
	accs []simulation.Account
)

// RandomizedGenState generates a random GenesisState for the module
func RandomizedGenState(simState *module.SimulationState) {
	accs = simState.Accounts
	params := randomizedParams(simState.Rand)
	gs := types.NewGenesisState(params)
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, gs))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(gs)
}

func randomizedParams(r *rand.Rand) types.Params {
	assets := randomizedAssets(r)
	return types.NewParams(assets)
}

func randomizedAssets(r *rand.Rand) types.Assets {
	randomAssets := types.Assets{}
	numAssets := Max(1, r.Intn(5))
	for i := 0; i < numAssets; i++ {
		denom := strings.ToLower(simulation.RandStringOfLength(r, (r.Intn(3) + 3)))
		owner := randomOwner(r)
		paused := r.Intn(2) == 0
		randomAsset := types.NewAsset(owner.Address, denom, []sdk.AccAddress{}, paused)
		randomAssets = append(randomAssets, randomAsset)
	}
	return randomAssets
}

func randomOwner(r *rand.Rand) simulation.Account {
	acc, _ := simulation.RandomAcc(r, accs)
	return acc
}

// Max return max of two ints
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
