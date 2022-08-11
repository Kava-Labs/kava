package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)

	// Query again for added vault
	res, err = suite.queryClient.Params(context.Background(), types.NewQueryParamsRequest())
	suite.Require().NoError(err)
	suite.Require().Equal(
		types.AllowedVaults{
			types.NewAllowedVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil),
		},
		res.Params.AllowedVaults,
	)
}

func (suite *grpcQueryTestSuite) TestVaults_ZeroSupply() {
	// Add vaults
	suite.CreateVault("usdx", types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("busd", types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)

	suite.Run("single", func() {
		res, err := suite.queryClient.Vault(context.Background(), types.NewQueryVaultRequest("usdx"))
		suite.Require().NoError(err)
		suite.Require().Equal(
			types.VaultResponse{
				Denom:             "usdx",
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       sdk.NewDec(0).String(),
				TotalValue:        sdk.NewInt(0),
			},
			res.Vault,
		)
	})

	suite.Run("all", func() {
		res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest())
		suite.Require().NoError(err)
		suite.Require().Len(res.Vaults, 2)
		suite.Require().ElementsMatch(
			[]types.VaultResponse{
				{
					Denom:             "usdx",
					Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
					IsPrivateVault:    false,
					AllowedDepositors: nil,
					TotalShares:       sdk.NewDec(0).String(),
					TotalValue:        sdk.NewInt(0),
				},
				{
					Denom:             "busd",
					Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
					IsPrivateVault:    false,
					AllowedDepositors: nil,
					TotalShares:       sdk.NewDec(0).String(),
					TotalValue:        sdk.NewInt(0),
				},
			},
			res.Vaults,
		)
	})
}

func (suite *grpcQueryTestSuite) TestVaults_WithSupply() {
	vaultDenom := "usdx"
	vault2Denom := "bkava-kavavaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l"

	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	deposit2Amount := sdk.NewInt64Coin(vault2Denom, 100)

	suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("bkava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(
		sdk.NewInt64Coin(vaultDenom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
	), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), deposit2Amount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest())
	suite.Require().NoError(err)
	suite.Require().Len(res.Vaults, 2)
	suite.Require().ElementsMatch(
		[]types.VaultResponse{
			{
				Denom:             vaultDenom,
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       depositAmount.Amount.ToDec().String(),
				TotalValue:        depositAmount.Amount,
			},
			{
				Denom:             vault2Denom,
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_SAVINGS},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       deposit2Amount.Amount.ToDec().String(),
				TotalValue:        deposit2Amount.Amount,
			},
		},
		res.Vaults,
	)
}

func (suite *grpcQueryTestSuite) TestVault_NotFound() {
	_, err := suite.queryClient.Vault(context.Background(), types.NewQueryVaultRequest("usdx"))
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Errorf(codes.NotFound, "vault not found with specified denom"))
}

func (suite *grpcQueryTestSuite) TestDeposits() {
	vault1Denom := "usdx"
	vault2Denom := "busd"
	vault3Denom := "bkava-kavavaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l"

	// Add vaults
	suite.CreateVault(vault1Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault(vault2Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("bkava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	startBalance := sdk.NewCoins(
		sdk.NewInt64Coin(vault1Denom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
		sdk.NewInt64Coin(vault3Denom, 1000),
	)
	deposit1Amount := sdk.NewInt64Coin(vault1Denom, 100)
	deposit2Amount := sdk.NewInt64Coin(vault2Denom, 200)
	deposit3Amount := sdk.NewInt64Coin(vault3Denom, 200)

	// Accounts
	acc1 := suite.CreateAccount(startBalance, 0).GetAddress()
	acc2 := suite.CreateAccount(startBalance, 1).GetAddress()

	// Deposit into each vault from each account - 4 total deposits
	// Acc 1: usdx + busd
	// Acc 2: usdx + usdc
	err := suite.Keeper.Deposit(suite.Ctx, acc1, deposit1Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)
	err = suite.Keeper.Deposit(suite.Ctx, acc1, deposit2Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc2, deposit1Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)
	err = suite.Keeper.Deposit(suite.Ctx, acc2, deposit3Amount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	suite.Run("1) 1 vault for 1 account", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), vault1Denom, nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatchf(
			[]types.DepositResponse{
				{
					Depositor: acc1.String(),
					// Still includes all deposits
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit1Amount.Denom, deposit1Amount.Amount.ToDec()),
						types.NewVaultShare(deposit2Amount.Denom, deposit2Amount.Amount.ToDec()),
					),
					Value: sdk.NewCoins(deposit1Amount, deposit2Amount),
				},
			},
			res.Deposits,
			"deposits should match, got %v",
			res.Deposits,
		)
	})

	suite.Run("1) invalid vault for 1 account", func() {
		_, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), "notavaliddenom", nil),
		)
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, status.Errorf(codes.NotFound, "No deposit for denom notavaliddenom found for owner"))
	})

	suite.Run("3) all vaults for 1 account", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), "", nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatch(
			[]types.DepositResponse{
				{
					Depositor: acc1.String(),
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit1Amount.Denom, deposit1Amount.Amount.ToDec()),
						types.NewVaultShare(deposit2Amount.Denom, deposit2Amount.Amount.ToDec()),
					),
					Value: sdk.NewCoins(deposit1Amount, deposit2Amount),
				},
			},
			res.Deposits,
		)
	})

	suite.Run("2) all accounts, specific vault", func() {
		// Query all deposits for vault 3
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest("", vault3Denom, nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatch(
			[]types.DepositResponse{
				{
					Depositor: acc2.String(),
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit1Amount.Denom, deposit1Amount.Amount.ToDec()),
						types.NewVaultShare(deposit3Amount.Denom, deposit3Amount.Amount.ToDec()),
					),
					Value: sdk.NewCoins(deposit1Amount, deposit3Amount),
				},
			},
			res.Deposits,
		)
	})

	suite.Run("4) all vaults and all accounts", func() {
		// Query all deposits for all vaults
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest("", "", nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 2)
		suite.Require().ElementsMatchf(
			[]types.DepositResponse{
				{
					Depositor: acc1.String(),
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit1Amount.Denom, deposit1Amount.Amount.ToDec()),
						types.NewVaultShare(deposit2Amount.Denom, deposit2Amount.Amount.ToDec()),
					),
					Value: sdk.NewCoins(deposit1Amount, deposit2Amount),
				},
				{
					Depositor: acc2.String(),
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit1Amount.Denom, deposit1Amount.Amount.ToDec()),
						types.NewVaultShare(deposit3Amount.Denom, deposit3Amount.Amount.ToDec()),
					),
					Value: sdk.NewCoins(deposit1Amount, deposit3Amount),
				},
			},
			res.Deposits,
			"deposits should match, got %v",
			res.Deposits,
		)
	})
}

func (suite *grpcQueryTestSuite) TestDeposits_NotFound() {
	_, err := suite.queryClient.Deposits(
		context.Background(),
		types.NewQueryDepositsRequest("", "usdx", nil),
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Error(codes.NotFound, "Vault record for denom not found"))
}

func (suite *grpcQueryTestSuite) TestDeposits_InvalidAddress() {
	_, err := suite.queryClient.Deposits(
		context.Background(),
		types.NewQueryDepositsRequest("asdf", "usdx", nil),
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "Invalid address"))

	_, err = suite.queryClient.Deposits(
		context.Background(),
		types.NewQueryDepositsRequest("asdf", "", nil),
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "Invalid address"))
}

func (suite *grpcQueryTestSuite) TestVault_bKava() {
	vaultDenom := "bkava"
	coinDenom := vaultDenom + "-kavavaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l"

	startBalance := sdk.NewInt64Coin(coinDenom, 1000)
	depositAmount := sdk.NewInt64Coin(coinDenom, 100)

	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	// vault denom is only "bkava" which has it's own special handler
	suite.CreateVault(
		vaultDenom,
		types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS},
		false,
		[]sdk.AccAddress{},
	)

	err := suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(
		err,
		"should be able to deposit bkava derivative denom in bkava vault",
	)

	res, err := suite.queryClient.Vault(
		context.Background(),
		types.NewQueryVaultRequest(coinDenom),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(
		types.VaultResponse{
			Denom: coinDenom,
			Strategies: types.StrategyTypes{
				types.STRATEGY_TYPE_SAVINGS,
			},
			IsPrivateVault:    false,
			AllowedDepositors: []string(nil),
			TotalShares:       "100.000000000000000000",
			TotalValue:        sdk.NewInt(100),
		},
		res.Vault,
	)
}
