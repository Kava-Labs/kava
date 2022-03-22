package savings_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/savings/keeper"
	"github.com/kava-labs/kava/x/savings/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app     app.TestApp
	genTime time.Time
	ctx     sdk.Context
	keeper  keeper.Keeper
	addrs   []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.genTime = tmtime.Canonical(time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC))
	suite.ctx = tApp.NewContext(true, tmproto.Header{Height: 1, Time: suite.genTime})
	suite.keeper = tApp.GetSavingsKeeper()
	suite.app = tApp

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.addrs = addrs
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	params := types.NewParams()
	savingsGenesis := types.NewGenesisState(params)

	suite.NotPanics(
		func() {
			suite.app.InitializeFromGenesisStatesWithTime(
				suite.genTime,
				app.GenesisState{types.ModuleName: suite.app.AppCodec().MustMarshalJSON(&savingsGenesis)},
			)
		},
	)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
