package v0_13

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)

	os.Exit(m.Run())
}

func TestMigrateCdp(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-4-cdp-state-block-500000.json"))
	require.NoError(t, err)
	var oldGenState v0_11cdp.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateCDP(oldGenState)
	err = newGenState.Validate()
	require.NoError(t, err)

	require.Equal(t, len(newGenState.Params.CollateralParams), len(newGenState.PreviousAccumulationTimes))

	cdp1 := newGenState.CDPs[0]
	require.Equal(t, sdk.OneDec(), cdp1.InterestFactor)

}

func TestMigrateAuth(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-4-auth-state-block-500000.json"))
	require.NoError(t, err)
	var oldGenState auth.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})
	newGenState := MigrateAuth(oldGenState)
	err = auth.ValidateGenesis(newGenState)
	require.NoError(t, err)
	require.Equal(t, len(oldGenState.Accounts), len(newGenState.Accounts)+1)

}

func TestBep3(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-4-bep3-state.json"))
	require.NoError(t, err)
	var oldGenState bep3.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})
	newGenState := Bep3(oldGenState)
	err = newGenState.Validate()
	require.NoError(t, err)

	var oldBNBSupply bep3.AssetSupply
	var newBNBSupply bep3.AssetSupply

	for _, supply := range oldGenState.Supplies {
		if supply.GetDenom() == "bnb" {
			oldBNBSupply = supply
		}
	}

	for _, supply := range newGenState.Supplies {
		if supply.GetDenom() == "bnb" {
			newBNBSupply = supply
		}
	}

	require.Equal(t, oldBNBSupply.CurrentSupply.Sub(sdk.NewCoin("bnb", sdk.NewInt(1000000000000))), newBNBSupply.CurrentSupply)

}
