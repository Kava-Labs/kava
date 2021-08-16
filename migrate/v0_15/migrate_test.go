package v0_15

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/stretchr/testify/assert"
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

	require.Equal(t, len(oldGenState.Committees)+2, len(newGenState.Committees)) // New gen state has 2 additional committees
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

// exportGenesisJSON is a utility testing method
func exportGenesisJSON(genState v0_15committee.GenesisState) {
	v15Cdc := app.MakeCodec()
	ioutil.WriteFile(filepath.Join("testdata", "kava-8-committee-state.json"), v15Cdc.MustMarshalJSON(genState), 0644)
}

func TestIncentive_MainnetState(t *testing.T) {
	// TODO add copy of mainnet state to json
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
	// 1 new DelegatorClaim should have been created for each existing HardLiquidityProviderClaim
	require.Equal(t, len(oldIncentiveGenState.HardLiquidityProviderClaims), len(newGenState.DelegatorClaims))
}

func TestIncentive(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "v0_14-incentive-state.json"))
	require.NoError(t, err)
	appState := genutil.AppMap{v0_14incentive.ModuleName: bz}

	MigrateAppState(appState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "v0_15-incentive-state.json"))
	require.NoError(t, err)

	require.JSONEq(t, string(bz), string(appState[v0_15incentive.ModuleName]))
}

func TestSwap(t *testing.T) {
	swapGS := Swap()
	err := swapGS.Validate()
	require.NoError(t, err)
	require.Equal(t, 7, len(swapGS.Params.AllowedPools))
	require.Equal(t, 0, len(swapGS.PoolRecords))
	require.Equal(t, 0, len(swapGS.ShareRecords))
}

// Compare migration against auto-generated snapshot to catch regressions
func TestAuth_Snapshot(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-test-auth-state.json"))
	require.NoError(t, err)
	appState := genutil.AppMap{auth.ModuleName: bz}

	MigrateAppState(appState)

	if _, err := os.Stat(filepath.Join("testdata", "kava-8-test-auth-state.json")); os.IsNotExist(err) {
		err := ioutil.WriteFile(filepath.Join("testdata", "kava-8-test-auth-state.json"), appState[auth.ModuleName], 0644)
		require.NoError(t, err)
	}

	snapshot, err := ioutil.ReadFile(filepath.Join("testdata", "kava-8-test-auth-state.json"))
	require.NoError(t, err)

	assert.JSONEq(t, string(snapshot), string(appState[auth.ModuleName]), "expected auth state snapshot to be equal")
}

func TestAuth_ParametersEqual(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-test-auth-state.json"))
	require.NoError(t, err)

	var genesisState auth.GenesisState
	cdc := app.MakeCodec()
	cdc.MustUnmarshalJSON(bz, &genesisState)

	migratedGenesisState := Auth(genesisState, GenesisTime)

	assert.Equal(t, genesisState.Params, migratedGenesisState.Params, "expected auth parameters to not change")
}

func TestAuth_AccountConversion(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-test-auth-state.json"))
	require.NoError(t, err)

	cdc := app.MakeCodec()

	var genesisState auth.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesisState)

	var originalGenesisState auth.GenesisState
	cdc.MustUnmarshalJSON(bz, &originalGenesisState)

	migratedGenesisState := Auth(genesisState, GenesisTime)
	require.Equal(t, len(genesisState.Accounts), len(migratedGenesisState.Accounts), "expected the number of accounts after migration to be equal")
	err = auth.ValidateGenesis(migratedGenesisState)
	require.NoError(t, err, "expected migrated genesis to be valid")

	for i, acc := range migratedGenesisState.Accounts {
		oldAcc := originalGenesisState.Accounts[i]

		// total owned coins does not change
		require.Equal(t, oldAcc.GetCoins(), acc.GetCoins(), "expected base coins to not change")

		// ensure spenable coins at genesis time is equal
		require.Equal(t, oldAcc.SpendableCoins(GenesisTime), acc.SpendableCoins(GenesisTime), "expected spendable coins to not change")
		// check 30 days
		futureDate := GenesisTime.Add(30 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), acc.SpendableCoins(futureDate), "expected spendable coins to not change")
		// check 90 days
		futureDate = GenesisTime.Add(90 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), acc.SpendableCoins(futureDate), "expected spendable coins to not change")
		// check 180 days
		futureDate = GenesisTime.Add(180 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), acc.SpendableCoins(futureDate), "expected spendable coins to not change")
		// check 365 days
		futureDate = GenesisTime.Add(365 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), acc.SpendableCoins(futureDate), "expected spendable coins to not change")
		// check 2 years
		futureDate = GenesisTime.Add(2 * 365 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), acc.SpendableCoins(futureDate), "expected spendable coins to not change")

		if vacc, ok := acc.(*vesting.PeriodicVestingAccount); ok {
			// old account must be a periodic vesting account
			oldVacc, ok := oldAcc.(*vesting.PeriodicVestingAccount)
			require.True(t, ok)

			// total delegated coins must match
			oldTotalDelegated := oldVacc.DelegatedFree.Add(oldVacc.DelegatedVesting...)
			newTotalDelegated := vacc.DelegatedFree.Add(vacc.DelegatedVesting...)
			require.Equal(t, oldTotalDelegated, newTotalDelegated, "expected total amount of tracked delegations to not change")

			// delegated vesting must be less or equal to original vesting
			require.True(t, vacc.DelegatedVesting.IsAllLTE(vacc.OriginalVesting), "expected delegated vesting to be less or equal to original vesting")

			// vested coins must be nil for the new account
			require.Equal(t, sdk.Coins(nil), vacc.GetVestedCoins(GenesisTime), "expected no vested coins at genesis time")

			// vesting coins must not be nil
			require.NotEqual(t, sdk.Coins(nil), vacc.GetVestingCoins(GenesisTime), "expected vesting coins to be greater than 0")

			// new account as less than or equal
			require.LessOrEqual(t, len(vacc.VestingPeriods), len(oldVacc.VestingPeriods), "expected vesting periods of new account to be less than or equal to old")

			// end time should not change
			require.Equal(t, oldVacc.EndTime, vacc.EndTime, "expected end time to not change")
		}
	}
}
