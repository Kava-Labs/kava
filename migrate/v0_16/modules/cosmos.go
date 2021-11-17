package modules

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	v040genutil "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v040"
	v043genutil "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v043"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v043gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v043"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"

	"github.com/kava-labs/kava/app"
	v015kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_15"
)

func MigrateCosmosAppState(appState genutiltypes.AppMap, clientCtx client.Context) genutiltypes.AppMap {
	app.SetSDKConfig()

	// Remove x/gov from genutil migration so we can run our own migration on it
	govState := appState[v036gov.ModuleName]
	delete(appState, v036gov.ModuleName)

	appState = v040genutil.Migrate(appState, clientCtx)
	appState = v043genutil.Migrate(appState, clientCtx)

	// Create the codec needed to unmarshal v016 and marshal v016 app states
	v016Codec := clientCtx.Codec
	v015Codec := codec.NewLegacyAmino()
	v015kavadist.RegisterLegacyAminoCodec(v015Codec)
	v039auth.RegisterLegacyAminoCodec(v015Codec)
	v036gov.RegisterLegacyAminoCodec(v015Codec)
	v036params.RegisterLegacyAminoCodec(v015Codec)

	// Migrate x/gov
	if govState != nil {
		appState[v043gov.ModuleName] = migrateGov(govState, v015Codec, v016Codec)
	}

	return appState
}

// MigrateGov migrates x/gov from cosmos sdk v036 to v043
func migrateGov(govJson json.RawMessage, v015Codec *codec.LegacyAmino, v016Codec codec.Codec) []byte {
	// unmarshal v036 gov state
	var oldGovState v036gov.GenesisState
	v015Codec.MustUnmarshalJSON(govJson, &oldGovState)

	// Migrate x/gov from v036 to v040
	v40govState := MigrateGovV036(oldGovState)

	// Migrate x/gov from v040 to v043
	// Note: x/gov v043 just migrates the votes property, which should not matter
	// to us since we don't have any votes data on mainnet for x/gov.
	return v016Codec.MustMarshalJSON(v043gov.MigrateJSON(v40govState))
}
