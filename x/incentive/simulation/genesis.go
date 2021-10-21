package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kava-labs/kava/x/incentive/types"
)

var (
	CollateralDenoms    = []string{}
	RewardDenom         = "ukava"
	MaxTotalAssetReward = sdk.NewInt(1000000000)
)

// RandomizedGenState generates a random GenesisState for incentive module
func RandomizedGenState(simState *module.SimulationState) {

	// New genesis state holds valid, linked reward periods, claim periods, and claim period IDs
	incentiveGenesis := types.DefaultGenesisState()

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, incentiveGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(incentiveGenesis)
}
