package v0_15

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	v0_14committee "github.com/kava-labs/kava/x/committee/legacy/v0_14"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)

	os.Exit(m.Run())
}

func TestCommittee(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-committee-state.json"))
	require.NoError(t, err)

	var oldGenState v0_14committee.GenesisState
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	v0_14committee.RegisterCodec(cdc)

	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := Committee(oldGenState)
	err = newGenState.Validate()
	require.NoError(t, err)

	require.Equal(t, len(oldGenState.Committees), len(newGenState.Committees))
	for i := 0; i < len(oldGenState.Committees); i++ {
		require.Equal(t, len(oldGenState.Committees[i].Permissions), len(newGenState.Committees[i].GetPermissions()))
	}

	oldSPCP := oldGenState.Committees[0].Permissions[0].(v0_14committee.SubParamChangePermission)
	newSPCP := newGenState.Committees[0].GetPermissions()[0].(v0_15committee.SubParamChangePermission)
	require.Equal(t, len(oldSPCP.AllowedParams), len(newSPCP.AllowedParams))
	require.Equal(t, len(oldSPCP.AllowedAssetParams), len(newSPCP.AllowedAssetParams))
	require.Equal(t, len(oldSPCP.AllowedCollateralParams), len(newSPCP.AllowedCollateralParams))
	require.Equal(t, len(oldSPCP.AllowedMarkets), len(newSPCP.AllowedMarkets))
	require.Equal(t, len(oldSPCP.AllowedMoneyMarkets), len(newSPCP.AllowedMoneyMarkets))
}

func TestIncentive(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-incentive-state.json"))
	require.NoError(t, err)
	var oldIncentiveGenState v0_14incentive.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldIncentiveGenState)
	})

	newGenState := v0_15incentive.GenesisState{}
	require.NotPanics(t, func() {
		newGenState = Incentive(oldIncentiveGenState)
	})
	err = newGenState.Validate()
	require.NoError(t, err)

	require.Equal(t, len(oldIncentiveGenState.USDXMintingClaims), len(newGenState.USDXMintingClaims))
	require.Equal(t, len(oldIncentiveGenState.HardLiquidityProviderClaims), len(newGenState.HardLiquidityProviderClaims))
}
