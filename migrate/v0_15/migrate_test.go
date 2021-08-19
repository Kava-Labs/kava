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
	v0_15cdp "github.com/kava-labs/kava/x/cdp/types"
	v0_14committee "github.com/kava-labs/kava/x/committee/legacy/v0_14"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
	"github.com/kava-labs/kava/x/hard"
	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/swap"
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

	var oldCDPGenState v0_15cdp.GenesisState
	cdc.MustUnmarshalJSON(oldState[v0_15cdp.ModuleName], &oldCDPGenState)

	newGenState := Incentive(app.MakeCodec(), oldIncentiveGenState, oldCDPGenState.CDPs)
	require.NoError(t, newGenState.Validate())

	// TODO check params, indexes, and accumulation times

	// Ensure the usdx claim indexes match global
	globalIndexes := newGenState.USDXRewardState.MultiRewardIndexes
	for _, claim := range newGenState.USDXMintingClaims {

		for _, globalIndex := range globalIndexes {
			expectedFactor, found := globalIndex.RewardIndexes.Get(v0_15incentive.USDXMintingRewardDenom)
			require.True(t, found)

			factor, found := claim.RewardIndexes.Get(globalIndex.CollateralType)
			require.True(t, found)

			require.Equal(t, expectedFactor, factor)
		}
	}

	// Ensure there is a usdx claim for every cdp
	for _, cdp := range oldCDPGenState.CDPs {
		numClaims := 0
		for _, claim := range newGenState.USDXMintingClaims {
			if cdp.Owner.Equals(claim.Owner) {
				numClaims++
			}
		}
		require.Equal(t, 1, numClaims, "cdp '%s' has invalid number of claims '%d'", cdp.Owner, numClaims)

		// also check cdp indexes are valid
		require.True(t, cdp.InterestFactor.GTE(sdk.OneDec()), "found cdp with interest factor < 1")
	}

	// Check reward amounts
	for _, claim := range newGenState.USDXMintingClaims {

		// check a few high value accounts
		switch claim.Owner.String() {
		// check reward is: additional reward + existing unclaimed reward
		// note, non zero unclaimed rewards could change if the user submits a claim tx before launch
		case "kava1k8lymw58tduy9gm6jkt04ddkjd83nf7sm8xthl":
			require.Equal(t, sdk.NewInt(370982556999+0), claim.Reward.Amount)
		case "kava1p3ucd3ptpw902fluyjzhq3ffgq4ntddaysyq8h":
			require.Equal(t, sdk.NewInt(77550672285+16960713469), claim.Reward.Amount)
		case "kava1qe6ahdnhnfugle29054d8uqg7fa44ryx934yc6":
			require.Equal(t, sdk.NewInt(40874651319+0), claim.Reward.Amount)
		case "kava12h6pq2xqzgtxttrzg7q2rplsyxtv2dc5gwh8rl":
			require.Equal(t, sdk.NewInt(30867752254+0), claim.Reward.Amount)
		case "kava10hczxv0p3eadcwgt5u79yhahsyuw98u26qan50":
			require.Equal(t, sdk.NewInt(22429344254+0), claim.Reward.Amount)
		case "kava15wyjwhj6zh79m7adm69pwl3nsq9z8gs9ezs4k7":
			require.Equal(t, sdk.NewInt(10252596901+0), claim.Reward.Amount)
		case "kava1yg4840l77dfs5zqflldhut27en2mhvvc8vj93x":
			require.Equal(t, sdk.NewInt(9898765520+0), claim.Reward.Amount)
		case "kava1x242qk6jf2rv23ruvk6fmxp97gg2y75a9r2caq":
			require.Equal(t, sdk.NewInt(7761701231+0), claim.Reward.Amount)
		case "kava1tstf3u4cw7u4xyu7wxdrnmrpvvmfamq3twcj7f":
			require.Equal(t, sdk.NewInt(2466900572+0), claim.Reward.Amount)
		}

		// check no rewards have been reduced
		for _, oldClaim := range oldIncentiveGenState.USDXMintingClaims {
			if oldClaim.Owner.Equals(claim.Owner) {
				require.Truef(t, claim.Reward.IsGTE(oldClaim.Reward), "found claim with reduced rewards, old %s, new %s", oldClaim, claim)
			}
		}
	}

	require.Equal(t, len(oldIncentiveGenState.HardLiquidityProviderClaims), len(newGenState.HardLiquidityProviderClaims))
	// 1 new DelegatorClaim should have been created for each existing HardLiquidityProviderClaim
	require.Equal(t, len(oldIncentiveGenState.HardLiquidityProviderClaims), len(newGenState.DelegatorClaims))
}

func TestIncentive_Full_TotalRewards(t *testing.T) {
	t.Skip() // skip to avoid having to commit a large genesis file to the repo

	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis.json"))
	require.NoError(t, err)

	cdc := makeV014Codec()

	var oldState genutil.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &oldState)

	var oldIncentiveGenState v0_14incentive.GenesisState
	cdc.MustUnmarshalJSON(oldState[v0_14incentive.ModuleName], &oldIncentiveGenState)

	var oldCDPGenState v0_15cdp.GenesisState
	cdc.MustUnmarshalJSON(oldState[v0_15cdp.ModuleName], &oldCDPGenState)

	newGenState := Incentive(app.MakeCodec(), oldIncentiveGenState, oldCDPGenState.CDPs)

	// total previous rewards
	oldTotalRewards := sdk.NewCoins() // total synced unclaimed rewards
	for _, claim := range oldIncentiveGenState.HardLiquidityProviderClaims {
		oldTotalRewards = oldTotalRewards.Add(claim.Reward...)
	}
	for _, claim := range oldIncentiveGenState.USDXMintingClaims {
		oldTotalRewards = oldTotalRewards.Add(claim.Reward)
	}

	// total new rewards
	newTotalRewards := sdk.NewCoins() // total synced unclaimed rewards
	for _, claim := range newGenState.USDXMintingClaims {
		newTotalRewards = newTotalRewards.Add(claim.Reward)
	}
	for _, claim := range newGenState.HardLiquidityProviderClaims {
		newTotalRewards = newTotalRewards.Add(claim.Reward...)
	}

	// rewards added in migration
	additionalRewards := sdk.NewCoins()
	var missedRewards map[string]sdk.Coin
	cdc.MustUnmarshalJSON([]byte(missedUSDXMintingRewards), &missedRewards)
	for _, c := range missedRewards {
		additionalRewards = additionalRewards.Add(c)
	}

	require.Equal(t, oldTotalRewards.Add(additionalRewards...), newTotalRewards)
}

func TestIncentive_SwpRewards(t *testing.T) {
	expectedAnnualRewards := sdk.NewCoin("swp", sdk.NewInt(24000000e6))
	expectedAnnualRewardsMap := map[string]sdk.Coin{
		"ukava:usdx": sdk.NewCoin("swp", sdk.NewInt(7000000e6)),
		"swp:usdx":   sdk.NewCoin("swp", sdk.NewInt(6000000e6)),
		"busd:usdx":  sdk.NewCoin("swp", sdk.NewInt(5000000e6)),
		"bnb:usdx":   sdk.NewCoin("swp", sdk.NewInt(1500000e6)),
		"btcb:usdx":  sdk.NewCoin("swp", sdk.NewInt(1500000e6)),
		"hard:usdx":  sdk.NewCoin("swp", sdk.NewInt(1500000e6)),
		"usdx:xrpb":  sdk.NewCoin("swp", sdk.NewInt(1500000e6)),
	}

	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-test-incentive-state.json"))
	require.NoError(t, err)
	var oldIncentiveGenState v0_14incentive.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldIncentiveGenState)
	})
	secondsPerYear := sdk.NewInt(31536000)

	newGenState := v0_15incentive.GenesisState{}
	require.NotPanics(t, func() {
		newGenState = Incentive(app.MakeCodec(), oldIncentiveGenState, v0_15cdp.CDPs{})
	})
	err = newGenState.Validate()
	require.NoError(t, err)

	actualAnnualRewards := sdk.NewCoin("swp", sdk.ZeroInt())

	for _, rp := range newGenState.Params.SwapRewardPeriods {
		expected, found := expectedAnnualRewardsMap[rp.CollateralType]
		require.True(t, found)
		require.True(t, len(rp.RewardsPerSecond) == 1)
		require.True(t, rp.RewardsPerSecond[0].Denom == "swp")
		annualRewardsAmount := rp.RewardsPerSecond[0].Amount.ToDec().Mul(secondsPerYear.ToDec()).RoundInt()
		annualRewardsCoin := sdk.NewCoin("swp", annualRewardsAmount)
		actualAnnualRewards = actualAnnualRewards.Add(annualRewardsCoin)
		// allow for +- 0.1% deviation for targeted annual rewards
		inRange := assertRewardsWithinRange(expected, annualRewardsCoin, sdk.MustNewDecFromStr("0.001"))
		require.True(t, inRange, fmt.Sprintf("expected annual rewards: %s, actual %s", expected, annualRewardsCoin))
	}
	inRange := assertRewardsWithinRange(expectedAnnualRewards, actualAnnualRewards, sdk.MustNewDecFromStr("0.001"))
	require.True(t, inRange, fmt.Sprintf("expected annual rewards: %s, actual %s", expectedAnnualRewards, actualAnnualRewards))
}

func TestIncentive_SwpPoolsValid(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-7-test-incentive-state.json"))
	require.NoError(t, err)
	appState := genutil.AppMap{v0_14incentive.ModuleName: bz, v0_15cdp.ModuleName: app.MakeCodec().MustMarshalJSON(v0_15cdp.DefaultGenesisState())}

	MigrateAppState(appState)
	cdc := app.MakeCodec()
	var swapGS swap.GenesisState
	cdc.MustUnmarshalJSON(appState[swap.ModuleName], &swapGS)
	var incentiveGS v0_15incentive.GenesisState
	cdc.MustUnmarshalJSON(appState[v0_15incentive.ModuleName], &incentiveGS)

	for _, rp := range incentiveGS.Params.SwapRewardPeriods {
		found := false
		for _, pool := range swapGS.Params.AllowedPools {
			if pool.Name() == rp.CollateralType {
				found = true
			}
		}
		require.True(t, found, fmt.Sprintf("incentivized pool %s not found in swap genesis state", rp.CollateralType))
	}

}

func assertRewardsWithinRange(expected, actual sdk.Coin, tolerance sdk.Dec) bool {
	upper := expected.Amount.ToDec().Mul(sdk.OneDec().Add(tolerance))
	lower := expected.Amount.ToDec().Mul(sdk.OneDec().Sub(tolerance))
	if actual.Amount.ToDec().GTE(lower) && actual.Amount.ToDec().LTE(upper) {
		return true
	}
	return false
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

func TestAuth_Regression_BalancesEqual(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "block-1543671-auth-state.json"))
	require.NoError(t, err)

	cdc := app.MakeCodec()

	var originalGenesisState auth.GenesisState
	cdc.MustUnmarshalJSON(bz, &originalGenesisState)

	var genesisState auth.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesisState)

	genesisState = Auth(cdc, genesisState, GenesisTime)

	for i, oldAcc := range originalGenesisState.Accounts {
		acc := genesisState.Accounts[i]

		// total owned coins does not change, despite additional swp tokens
		swpBalance := sdk.NewCoins(sdk.NewCoin("swp", acc.GetCoins().AmountOf("swp")))
		require.Equal(t, oldAcc.GetCoins(), acc.GetCoins().Sub(swpBalance), "expected base coins to not change")

		// ensure spendable coins at genesis time is equal, despite additional swp tokens
		swpBalance = sdk.NewCoins(sdk.NewCoin("swp", acc.SpendableCoins(GenesisTime).AmountOf("swp")))
		require.Equal(t, oldAcc.SpendableCoins(GenesisTime), acc.SpendableCoins(GenesisTime).Sub(swpBalance), "expected spendable coins to not change")
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

func TestAuth_SwpSupply_AfterAirdrop(t *testing.T) {
	swpSupply := sdk.NewCoin("swp", sdk.ZeroInt())
	expectedSwpSupply := sdk.NewCoin("swp", sdk.NewInt(1000000000000))
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "block-1543671-auth-state.json"))
	require.NoError(t, err)

	var genesisState auth.GenesisState
	cdc := app.MakeCodec()
	cdc.MustUnmarshalJSON(bz, &genesisState)

	migratedGenesisState := ApplySwpAirdrop(cdc, genesisState)

	for _, acc := range migratedGenesisState.Accounts {
		swpAmount := acc.GetCoins().AmountOf("swp")
		if swpAmount.IsPositive() {
			swpCoin := sdk.NewCoin("swp", swpAmount)
			swpSupply = swpSupply.Add(swpCoin)
		}
	}
	require.Equal(t, expectedSwpSupply, swpSupply)
}

func TestAuth_SwpSupply_TotalSupply(t *testing.T) {
	swpSupply := sdk.NewCoin("swp", sdk.ZeroInt())
	expectedSwpSupply := sdk.NewCoin("swp", sdk.NewInt(250000000e6))
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

func TestAuth_SwpSupply_SpendableCoins(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "block-1543671-auth-state.json"))
	require.NoError(t, err)

	var genesisState auth.GenesisState
	cdc := app.MakeCodec()
	cdc.MustUnmarshalJSON(bz, &genesisState)

	swpTreasuryExpectedSpendableCoins := sdk.NewCoins(SwpTreasuryCoins.Sub(SwpTreasuryOriginalVesting))
	foundTreasury := false

	swpTeamExpectedSpendableCoins := sdk.Coins(nil)
	foundTeam := false

	kavaDistExpectedSpendableCoins := KavaDistCoins
	foundKavadist := false

	swpEcosystemExpectedSpendableCoins := sdk.NewCoins(EcoSystemCoins)
	foundEcosystem := false

	migratedGenesisState := Auth(cdc, genesisState, GenesisTime)

	for _, acc := range migratedGenesisState.Accounts {
		if acc.GetAddress().Equals(SwpTreasuryAddr) {
			foundTreasury = true
			pva, ok := acc.(*vesting.PeriodicVestingAccount)
			require.True(t, ok)
			spendableCoins := pva.SpendableCoins(GenesisTime)
			require.Equal(t, swpTreasuryExpectedSpendableCoins, spendableCoins)
		}
		if acc.GetAddress().Equals(SwpTeamAddr) {
			foundTeam = true
			pva, ok := acc.(*vesting.PeriodicVestingAccount)
			require.True(t, ok)
			spendableCoins := pva.SpendableCoins(GenesisTime)
			require.Equal(t, swpTeamExpectedSpendableCoins, spendableCoins)
		}
		if acc.GetAddress().Equals(KavaDistAddr) {
			foundKavadist = true
			spendableCoins := acc.SpendableCoins(GenesisTime)
			spendableSwp := sdk.NewCoin("swp", spendableCoins.AmountOf("swp"))
			require.Equal(t, kavaDistExpectedSpendableCoins, spendableSwp)
		}
		if acc.GetAddress().Equals(SwpEcoSystemAddr) {
			foundEcosystem = true
			spendableCoins := acc.SpendableCoins(GenesisTime)
			require.Equal(t, swpEcosystemExpectedSpendableCoins, spendableCoins)
		}
	}
	require.True(t, foundTreasury)
	require.True(t, foundTeam)
	require.True(t, foundKavadist)
	require.True(t, foundEcosystem)

}
