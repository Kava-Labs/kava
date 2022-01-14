package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	v036supply "github.com/cosmos/cosmos-sdk/x/bank/legacy/v036"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v038"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/migrate/v0_16/legacyaccounts"
)

func TestMigrateGenesisDoc(t *testing.T) {
	expected := getTestDataJSON("genesis-v16.json")
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis-v15.json"))
	assert.NoError(t, err)
	actualGenDoc, err := Migrate(genDoc, newClientContext())
	assert.NoError(t, err)
	actualJson, err := tmjson.Marshal(actualGenDoc)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(actualJson))
}

func TestMigrateFull(t *testing.T) {
	t.Skip() // avoid committing mainnet state - test also currently fails due to https://github.com/cosmos/cosmos-sdk/issues/10862. If you apply the patch, it will pass
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "kava-8-block-1627000.json"))
	assert.NoError(t, err)
	newGenDoc, err := Migrate(genDoc, newClientContext())
	assert.NoError(t, err)

	encodingConfig := app.MakeEncodingConfig()
	var appMap genutiltypes.AppMap
	err = tmjson.Unmarshal(newGenDoc.AppState, &appMap)
	assert.NoError(t, err)
	err = app.ModuleBasics.ValidateGenesis(encodingConfig.Marshaler, encodingConfig.TxConfig, appMap)
	assert.NoError(t, err)
	tApp := app.NewTestApp()
	require.NotPanics(t, func() {
		tApp.InitializeFromGenesisStatesWithTimeAndChainID(newGenDoc.GenesisTime, newGenDoc.ChainID, app.GenesisState(appMap))
	})
}

func TestAccountBalances(t *testing.T) {
	t.Skip() // avoid committing test data
	app.SetSDKConfig()

	// load auth state from kava-8 with empty accounts removed (keeps size down)
	authbz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-8-test-auth-state.json"))
	require.NoError(t, err)
	// load bank state from kava-8, required for migration
	bankbz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-8-test-bank-state.json"))
	require.NoError(t, err)
	// load supply state from kava-8, required for migration
	supplybz, err := ioutil.ReadFile(filepath.Join("testdata", "kava-8-test-supply-state.json"))
	require.NoError(t, err)

	// codec to read old auth data
	v039Codec := codec.NewLegacyAmino()
	legacyaccounts.RegisterLegacyAminoCodec(v039Codec)
	// store old auth genesis state for comparison to migrated state
	var prevAuthGenState legacyaccounts.GenesisState
	v039Codec.MustUnmarshalJSON(authbz, &prevAuthGenState)

	// migrate auth state to kava-9, with periodic vesting account changes and bank conversion
	appState := genutiltypes.AppMap{
		v039auth.ModuleName:   authbz,
		v038bank.ModuleName:   bankbz,
		v036supply.ModuleName: supplybz,
	}
	clientCtx := newClientContext()
	appState = MigrateCosmosAppState(appState, clientCtx, GenesisTime)

	// store new auth and bank state
	var migratedAuthGenState authtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appState[authtypes.ModuleName], &migratedAuthGenState)
	var migratedBankGenState banktypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appState[banktypes.ModuleName], &migratedBankGenState)

	// store map of accouauthexportednt coins from the bank module
	migratedAccCoins := make(map[string]sdk.Coins)
	for _, bal := range migratedBankGenState.Balances {
		migratedAccCoins[bal.Address] = bal.Coins
	}

	for i, anyAcc := range migratedAuthGenState.Accounts {
		var acc authtypes.AccountI
		err := clientCtx.InterfaceRegistry.UnpackAny(anyAcc, &acc)
		require.NoError(t, err)

		oldAcc := prevAuthGenState.Accounts[i]
		oldAcc.SpendableCoins(GenesisTime)
		newBalance, ok := migratedAccCoins[acc.GetAddress().String()]
		// all accounts have a corresponding balance
		require.True(t, ok, "expected balance to exist")

		// this is the same account
		require.Equal(t, oldAcc.GetAddress(), acc.GetAddress(), "expected account address to match")
		// the owned coins did not change
		require.Equal(t, oldAcc.GetCoins(), newBalance, "expected base coins to not change")

		// ensure spenable coins at genesis time is equal
		require.Equal(t, oldAcc.SpendableCoins(GenesisTime), getSpendableCoinsForAccount(acc, newBalance, GenesisTime), "expected spendable coins to not change")
		// check 30 days
		futureDate := GenesisTime.Add(30 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), getSpendableCoinsForAccount(acc, newBalance, futureDate), "expected spendable coins to not change")
		// check 90 days
		futureDate = GenesisTime.Add(90 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), getSpendableCoinsForAccount(acc, newBalance, futureDate), "expected spendable coins to not change")
		// check 180 days
		futureDate = GenesisTime.Add(180 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), getSpendableCoinsForAccount(acc, newBalance, futureDate), "expected spendable coins to not change")
		// check 365 days
		futureDate = GenesisTime.Add(365 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), getSpendableCoinsForAccount(acc, newBalance, futureDate), "expected spendable coins to not change")
		// check 2 years
		futureDate = GenesisTime.Add(2 * 365 * 24 * time.Hour)
		require.Equal(t, oldAcc.SpendableCoins(futureDate), getSpendableCoinsForAccount(acc, newBalance, futureDate), "expected spendable coins to not change")

		if vacc, ok := acc.(*vestingtypes.PeriodicVestingAccount); ok {
			// old account must be a periodic vesting account
			oldVacc, ok := oldAcc.(*legacyaccounts.PeriodicVestingAccount)
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

			// start time should equal genesis time
			require.Equal(t, vacc.StartTime, GenesisTime.Unix(), "expected start time to be equal to genesis time")

			// end time should not change
			require.Equal(t, oldVacc.EndTime, vacc.EndTime, "expected end time to not change")

			// end time should be after genesis time
			require.Greater(t, vacc.EndTime, GenesisTime.Unix(), "expected end time to be after genesis time")

			// vesting account must have one or more periods
			require.GreaterOrEqual(t, len(vacc.VestingPeriods), 1, "expected one or more vesting periods")
		}
	}
}

func getSpendableCoinsForAccount(acc authtypes.AccountI, balance sdk.Coins, blockTime time.Time) sdk.Coins {
	if vacc, ok := acc.(vestexported.VestingAccount); ok {
		return balance.Sub(vacc.LockedCoins(blockTime))
	}

	return balance
}
