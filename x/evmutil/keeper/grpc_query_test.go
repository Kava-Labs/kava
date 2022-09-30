package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type GrpcQueryTestSuite struct {
	testutil.Suite

	msgServer types.MsgServer
}

func (suite *GrpcQueryTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.msgServer = keeper.NewMsgServerImpl(suite.App.GetEvmutilKeeper())
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GrpcQueryTestSuite))
}

func (suite *GrpcQueryTestSuite) TestQueryParams() {
	params, err := suite.QueryClient.Params(
		context.Background(),
		&types.QueryParamsRequest{},
	)
	suite.Require().NoError(err)

	suite.Require().Len(params.Params.EnabledConversionPairs, 1)
}
