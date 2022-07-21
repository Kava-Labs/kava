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

func (suite *grpcQueryTestSuite) TestVaults_ZeroSupply() {
	// Add vaults
	suite.CreateVault("usdx", types.STRATEGY_TYPE_HARD)
	suite.CreateVault("busd", types.STRATEGY_TYPE_HARD)

	suite.Run("single", func() {
		res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest("usdx"))
		suite.Require().NoError(err)
		suite.Require().Len(res.Vaults, 1)
		suite.Require().Equal(
			types.VaultResponse{
				Denom:         "usdx",
				VaultStrategy: types.STRATEGY_TYPE_HARD,
				TotalSupplied: sdk.NewInt(0),
				TotalValue:    sdk.NewInt(0),
			},
			res.Vaults[0],
		)
	})

	suite.Run("all", func() {
		res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest(""))
		suite.Require().NoError(err)
		suite.Require().Len(res.Vaults, 2)
		suite.Require().ElementsMatch(
			[]types.VaultResponse{
				{
					Denom:         "usdx",
					VaultStrategy: types.STRATEGY_TYPE_HARD,
					TotalSupplied: sdk.NewInt(0),
					TotalValue:    sdk.NewInt(0),
				},
				{
					Denom:         "busd",
					VaultStrategy: types.STRATEGY_TYPE_HARD,
					TotalSupplied: sdk.NewInt(0),
					TotalValue:    sdk.NewInt(0),
				},
			},
			res.Vaults,
		)
	})
}

func (suite *grpcQueryTestSuite) TestVaults_WithSupply() {
	vaultDenom := "usdx"

	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest("usdx"))
	suite.Require().NoError(err)
	suite.Require().Len(res.Vaults, 1)
	suite.Require().Equal(
		types.VaultResponse{
			Denom:         "usdx",
			VaultStrategy: types.STRATEGY_TYPE_HARD,
			TotalSupplied: depositAmount.Amount,
			TotalValue:    depositAmount.Amount,
		},
		res.Vaults[0],
	)
}

func (suite *grpcQueryTestSuite) TestTotalDeposited_NoSupply() {
	// Add vaults
	suite.CreateVault("usdx", types.STRATEGY_TYPE_HARD)
	suite.CreateVault("cats", types.STRATEGY_TYPE_HARD)

	res, err := suite.queryClient.TotalDeposited(context.Background(), types.NewQueryTotalDepositedRequest(""))
	suite.Require().NoError(err)
	suite.Require().True(res.SuppliedCoins.Empty(), "supplied coins should be empty")
}

func (suite *grpcQueryTestSuite) TestTotalDeposited_All() {
	vault1Denom := "usdx"
	vault2Denom := "busd"

	// Add vaults
	suite.CreateVault(vault1Denom, types.STRATEGY_TYPE_HARD)
	suite.CreateVault(vault2Denom, types.STRATEGY_TYPE_HARD)

	startBalance := sdk.NewCoins(
		sdk.NewInt64Coin(vault1Denom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
	)
	deposit1Amount := sdk.NewInt64Coin(vault1Denom, 100)
	deposit2Amount := sdk.NewInt64Coin(vault2Denom, 100)

	acc := suite.CreateAccount(startBalance, 0).GetAddress()
	err := suite.Keeper.Deposit(suite.Ctx, acc, deposit1Amount)
	suite.Require().NoError(err)

	res, err := suite.queryClient.TotalDeposited(
		context.Background(),
		types.NewQueryTotalDepositedRequest(""), // query all
	)
	suite.Require().NoError(err)
	suite.Require().Equal(
		sdk.NewCoins(deposit1Amount),
		res.SuppliedCoins,
		"supplied coins should be sum of all supplied coins",
	)

	err = suite.Keeper.Deposit(suite.Ctx, acc, deposit2Amount)
	suite.Require().NoError(err)

	res, err = suite.queryClient.TotalDeposited(
		context.Background(),
		types.NewQueryTotalDepositedRequest(""), // query all
	)
	suite.Require().NoError(err)
	suite.Require().Equal(
		sdk.NewCoins(deposit1Amount, deposit2Amount),
		res.SuppliedCoins,
		"supplied coins should be sum of all supplied coins for multiple coins",
	)
}

func (suite *grpcQueryTestSuite) TestTotalDeposited_Single() {
	vault1Denom := "usdx"
	vault2Denom := "busd"

	// Add vaults
	suite.CreateVault(vault1Denom, types.STRATEGY_TYPE_HARD)
	suite.CreateVault(vault2Denom, types.STRATEGY_TYPE_HARD)

	startBalance := sdk.NewCoins(
		sdk.NewInt64Coin(vault1Denom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
	)
	deposit1Amount := sdk.NewInt64Coin(vault1Denom, 100)
	deposit2Amount := sdk.NewInt64Coin(vault2Denom, 100)

	acc := suite.CreateAccount(startBalance, 0).GetAddress()
	err := suite.Keeper.Deposit(suite.Ctx, acc, deposit1Amount)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc, deposit2Amount)
	suite.Require().NoError(err)

	res, err := suite.queryClient.TotalDeposited(
		context.Background(),
		types.NewQueryTotalDepositedRequest(vault1Denom),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(
		sdk.NewCoins(deposit1Amount),
		res.SuppliedCoins,
		"should only contain queried denom",
	)

	res, err = suite.queryClient.TotalDeposited(
		context.Background(),
		types.NewQueryTotalDepositedRequest(vault2Denom),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(
		sdk.NewCoins(deposit2Amount),
		res.SuppliedCoins,
		"should only contain queried denom",
	)
}
