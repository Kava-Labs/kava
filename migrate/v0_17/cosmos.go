package v0_17

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client"

	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	v040authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func MigrateCosmosAppState(appState genutiltypes.AppMap, clientCtx client.Context, genesisTime time.Time) genutiltypes.AppMap {
	appState = migrateV040(appState, clientCtx, genesisTime)
	return appState
}

// reset periodic vesting data for accounts
func migrateV040(appState genutiltypes.AppMap, clientCtx client.Context, genesisTime time.Time) genutiltypes.AppMap {
	setConfigIfUnsealed()

	v040Codec := clientCtx.Codec
	// reset periodic vesting data for accounts
	if appState[v040auth.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var authGenState v040authtypes.GenesisState
		v040Codec.MustUnmarshalJSON(appState[v040auth.ModuleName], &authGenState)

		// reset periodic vesting data for accounts
		appState[v040auth.ModuleName] = v040Codec.MustMarshalJSON(MigrateAuthV040(authGenState, genesisTime, clientCtx))
	}

	return appState
}
