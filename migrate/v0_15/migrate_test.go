package v0_15

import (
	"fmt"
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
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	v0_14committee "github.com/kava-labs/kava/x/committee/legacy/v0_14"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
	v0_15hard "github.com/kava-labs/kava/x/hard/types"
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

func TestIncentive_Full(t *testing.T) {
	t.Skip() // skip to avoid having to commit a large genesis file to the repo

	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis.json"))
	require.NoError(t, err)

	cdc := makeV014Codec()

	var oldState genutil.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &oldState)

	var oldIncentiveGenState v0_14incentive.GenesisState
	cdc.MustUnmarshalJSON(oldState[v0_14incentive.ModuleName], &oldIncentiveGenState)

	var oldHardGenState v0_15hard.GenesisState
	cdc.MustUnmarshalJSON(oldState[v0_15hard.ModuleName], &oldHardGenState)

	newGenState := Incentive(oldIncentiveGenState, oldHardGenState)
	require.NoError(t, newGenState.Validate())

	// TODO check params, indexes, and accumulation times

	// ensure every hard deposit has a claim with correct indexes
	for _, deposit := range oldHardGenState.Deposits {
		foundClaim := false
		for _, claim := range newGenState.HardLiquidityProviderClaims {
			if claim.Owner.Equals(deposit.Depositor) {
				foundClaim = true
				// check indexes are valid
				err := collateralTypesMatch(deposit.Amount, claim.SupplyRewardIndexes)
				require.NoErrorf(t, err, "invalid claim %s, %s", claim, deposit)
				break
			}
		}
		require.True(t, foundClaim, "missing claim for hard deposit")

		// also ensure hard deposit indexes are valid
		for _, i := range deposit.Index {
			require.Truef(t, i.Value.GTE(sdk.OneDec()), "found invalid hard deposit index %s", deposit)
		}
	}

	// ensure every hard borrow has a claim with correct indexes
	for _, borrow := range oldHardGenState.Borrows {
		foundClaim := false
		for _, claim := range newGenState.HardLiquidityProviderClaims {
			if claim.Owner.Equals(borrow.Borrower) {
				foundClaim = true
				// check indexes are valid
				err := collateralTypesMatch(borrow.Amount, claim.BorrowRewardIndexes)
				require.NoErrorf(t, err, "invalid claim %s, %s", claim, borrow)
				break
			}
		}
		require.True(t, foundClaim, "missing claim for hard borrow")

		// also ensure hard borrow indexes are valid
		for _, i := range borrow.Index {
			require.Truef(t, i.Value.GTE(sdk.OneDec()), "found invalid hard borrow index %s", borrow)
		}
	}

	// ensure all claim indexes are ≤ global values
	for _, claim := range newGenState.HardLiquidityProviderClaims {

		for _, ri := range claim.BorrowRewardIndexes {
			global, found := newGenState.HardBorrowRewardState.MultiRewardIndexes.Get(ri.CollateralType)
			if !found {
				global = v0_15incentive.RewardIndexes{}
			}
			require.Truef(t, indexesAllLessThanOrEqual(ri.RewardIndexes, global), "invalid claim supply indexes %s %s", ri.RewardIndexes, global)
		}

		for _, ri := range claim.SupplyRewardIndexes {
			global, found := newGenState.HardSupplyRewardState.MultiRewardIndexes.Get(ri.CollateralType)
			if !found {
				global = v0_15incentive.RewardIndexes{}
			}
			require.Truef(t, indexesAllLessThanOrEqual(ri.RewardIndexes, global), "invalid claim borrow indexes %s %s", ri.RewardIndexes, global)
		}
	}

	// ensure (synced) reward amounts are unchanged
	for _, claim := range newGenState.HardLiquidityProviderClaims {
		for _, oldClaim := range oldIncentiveGenState.HardLiquidityProviderClaims {
			if oldClaim.Owner.Equals(claim.Owner) {
				require.Equal(t, claim.Reward, oldClaim.Reward)
			}
		}
	}

	require.Equal(t, len(oldIncentiveGenState.USDXMintingClaims), len(newGenState.USDXMintingClaims))

	// 1 new DelegatorClaim should have been created for each existing HardLiquidityProviderClaim
	require.Equal(t, len(oldIncentiveGenState.HardLiquidityProviderClaims), len(newGenState.DelegatorClaims))
}

// collateralTypesMatch checks if the set of coin denoms is equal to the set of CollateralTypes in the indexes.
func collateralTypesMatch(coins sdk.Coins, indexes v0_15incentive.MultiRewardIndexes) error {
	for _, index := range indexes {
		if coins.AmountOf(index.CollateralType).Equal(sdk.ZeroInt()) {
			return fmt.Errorf("index contains denom not found in coins")
		}
	}
	for _, coin := range coins {
		_, found := indexes.Get(coin.Denom)
		if !found {
			return fmt.Errorf("coins contain denom not found in indexes")
		}
	}
	return nil
}

// indexesAllLessThanOrEqual computes if all factors in A are ≤ factors in B.
// Missing indexes are taken to be zero.
func indexesAllLessThanOrEqual(indexesA, indexesB v0_15incentive.RewardIndexes) bool {
	allLT := true
	for _, ri := range indexesA {
		factor, found := indexesB.Get(ri.CollateralType)
		if !found {
			// value not found is same as it being zero
			factor = sdk.ZeroDec()
		}
		allLT = allLT && ri.RewardFactor.LTE(factor)
	}
	return allLT
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
