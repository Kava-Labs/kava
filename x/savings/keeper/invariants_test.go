package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/savings/keeper"
	"github.com/kava-labs/kava/x/savings/types"
)

type invariantTestSuite struct {
	suite.Suite

	tApp       app.TestApp
	ctx        sdk.Context
	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper
	addrs      []sdk.AccAddress
	invariants map[string]map[string]sdk.Invariant
}

func (suite *invariantTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	suite.addrs = addrs

	suite.tApp = tApp
	suite.ctx = ctx
	suite.keeper = tApp.GetSavingsKeeper()
	suite.bankKeeper = tApp.GetBankKeeper()

	suite.invariants = make(map[string]map[string]sdk.Invariant)
	keeper.RegisterInvariants(suite, suite.keeper)
}

func (suite *invariantTestSuite) RegisterRoute(moduleName string, route string, invariant sdk.Invariant) {
	_, exists := suite.invariants[moduleName]

	if !exists {
		suite.invariants[moduleName] = make(map[string]sdk.Invariant)
	}

	suite.invariants[moduleName][route] = invariant
}

func (suite *invariantTestSuite) SetupValidState() {
	depositAmt := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(2e8)))

	suite.keeper.SetDeposit(suite.ctx, types.NewDeposit(
		suite.addrs[0],
		depositAmt,
	))

	err := suite.tApp.FundModuleAccount(suite.ctx, types.ModuleName, depositAmt)
	suite.Require().NoError(err)
}

func (suite *invariantTestSuite) runInvariant(route string, invariant func(k keeper.Keeper) sdk.Invariant) (string, bool) {
	ctx := suite.ctx
	registeredInvariant := suite.invariants[types.ModuleName][route]
	suite.Require().NotNil(registeredInvariant)

	// direct call
	dMessage, dBroken := invariant(suite.keeper)(ctx)
	// registered call
	rMessage, rBroken := registeredInvariant(ctx)
	// all call
	aMessage, aBroken := keeper.AllInvariants(suite.keeper)(ctx)

	// require matching values for direct call and registered call
	suite.Require().Equal(dMessage, rMessage, "expected registered invariant message to match")
	suite.Require().Equal(dBroken, rBroken, "expected registered invariant broken to match")
	// require matching values for direct call and all invariants call if broken
	suite.Require().Equal(dBroken, aBroken, "expected all invariant broken to match")
	if dBroken {
		suite.Require().Equal(dMessage, aMessage, "expected all invariant message to match")
	}

	// return message, broken
	return dMessage, dBroken
}

func (suite *invariantTestSuite) TestDepositsInvariant() {
	message, broken := suite.runInvariant("deposits", keeper.DepositsInvariant)
	suite.Equal("savings: validate deposits broken invariant\ndeposit invalid\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("deposits", keeper.DepositsInvariant)
	suite.Equal("savings: validate deposits broken invariant\ndeposit invalid\n", message)
	suite.Equal(false, broken)

	// broken with invalid deposit
	suite.keeper.SetDeposit(suite.ctx, types.NewDeposit(
		suite.addrs[0],
		sdk.Coins{},
	))

	message, broken = suite.runInvariant("deposits", keeper.DepositsInvariant)
	suite.Equal("savings: validate deposits broken invariant\ndeposit invalid\n", message)
	suite.Equal(true, broken)
}

func (suite *invariantTestSuite) TestSolvencyInvariant() {
	message, broken := suite.runInvariant("solvency", keeper.SolvencyInvariant)
	suite.Equal("savings: module solvency broken invariant\ntotal deposited amount does not match module account\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("solvency", keeper.SolvencyInvariant)
	suite.Equal("savings: module solvency broken invariant\ntotal deposited amount does not match module account\n", message)
	suite.Equal(false, broken)

	// broken when deposits are greater than module balance
	suite.keeper.SetDeposit(suite.ctx, types.NewDeposit(
		suite.addrs[0],
		sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(3e8))),
	))

	message, broken = suite.runInvariant("solvency", keeper.SolvencyInvariant)
	suite.Equal("savings: module solvency broken invariant\ntotal deposited amount does not match module account\n", message)
	suite.Equal(true, broken)

	// broken when deposits are less than the module balance
	suite.keeper.SetDeposit(suite.ctx, types.NewDeposit(
		suite.addrs[0],
		sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e8))),
	))

	message, broken = suite.runInvariant("solvency", keeper.SolvencyInvariant)
	suite.Equal("savings: module solvency broken invariant\ntotal deposited amount does not match module account\n", message)
	suite.Equal(true, broken)
}

func TestInvariantTestSuite(t *testing.T) {
	suite.Run(t, new(invariantTestSuite))
}
