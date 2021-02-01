package v0_11

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v39_1auth "github.com/cosmos/cosmos-sdk/x/auth"
	v39_1auth_vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	v39_1supply "github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/app"
	v38_5auth "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/auth"
	v38_5supply "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/supply"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_9cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_9"
	v0_9committee "github.com/kava-labs/kava/x/committee/legacy/v0_9"
	v0_9incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_9"
	v0_9pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_9"
	v0_11validator_vesting "github.com/kava-labs/kava/x/validator-vesting"
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

func TestMigrateCommittee(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "committee-v09.json"))
	require.NoError(t, err)
	var oldGenState v0_9committee.GenesisState
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	v0_9committee.RegisterCodec(cdc)

	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateCommittee(oldGenState)
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
	require.Equal(t, len(oldGenState.Accounts)+5, len(newGenState.Accounts))
}

func TestMigrateAuthExact(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-v09-simplified.json"))
	require.NoError(t, err)

	var oldGenState v38_5auth.GenesisState

	v09_cdc := codec.New()
	codec.RegisterCrypto(v09_cdc)
	v38_5auth.RegisterCodec(v09_cdc)
	v38_5auth.RegisterCodecVesting(v09_cdc)
	v38_5supply.RegisterCodec(v09_cdc)
	v0_9validator_vesting.RegisterCodec(v09_cdc)

	require.NotPanics(t, func() {
		v09_cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateAuth(oldGenState)

	v011_cdc := codec.New()
	codec.RegisterCrypto(v011_cdc)
	v39_1auth.RegisterCodec(v011_cdc)
	v39_1auth_vesting.RegisterCodec(v011_cdc)
	v39_1supply.RegisterCodec(v011_cdc)
	v0_11validator_vesting.RegisterCodec(v011_cdc)

	newGenStateBz, err := v011_cdc.MarshalJSON(newGenState)
	require.NoError(t, err)

	expectedGenStateBz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-v011-simplified.json"))
	require.NoError(t, err)

	require.JSONEq(t, string(expectedGenStateBz), string(newGenStateBz))

}

func TestMigrateHarvest(t *testing.T) {
	newGenState := MigrateHarvest()
	err := newGenState.Validate()
	require.NoError(t, err)
}
func TestMigrateCdp(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "cdp-v09.json"))
	require.NoError(t, err)
	var oldGenState v0_9cdp.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateCDP(oldGenState)
	err = newGenState.Validate()
	require.NoError(t, err)
}
func TestMigrateIncentive(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "incentive-v09.json"))
	require.NoError(t, err)
	var oldGenState v0_9incentive.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateIncentive(oldGenState)
	err = newGenState.Validate()
	require.NoError(t, err)
}
func TestMigratePricefeed(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "pricefeed-v09.json"))
	require.NoError(t, err)
	var oldGenState v0_9pricefeed.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})
	newGenState := MigratePricefeed(oldGenState)
	err = newGenState.Validate()
	require.NoError(t, err)
}
