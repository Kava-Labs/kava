package testutil

import (
	"fmt"
	"reflect"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	BankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// Suite implements a test suite for the swap module integration tests
type Suite struct {
	suite.Suite
	Keeper        keeper.Keeper
	App           app.TestApp
	Ctx           sdk.Context
	BankKeeper    BankKeeper.Keeper
	AccountKeeper authkeeper.AccountKeeper
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	suite.Ctx = ctx
	suite.App = tApp
	suite.Keeper = tApp.GetEarnKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()
}

// GetEvents returns emitted events on the sdk context
func (suite *Suite) GetEvents() sdk.Events {
	return suite.Ctx.EventManager().Events()
}

// AddCoinsToModule adds coins to the swap module account
func (suite *Suite) AddCoinsToModule(amount sdk.Coins) {
	// Does not use suite.BankKeeper.MintCoins as module account would not have permission to mint
	err := simapp.FundModuleAccount(suite.BankKeeper, suite.Ctx, types.ModuleName, amount)
	suite.Require().NoError(err)
}

// RemoveCoinsFromModule removes coins to the swap module account
func (suite *Suite) RemoveCoinsFromModule(amount sdk.Coins) {
	// Swap module does not have BurnCoins permission so we need to transfer to gov first to burn
	err := suite.BankKeeper.SendCoinsFromModuleToModule(suite.Ctx, types.ModuleAccountName, govtypes.ModuleName, amount)
	suite.Require().NoError(err)
	err = suite.BankKeeper.BurnCoins(suite.Ctx, govtypes.ModuleName, amount)
	suite.Require().NoError(err)
}

// CreateAccount creates a new account from the provided balance
func (suite *Suite) CreateAccount(initialBalance sdk.Coins) authtypes.AccountI {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addrs[0])
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

	// allowedVaults := suite.Keeper.GetAllowedVaults(suite.Ctx)
	// allowedVaults = append(allowedVaults, vault)

	suite.Keeper.SetParams(
		suite.Ctx,
		types.NewParams(types.AllowedVaults{vault}),
	)
}

// AccountBalanceEqual asserts that the coins match the account balance
func (suite *Suite) AccountBalanceEqual(addr sdk.AccAddress, coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(suite.Ctx, addr)
	suite.Equal(coins, balance, fmt.Sprintf("expected account balance to equal coins %s, but got %s", coins, balance))
}

// ModuleAccountBalanceEqual asserts that the swap module account balance matches the provided coins
func (suite *Suite) ModuleAccountBalanceEqual(coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(
		suite.Ctx,
		suite.AccountKeeper.GetModuleAddress(types.ModuleAccountName),
	)
	suite.Equal(coins, balance, fmt.Sprintf("expected module account balance to equal coins %s, but got %s", coins, balance))
}

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
