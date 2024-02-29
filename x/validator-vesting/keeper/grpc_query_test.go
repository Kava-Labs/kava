package keeper_test

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/validator-vesting/keeper"
	"github.com/kava-labs/kava/x/validator-vesting/types"
)

type grpcQueryTestSuite struct {
	suite.Suite
	app         app.TestApp
	ctx         sdk.Context
	queryClient types.QueryClient
	bk          *mockBankKeeper
}

type mockBankKeeper struct {
	supply sdk.Coin
}

func (m *mockBankKeeper) SetSupply(ctx sdk.Context, denom string, amt sdkmath.Int) {
	m.supply = sdk.NewCoin(denom, amt)
}

func (m *mockBankKeeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	return m.supply
}

func (suite *grpcQueryTestSuite) SetupTest() {
	testTime := time.Date(2024, 2, 29, 12, 00, 00, 00, time.UTC)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: testTime})
	suite.app = tApp
	suite.ctx = ctx
	suite.bk = &mockBankKeeper{}
	suite.queryClient = suite.queryClientWithBlockTime(testTime)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestCirculatingSupply() {
	suite.Run("vesting period supply", func() {
		suite.bk.SetSupply(suite.ctx, "ukava", sdkmath.NewInt(2_500_000_000_000))
		lastVestingPeriod := time.Date(2022, 8, 5, 24, 0, 0, 0, time.UTC)
		queryClient := suite.queryClientWithBlockTime(lastVestingPeriod)
		res, err := queryClient.CirculatingSupply(context.Background(), &types.QueryCirculatingSupplyRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(sdkmath.NewInt(15_625), res.Amount)
	})

	suite.Run("supply after last vesting period", func() {
		suite.bk.SetSupply(suite.ctx, "ukava", sdkmath.NewInt(100_000_000))
		res, err := suite.queryClient.CirculatingSupply(context.Background(), &types.QueryCirculatingSupplyRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(sdkmath.NewInt(100), res.Amount)
	})
}

func (suite *grpcQueryTestSuite) TestTotalSupply() {
	suite.bk.SetSupply(suite.ctx, "ukava", sdkmath.NewInt(100_000_000))
	res, err := suite.queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(100), res.Amount)
}

func (suite *grpcQueryTestSuite) TestCirculatingSupplyHARD() {
	res, err := suite.queryClient.CirculatingSupplyHARD(context.Background(), &types.QueryCirculatingSupplyHARDRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(188333338), res.Amount)
}

func (suite *grpcQueryTestSuite) TestCirculatingSupplyUSDX() {
	suite.bk.SetSupply(suite.ctx, "usdx", sdkmath.NewInt(150_000_000))
	res, err := suite.queryClient.CirculatingSupplyUSDX(context.Background(), &types.QueryCirculatingSupplyUSDXRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(150), res.Amount)
}

func (suite *grpcQueryTestSuite) TestCirculatingSupplySWP() {
	res, err := suite.queryClient.CirculatingSupplySWP(suite.ctx, &types.QueryCirculatingSupplySWPRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(201302073), res.Amount)
}

func (suite *grpcQueryTestSuite) TestTotalSupplyHARD() {
	suite.bk.SetSupply(suite.ctx, "hard", sdkmath.NewInt(150_000_000))
	res, err := suite.queryClient.TotalSupplyHARD(context.Background(), &types.QueryTotalSupplyHARDRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(150), res.Amount)
}

func (suite *grpcQueryTestSuite) TestTotalSupplyUSDX() {
	suite.bk.SetSupply(suite.ctx, "usdx", sdkmath.NewInt(150_000_000))
	res, err := suite.queryClient.TotalSupplyUSDX(context.Background(), &types.QueryTotalSupplyUSDXRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(150), res.Amount)
}

func (suite *grpcQueryTestSuite) queryClientWithBlockTime(blockTime time.Time) types.QueryClient {
	ctx := suite.ctx.WithBlockTime(blockTime)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.bk))
	return types.NewQueryClient(queryHelper)
}
