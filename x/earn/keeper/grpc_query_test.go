package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
)

type grpcQueryTestSuite struct {
	testutil.Suite

	queryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestQueryParams() {
	vaultDenom := "usdx"

	res, err := suite.queryClient.Params(context.Background(), types.NewQueryParamsRequest())
	suite.Require().NoError(err)
	// ElementsMatch instead of Equal because AllowedVaults{} != AllowedVaults(nil)
	suite.Require().ElementsMatch(types.DefaultParams().AllowedVaults, res.Params.AllowedVaults)

	// Add vault to params
	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	// Query again for added vault
	res, err = suite.queryClient.Params(context.Background(), types.NewQueryParamsRequest())
	suite.Require().NoError(err)
	suite.Require().Equal(
		types.AllowedVaults{
			types.NewAllowedVault(vaultDenom, types.STRATEGY_TYPE_HARD),
		},
		res.Params.AllowedVaults,
	)
}

func (suite *grpcQueryTestSuite) TestTotalDeposited_NoSupply() {
	// Add vaults
	suite.CreateVault("usdx", types.STRATEGY_TYPE_HARD)
	suite.CreateVault("cats", types.STRATEGY_TYPE_HARD)

	// Query again for added vault
	res, err := suite.queryClient.TotalDeposited(context.Background(), types.NewQueryTotalDepositedRequest(""))
	suite.Require().NoError(err)
	suite.Require().True(res.SuppliedCoins.Empty(), "supplied coins should be empty")
}

func (suite *grpcQueryTestSuite) TestTotalDeposited_All() {
	// Add vaults
	suite.CreateVault("usdx", types.STRATEGY_TYPE_HARD)
	suite.CreateVault("cats", types.STRATEGY_TYPE_HARD)

	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0).GetAddress()
	err := suite.Keeper.Deposit(suite.Ctx, acc, depositAmount)
	suite.Require().NoError(err)

	// Query again for added vault
	res, err := suite.queryClient.TotalDeposited(context.Background(), types.NewQueryTotalDepositedRequest(""))
	suite.Require().NoError(err)
	suite.Require().Equal(
		sdk.NewCoins(depositAmount),
		res.SuppliedCoins,
		"supplied coins should be sum of all supplied coins",
	)
}
