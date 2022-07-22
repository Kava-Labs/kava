package testutil

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/hard"

	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
	savingskeeper "github.com/kava-labs/kava/x/savings/keeper"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// Suite implements a test suite for the earn module integration tests
type Suite struct {
	suite.Suite
	Keeper        keeper.Keeper
	App           app.TestApp
	Ctx           sdk.Context
	BankKeeper    bankkeeper.Keeper
	AccountKeeper authkeeper.AccountKeeper

	// Strategy Keepers
	HardKeeper    hardkeeper.Keeper
	SavingsKeeper savingskeeper.Keeper
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	// Pricefeed required for withdrawing from hard
	pricefeedGS := pricefeedtypes.GenesisState{
		Params: pricefeedtypes.Params{
			Markets: []pricefeedtypes.Market{
				{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeedtypes.PostedPrice{
			{
				MarketID:      "usdx:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("1.00"),
				Expiry:        time.Now().Add(100 * time.Hour),
			},
			{
				MarketID:      "kava:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("2.00"),
				Expiry:        time.Now().Add(100 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("10.00"),
				Expiry:        time.Now().Add(100 * time.Hour),
			},
		},
	}

	hardGS := hardtypes.NewGenesisState(hardtypes.NewParams(
		hardtypes.MoneyMarkets{
			hardtypes.NewMoneyMarket(
				"usdx",
				hardtypes.NewBorrowLimit(
					true,
					sdk.MustNewDecFromStr("20000000"),
					sdk.MustNewDecFromStr("1"),
				),
				"usdx:usd",
				sdk.NewInt(1000000),
				hardtypes.NewInterestRateModel(
					sdk.MustNewDecFromStr("0.05"),
					sdk.MustNewDecFromStr("2"),
					sdk.MustNewDecFromStr("0.8"),
					sdk.MustNewDecFromStr("10"),
				),
				sdk.MustNewDecFromStr("0.05"),
				sdk.ZeroDec(),
			),
		},
		sdk.NewDec(10),
	),
		hardtypes.DefaultAccumulationTimes,
		hardtypes.DefaultDeposits,
		hardtypes.DefaultBorrows,
		hardtypes.DefaultTotalSupplied,
		hardtypes.DefaultTotalBorrowed,
		hardtypes.DefaultTotalReserves,
	)

	tApp := app.NewTestApp()

	tApp.InitializeFromGenesisStates(
		app.GenesisState{
			pricefeedtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&pricefeedGS),
			hardtypes.ModuleName:      tApp.AppCodec().MustMarshalJSON(&hardGS),
		},
	)

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	suite.Ctx = ctx
	suite.App = tApp
	suite.Keeper = tApp.GetEarnKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()

	suite.HardKeeper = tApp.GetHardKeeper()
	suite.SavingsKeeper = tApp.GetSavingsKeeper()

	hard.BeginBlocker(suite.Ctx, suite.HardKeeper)
}

// GetEvents returns emitted events on the sdk context
func (suite *Suite) GetEvents() sdk.Events {
	return suite.Ctx.EventManager().Events()
}

// AddCoinsToModule adds coins to the earn module account
func (suite *Suite) AddCoinsToModule(amount sdk.Coins) {
	// Does not use suite.BankKeeper.MintCoins as module account would not have permission to mint
	err := simapp.FundModuleAccount(suite.BankKeeper, suite.Ctx, types.ModuleName, amount)
	suite.Require().NoError(err)
}

// RemoveCoinsFromModule removes coins to the earn module account
func (suite *Suite) RemoveCoinsFromModule(amount sdk.Coins) {
	// Earn module does not have BurnCoins permission so we need to transfer to gov first to burn
	err := suite.BankKeeper.SendCoinsFromModuleToModule(suite.Ctx, types.ModuleAccountName, govtypes.ModuleName, amount)
	suite.Require().NoError(err)
	err = suite.BankKeeper.BurnCoins(suite.Ctx, govtypes.ModuleName, amount)
	suite.Require().NoError(err)
}

// CreateAccount creates a new account from the provided balance, using index
// to create different new addresses.
func (suite *Suite) CreateAccount(initialBalance sdk.Coins, index int) authtypes.AccountI {
	_, addrs := app.GeneratePrivKeyAddressPairs(index + 1)
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addrs[index])
	ak.SetAccount(suite.Ctx, acc)

	err := simapp.FundAccount(suite.BankKeeper, suite.Ctx, acc.GetAddress(), initialBalance)
	suite.Require().NoError(err)

	return acc
}

// NewAccountFromAddr creates a new account from the provided address with the provided balance
func (suite *Suite) NewAccountFromAddr(addr sdk.AccAddress, balance sdk.Coins) authtypes.AccountI {
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addr)
	ak.SetAccount(suite.Ctx, acc)

	err := simapp.FundAccount(suite.BankKeeper, suite.Ctx, acc.GetAddress(), balance)
	suite.Require().NoError(err)

	return acc
}

// CreateVault adds a new vault to the keeper parameters
func (suite *Suite) CreateVault(vaultDenom string, vaultStrategy types.StrategyType) {
	vault := types.NewAllowedVault(vaultDenom, vaultStrategy)
	suite.Require().NoError(vault.Validate())

	allowedVaults := suite.Keeper.GetAllowedVaults(suite.Ctx)
	allowedVaults = append(allowedVaults, vault)

	suite.Keeper.SetParams(
		suite.Ctx,
		types.NewParams(allowedVaults),
	)
}

// AccountBalanceEqual asserts that the coins match the account balance
func (suite *Suite) AccountBalanceEqual(addr sdk.AccAddress, coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(suite.Ctx, addr)
	suite.Equal(coins, balance, fmt.Sprintf("expected account balance to equal coins %s, but got %s", coins, balance))
}

// ModuleAccountBalanceEqual asserts that the earn module account balance matches the provided coins
func (suite *Suite) ModuleAccountBalanceEqual(coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(
		suite.Ctx,
		suite.AccountKeeper.GetModuleAddress(types.ModuleAccountName),
	)
	suite.Equal(coins, balance, fmt.Sprintf("expected module account balance to equal coins %s, but got %s", coins, balance))
}

// ----------------------------------------------------------------------------
// Earn

func (suite *Suite) VaultTotalValuesEqual(expected sdk.Coins) {
	for _, coin := range expected {
		vaultBal, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, coin.Denom)
		suite.Require().NoError(err, "failed to get vault balance")
		suite.Require().Equal(coin, vaultBal)
	}
}

func (suite *Suite) VaultTotalSuppliedEqual(expected sdk.Coins) {
	for _, coin := range expected {
		vaultBal, err := suite.Keeper.GetVaultTotalSupplied(suite.Ctx, coin.Denom)
		suite.Require().NoError(err, "failed to get vault balance")
		suite.Require().Equal(coin, vaultBal)
	}
}

func (suite *Suite) AccountTotalSuppliedEqual(accs []sdk.AccAddress, supplies []sdk.Coins) {
	for i, acc := range accs {
		coins := supplies[i]

		for _, coin := range coins {
			accVaultBal, err := suite.Keeper.GetVaultAccountSupplied(suite.Ctx, coin.Denom, acc)
			suite.Require().NoError(err)
			suite.Require().Equal(coin, accVaultBal)
		}
	}
}

// ----------------------------------------------------------------------------
// Hard

func (suite *Suite) HardDepositAmountEqual(expected sdk.Coins) {
	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)

	hardDeposit, found := suite.HardKeeper.GetSyncedDeposit(suite.Ctx, macc.GetAddress())
	if expected.IsZero() {
		suite.Require().False(found)
		return
	}

	suite.Require().True(found, "hard should have a deposit")
	suite.Require().Equalf(
		expected,
		hardDeposit.Amount,
		"hard should have a deposit with the amount %v",
		expected,
	)
}

// ----------------------------------------------------------------------------

// EventsContains asserts that the expected event is in the provided events
func (suite *Suite) EventsContains(events sdk.Events, expectedEvent sdk.Event) {
	foundMatch := false
	for _, event := range events {
		if event.Type == expectedEvent.Type {
			if reflect.DeepEqual(attrsToMap(expectedEvent.Attributes), attrsToMap(event.Attributes)) {
				foundMatch = true
			}
		}
	}

	suite.True(foundMatch, fmt.Sprintf("event of type %s not found or did not match", expectedEvent.Type))
}

func attrsToMap(attrs []abci.EventAttribute) []sdk.Attribute { // new cosmos changed the event attribute type
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}
