package v0_13

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/kava-labs/kava/app"
	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
	v0_11hard "github.com/kava-labs/kava/x/hard/legacy/v0_11"
	v0_13incentive "github.com/kava-labs/kava/x/incentive"
	v0_11incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_11"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"

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
	validatorVestingChangeAddress, err := sdk.AccAddressFromBech32("kava1a3qmze57knfj29a5knqs5ptewh76v4fg23xsvn")
	if err != nil {
		panic(err)
	}
	validatorVestingUpdatedValAddress, err := sdk.ConsAddressFromBech32("kavavalcons1ucxhn6zh7y2zun49m36psjffrhmux7ukqxdcte")
	if err != nil {
		panic(err)
	}
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-4-auth-state-block-500000.json"))
	require.NoError(t, err)
	var oldGenState auth.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})
	newGenState := MigrateAuth(oldGenState)
	for _, acc := range newGenState.Accounts {
		if acc.GetAddress().Equals(validatorVestingChangeAddress) {
			vacc := acc.(*validatorvesting.ValidatorVestingAccount)
			require.Equal(t, int64(0), vacc.CurrentPeriodProgress.MissedBlocks)
			require.Equal(t, validatorVestingUpdatedValAddress, vacc.ValidatorAddress)
		}
	}
	err = auth.ValidateGenesis(newGenState)
	require.NoError(t, err)
	require.Equal(t, len(oldGenState.Accounts), len(newGenState.Accounts)+3)

}

func TestMigrateIncentive(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-4-incentive-state.json"))
	require.NoError(t, err)
	var oldIncentiveGenState v0_11incentive.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldIncentiveGenState)
	})

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "kava-4-harvest-state.json"))
	require.NoError(t, err)
	var oldHarvestGenState v0_11hard.GenesisState
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldHarvestGenState)
	})
	newGenState := v0_13incentive.GenesisState{}
	require.NotPanics(t, func() {
		newGenState = MigrateIncentive(oldHarvestGenState, oldIncentiveGenState)
	})
	err = newGenState.Validate()
	require.NoError(t, err)
	fmt.Printf("Number of incentive claims in kava-4: %d\nNumber of incentive Claims in kava-5: %d\n",
		len(oldIncentiveGenState.Claims), len(newGenState.USDXMintingClaims),
	)
	fmt.Printf("Number of harvest claims in kava-4: %d\nNumber of hard claims in kava-5: %d\n", len(oldHarvestGenState.Claims), len(newGenState.HardLiquidityProviderClaims))
}
