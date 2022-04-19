package v0_17

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/kava-labs/kava/app"
	tmtypes "github.com/tendermint/tendermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

var (
	GenesisTime = time.Date(2022, 5, 10, 16, 0, 0, 0, time.UTC)
	ChainID     = "kava-2222-10"
)

func setConfigIfUnsealed() {
	config := sdk.GetConfig()
	if config.GetBech32AccountAddrPrefix() == "kava" {
		return
	}
	app.SetSDKConfig()
}

// Migrate converts v16 genesis doc to v17 genesis doc
func Migrate(genDoc *tmtypes.GenesisDoc, ctx client.Context) (*tmtypes.GenesisDoc, error) {
	setConfigIfUnsealed()

	var appState genutiltypes.AppMap
	var err error
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, fmt.Errorf("failed to marchal app state from genesis doc:  %w", err)
	}

	migrateNewModules(appState, ctx)

	genDoc.AppState, err = json.Marshal(appState)
	if err != nil {
		return nil, err
	}

	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID
	genDoc.InitialHeight = 1

	return genDoc, nil
}

func migrateNewModules(appState genutiltypes.AppMap, clientCtx client.Context) {
	// x/emvutil
	evmUtilGenState := evmutiltypes.NewGenesisState([]evmutiltypes.Account{})
	appState[evmutiltypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(evmUtilGenState)

	// x/evm
	evmGenState := &evmtypes.GenesisState{
		Accounts: []evmtypes.GenesisAccount{},
		Params: evmtypes.Params{
			EvmDenom:     "akava",
			EnableCreate: true,
			EnableCall:   true,
			ChainConfig:  evmtypes.DefaultChainConfig(),
			ExtraEIPs:    nil,
		},
	}
	appState[evmtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(evmGenState)
}
