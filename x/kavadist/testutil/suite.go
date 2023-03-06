package testutil

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// Suite implements a test suite for the kavadist module integration tests
type Suite struct {
	suite.Suite

	Keeper        keeper.Keeper
	BankKeeper    bankkeeper.Keeper
	AccountKeeper accountkeeper.AccountKeeper
	App           app.TestApp
	Ctx           sdk.Context
	TestPeriods   []types.Period
	Addrs         []sdk.AccAddress
	QueryClient   types.QueryClient
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	coins := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000000000)))
	authGS := app.NewFundedGenStateWithSameCoins(tApp.AppCodec(), coins, addrs)

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	testPeriods := []types.Period{
		{
			Start:     time.Date(2020, time.March, 1, 1, 0, 0, 0, time.UTC),
			End:       time.Date(2021, time.March, 1, 1, 0, 0, 0, time.UTC),
			Inflation: sdk.MustNewDecFromStr("1.000000003022265980"),
		},
	}
	params := types.NewParams(true, testPeriods, types.DefaultInfraParams)
	moduleGs := types.ModuleCdc.MustMarshalJSON(types.NewGenesisState(params, types.DefaultPreviousBlockTime))
	gs := app.GenesisState{types.ModuleName: moduleGs}
	suite.App = tApp.InitializeFromGenesisStates(authGS, gs)
	suite.Ctx = ctx
	suite.Addrs = addrs
	suite.Keeper = tApp.GetKavadistKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()
	suite.TestPeriods = testPeriods

	// Set query client
	queryHelper := tApp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))
	suite.QueryClient = types.NewQueryClient(queryHelper)
}

// CreateAccount creates a new account with the provided balance
func (suite *Suite) CreateAccount(initialBalance sdk.Coins) authtypes.AccountI {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	fmt.Println(addrs[0].String())
	acc := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, addrs[0])
	suite.AccountKeeper.SetAccount(suite.Ctx, acc)
	suite.Require().NoError(suite.App.FundAccount(suite.Ctx, addrs[0], initialBalance))
	suite.AccountKeeper.SetAccount(suite.Ctx, acc)
	return acc
}
