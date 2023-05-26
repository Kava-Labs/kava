package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type grpcQueryTestSuite struct {
	suite.Suite

	App         app.TestApp
	Keeper      keeper.Keeper
	Ctx         sdk.Context
	QueryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.App = app.NewTestApp()
	suite.App.InitializeFromGenesisStates()

	suite.Keeper = suite.App.GetEvmutilKeeper()
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.App.GetEvmutilKeeper()))
	suite.QueryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestQueryParams() {
	expectedParams := types.DefaultParams()
	expectedParams.AllowedCosmosDenoms = append(
		expectedParams.AllowedCosmosDenoms,
		types.NewAllowedCosmosCoinERC20Token("cosmos-denom", "Cosmos Coin", "COSMOS", 6),
	)
	expectedParams.EnabledConversionPairs = types.NewConversionPairs(
		types.NewConversionPair(testutil.RandomInternalEVMAddress(), "evm-denom"),
		types.NewConversionPair(testutil.RandomInternalEVMAddress(), "evm-denom2"),
	)
	suite.Keeper.SetParams(suite.Ctx, expectedParams)

	params, err := suite.QueryClient.Params(
		context.Background(),
		&types.QueryParamsRequest{},
	)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedParams, params.Params)
}
