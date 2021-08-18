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
	v0_15staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	v0_14committee "github.com/kava-labs/kava/x/committee/legacy/v0_14"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
	"github.com/kava-labs/kava/x/hard"
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

func TestIncentive_Full(t *testing.T) {
	t.Skip() // skip to avoid having to commit a large genesis file to the repo

	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis.json"))
	require.NoError(t, err)

	cdc := makeV014Codec()

	var oldState genutil.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &oldState)

	var oldIncentiveGenState v0_14incentive.GenesisState

	cdc.MustUnmarshalJSON(oldState[v0_14incentive.ModuleName], &oldIncentiveGenState)

	var oldStakingGenState v0_15staking.GenesisState
	cdc.MustUnmarshalJSON(oldState[v0_15staking.ModuleName], &oldStakingGenState)

	newGenState := Incentive(oldIncentiveGenState, oldStakingGenState.Delegations)
	require.NoError(t, newGenState.Validate())

	// TODO check params, indexes, and accumulation times

	// Ensure every delegation has a claim
	for _, delegation := range oldStakingGenState.Delegations {
		numClaims := 0
		for _, claim := range newGenState.DelegatorClaims {
			if delegation.DelegatorAddress.Equals(claim.Owner) {
				numClaims++
			}
		}
		require.Equal(t, 1, numClaims, "delegation '%s' has invalid number of claims '%d'", delegation.DelegatorAddress, numClaims)
	}

	// Ensure all delegation indexes are ≤ global values
	for _, claim := range newGenState.DelegatorClaims {
		for _, ri := range claim.RewardIndexes {
			require.Equal(t, v0_15incentive.BondDenom, ri.CollateralType)

			global, found := newGenState.DelegatorRewardState.MultiRewardIndexes.Get(ri.CollateralType)
			if !found {
				global = v0_15incentive.RewardIndexes{}
			}
			require.Truef(t, indexesAllLessThanOrEqual(ri.RewardIndexes, global), "invalid delegator claim indexes %s %s", ri.RewardIndexes, global)
		}
	}

	// Ensure delegator rewards are all initialized at 0
	for _, claim := range newGenState.DelegatorClaims {
		require.Equal(t, sdk.NewCoins(), claim.Reward)
	}
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

func TestAuth_ParametersEqual(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-test-auth-state.json"))
	require.NoError(t, err)

	var genesisState auth.GenesisState
	cdc := app.MakeCodec()
	cdc.MustUnmarshalJSON(bz, &genesisState)

	migratedGenesisState := Auth(cdc, genesisState, GenesisTime)

	assert.Equal(t, genesisState.Params, migratedGenesisState.Params, "expected auth parameters to not change")
}

func TestAuth_AccountConversion(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-test-auth-state.json"))
	require.NoError(t, err)

	cdc := app.MakeCodec()

	var genesisState auth.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesisState)

	migratedGenesisState := MigrateAccounts(genesisState, GenesisTime)
	var originalGenesisState auth.GenesisState
	cdc.MustUnmarshalJSON(bz, &originalGenesisState)
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

func TestAuth_MakeAirdropMap(t *testing.T) {
	cdc := app.MakeCodec()
	aidropTokenAmount := sdk.NewInt(1000000000000)
	totalSwpTokens := sdk.ZeroInt()
	var loadedAirdropMap map[string]sdk.Coin
	cdc.MustUnmarshalJSON([]byte(swpAirdropMap), &loadedAirdropMap)
	for _, coin := range loadedAirdropMap {
		totalSwpTokens = totalSwpTokens.Add(coin.Amount)
	}
	require.Equal(t, aidropTokenAmount, totalSwpTokens)
}

func TestAuth_TestAllDepositorsIncluded(t *testing.T) {
	var deposits hard.Deposits
	cdc := app.MakeCodec()
	bz, err := ioutil.ReadFile("./data/hard-deposits-block-1543671.json")
	if err != nil {
		panic(fmt.Sprintf("Couldn't open hard deposit snapshot file: %v", err))
	}
	cdc.MustUnmarshalJSON(bz, &deposits)

	depositorsInSnapShot := 0
	for _, dep := range deposits {
		if dep.Amount.AmountOf("usdx").IsPositive() {
			depositorsInSnapShot++
		}
	}
	var loadedAirdropMap map[string]sdk.Coin
	cdc.MustUnmarshalJSON([]byte(swpAirdropMap), &loadedAirdropMap)
	keys := make([]string, 0, len(loadedAirdropMap))
	for k := range loadedAirdropMap {
		keys = append(keys, k)
	}
	require.Equal(t, depositorsInSnapShot, len(keys))
}

func TestAuth_SwpSupply(t *testing.T) {
	swpSupply := sdk.NewCoin("swp", sdk.ZeroInt())
	// TODO update when additional swp are added to migration, final supply should be 250M at genesis
	expectedSwpSupply := sdk.NewCoin("swp", sdk.NewInt(1000000000000))
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "block-1543671-auth-state.json"))
	require.NoError(t, err)

	var genesisState auth.GenesisState
	cdc := app.MakeCodec()
	cdc.MustUnmarshalJSON(bz, &genesisState)

	migratedGenesisState := Auth(cdc, genesisState, GenesisTime)

	for _, acc := range migratedGenesisState.Accounts {
		swpAmount := acc.GetCoins().AmountOf("swp")
		if swpAmount.IsPositive() {
			swpCoin := sdk.NewCoin("swp", swpAmount)
			swpSupply = swpSupply.Add(swpCoin)
		}
	}
	require.Equal(t, expectedSwpSupply, swpSupply)
}
