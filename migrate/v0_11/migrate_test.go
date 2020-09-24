package v0_11

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v39_1auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/kava-labs/kava/app"
	v38_5auth "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/auth"
	v38_5supply "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/supply"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_9validator_vesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_9"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)

	os.Exit(m.Run())
}

func TestMigrateBep3(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "bep3-v09.json"))
	require.NoError(t, err)
	var oldGenState v0_9bep3.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateBep3(oldGenState)
	err = newGenState.Validate()
	require.NoError(t, err)
}

func TestMigrateAuth(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-v09.json"))
	require.NoError(t, err)
	var oldGenState v38_5auth.GenesisState
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	v38_5auth.RegisterCodec(cdc)
	v38_5auth.RegisterCodecVesting(cdc)
	v38_5supply.RegisterCodec(cdc)
	v0_9validator_vesting.RegisterCodec(cdc)

	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateAuth(oldGenState)
	err = v39_1auth.ValidateGenesis(newGenState)
	require.NoError(t, err)
	require.Equal(t, len(oldGenState.Accounts)+2, len(newGenState.Accounts))
}
