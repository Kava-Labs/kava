package testutil

import (
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

// Suite implements a test suite for the module integration tests
type Suite struct {
	suite.Suite

	Keeper      keeper.Keeper
	BankKeeper  bankkeeper.Keeper
	App         app.TestApp
	Ctx         sdk.Context
	QueryClient types.QueryClient
	Addresses   []sdk.AccAddress
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	suite.App = app.NewTestApp()
	suite.Keeper = suite.App.GetCommitteeKeeper()
	suite.BankKeeper = suite.App.GetBankKeeper()
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{})
	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	suite.Addresses = accAddresses

	// Set query client
	queryHelper := suite.App.NewQueryServerTestHelper(suite.Ctx)
	queryHandler := keeper.NewQueryServerImpl(suite.Keeper)
	types.RegisterQueryServer(queryHelper, queryHandler)
	suite.QueryClient = types.NewQueryClient(queryHelper)
}
