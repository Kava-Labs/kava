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

	// Remove module states that we don't want genutil to process.
	// The auth state might require removal here due to issues with migrating validator vesting.
	// TODO:
	// authState := appState[v039auth.ModuleName]
	// delete(appState, v039auth.ModuleName)
	// fmt.Println(authState)

	appState = v040genutil.Migrate(appState, clientCtx)
	appState = v043genutil.Migrate(appState, clientCtx)
	return appState
}
