package v0_12

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
)

var (
	GenesisTime = time.Date(2021, 3, 5, 6, 0, 0, 0, time.UTC)
	ChainID     = "kava-6"
)

func Migrate(genDoc tmtypes.GenesisDoc) tmtypes.GenesisDoc {
	cdc := app.MakeCodec()
	var appStateMap genutil.AppMap
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appStateMap); err != nil {
		panic(err)
	}
	newAppState := MigrateAppState(appStateMap)
	marshaledNewAppState, err := cdc.MarshalJSON(newAppState)
	if err != nil {
		panic(err)
	}
	genDoc.AppState = marshaledNewAppState

	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID

	return genDoc
}

func MigrateAppState(v0_12AppState genutil.AppMap) genutil.AppMap {
	cdc := app.MakeCodec()

	if v0_12AppState[bep3.ModuleName] != nil {
		var bep3GS bep3.GenesisState
		cdc.MustUnmarshalJSON(v0_12AppState[bep3.ModuleName], &bep3GS)
		delete(v0_12AppState, bep3.ModuleName)
		v0_12AppState[bep3.ModuleName] = cdc.MustMarshalJSON(Bep3(bep3GS))
	}
	return v0_12AppState

}

func Bep3(genesisState bep3.GenesisState) bep3.GenesisState {

	var newSwaps bep3.AtomicSwaps
	for _, swap := range genesisState.AtomicSwaps {
		if swap.Status == bep3.Completed {
			swap.ClosedBlock = 1 // reset closed block to one so completed swaps are removed from long term storage properly
		}
		if swap.Status == bep3.Open || swap.Status == bep3.Expired {
			swap.Status = bep3.Expired // set open swaps to expired so they can be refunded after chain start
			swap.ExpireHeight = 1      // set expire on first block as well to be safe
		}
		newSwaps = append(newSwaps, swap)
	}

	return bep3.NewGenesisState(genesisState.Params, newSwaps, genesisState.Supplies, genesisState.PreviousBlockTime)
}
