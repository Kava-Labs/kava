package modules

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/kava-labs/kava/app"

	// v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	v040genutil "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v040"
	v043genutil "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v043"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func MigrateCosmosAppState(appState genutiltypes.AppMap, clientCtx client.Context) genutiltypes.AppMap {
	app.SetSDKConfig()

	// To Be Removed Notes:
	// To avoid copying over migrations from genutil that we can reuse, we can remove app states
	// from modules that we don't want genutil to process before using genutil.Migrate.
	// In most cases, its the v40 migrations that are going to be an issue since it processes all interfaces
	// and we have a number of interfaces registered with cosmos modules.
	// We will need to copy over the code and process those manually.
	// Example:
	// authState := appState[v039auth.ModuleName]
	// delete(appState, v039auth.ModuleName)
	// fmt.Println(authState)
	// MigrateAuthState(authState, clientCtx)

	appState = v040genutil.Migrate(appState, clientCtx)
	appState = v043genutil.Migrate(appState, clientCtx)
	return appState
}
