package v0_15

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	v0_14committee "github.com/kava-labs/kava/x/committee/legacy/v0_14"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

var (
	// TODO: update GenesisTime and chain-id for kava-8 launch
	GenesisTime = time.Date(2021, 4, 8, 15, 0, 0, 0, time.UTC)
	ChainID     = "kava-8"
)

// Migrate translates a genesis file from kava v0.14 format to kava v0.15 format
func Migrate(genDoc tmtypes.GenesisDoc) tmtypes.GenesisDoc {
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

	MigrateAppState(appStateMap)

	v0_15Codec := app.MakeCodec()
	marshaledNewAppState, err := v0_15Codec.MarshalJSON(appStateMap)
	if err != nil {
		panic(err)
	}
	genDoc.AppState = marshaledNewAppState
	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID
	return genDoc
}

// MigrateAppState migrates application state from v0.14 format to a kava v0.15 format
// It modifies the provided genesis state in place.
func MigrateAppState(v0_14AppState genutil.AppMap) {
	v0_14Codec := makeV014Codec()
	v0_15Codec := app.MakeCodec()

	// Migrate commmittee app state
	if v0_14AppState[v0_14committee.ModuleName] != nil {
		// Unmarshal v14 committee genesis state and delete it
		var committeeGS v0_14committee.GenesisState
		v0_14Codec.MustUnmarshalJSON(v0_14AppState[v0_14committee.ModuleName], &committeeGS)
		delete(v0_14AppState, v0_14committee.ModuleName)
		// Marshal v15 committee genesis state
		v0_14AppState[v0_15committee.ModuleName] = v0_15Codec.MustMarshalJSON(Committee(committeeGS))
	}
	// Migrate incentive app state
	if v0_14AppState[v0_14incentive.ModuleName] != nil {
		var incentiveGS v0_14incentive.GenesisState
		v0_14Codec.MustUnmarshalJSON(v0_14AppState[v0_14incentive.ModuleName], &incentiveGS)
		delete(v0_14AppState, v0_14incentive.ModuleName)

		v0_14AppState[v0_15incentive.ModuleName] = v0_15Codec.MustMarshalJSON(Incentive(incentiveGS))
	}
}

func makeV014Codec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	v0_14committee.RegisterCodec(cdc)
	v0_14incentive.RegisterCodec(cdc)
	return cdc
}
