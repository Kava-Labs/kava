package v0_15

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/kava-labs/kava/app"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	tmtypes "github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// MigratePreview translates a genesis file from kava v0.14 format to kava v0.15 format for use in preview testnets, setting genesis time to 1 hour in the future.
func MigratePreview(genDoc tmtypes.GenesisDoc) tmtypes.GenesisDoc {
	genesisTime := tmtime.Now().Add(time.Hour).Truncate(time.Minute)
	// migrate app state
	var appStateMap genutil.AppMap
	cdc := codec.New()
	cryptoAmino.RegisterAmino(cdc)
	tmtypes.RegisterEvidences(cdc)

	// Old codec does not need all old modules registered on it to correctly decode at this stage
	// as it only decodes the app state into a map of module names to json encoded bytes.
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appStateMap); err != nil {
		panic(err)
	}

	MigrateAppState(appStateMap, genesisTime)
	MigrateStaking(appStateMap, &genDoc.Validators)

	v0_15Codec := app.MakeCodec()
	marshaledNewAppState, err := v0_15Codec.MarshalJSON(appStateMap)
	if err != nil {
		panic(err)
	}
	genDoc.AppState = marshaledNewAppState
	genDoc.GenesisTime = genesisTime
	genDoc.ChainID = "kava-8-preview"
	return genDoc
}
