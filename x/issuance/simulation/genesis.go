package simulation

import (
	"math/rand"
	"strings"

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
	types.NewGenesisState(params)
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

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
