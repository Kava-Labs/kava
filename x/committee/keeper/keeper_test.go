package keeper_test

import (
	"testing"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {

	suite.app = app.NewTestApp() 
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx =  suite.app.NewContext(true, abci.Header{})
}

func (suite *KeeperTestSuite) TestGetSetCommittee() {
}


func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}