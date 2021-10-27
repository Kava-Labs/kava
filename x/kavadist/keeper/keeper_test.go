package keeper_test

import (
	"testing"
	"time"

	"github.com/kava-labs/kava/x/kavadist/testutil"
	"github.com/kava-labs/kava/x/kavadist/types"
	"github.com/stretchr/testify/suite"
)

type keeperTestSuite struct {
	testutil.Suite
}

func (suite *keeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}

func (suite *keeperTestSuite) TestGetPreviousBlock_NoPreviousBlock() {
	blockTime, found := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
	suite.False(found)
	suite.Equal(blockTime, types.DefaultPreviousBlockTime)
}

func (suite *keeperTestSuite) TestSetAndGetPreviousBlockTime() {
	newTime := time.Date(2020, time.March, 1, 1, 0, 0, 0, time.UTC)
	suite.Keeper.SetPreviousBlockTime(suite.Ctx, newTime)
	blockTime, found := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
	suite.True(found)
	suite.Equal(newTime, blockTime)
}
