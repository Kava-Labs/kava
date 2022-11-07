package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/keeper"
	"github.com/kava-labs/kava/x/liquid/types"
)

type grpcQueryTestSuite struct {
	KeeperTestSuite

	queryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestQueryDelegatedBalance() {
	zeroResponse := &types.QueryDelegatedBalanceResponse{
		Vested:  suite.NewBondCoin(sdk.ZeroInt()),
		Vesting: suite.NewBondCoin(sdk.ZeroInt()),
	}

	testCases := []struct {
		name        string
		setup       func() string
		expectedRes *types.QueryDelegatedBalanceResponse
		expectedErr error
	}{
		{
			name: "vesting account with stake less than vesting",
			setup: func() string {
				initBalance := suite.NewBondCoin(i(1e9))
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				valAddr, delAddr := addrs[0], addrs[1]

				suite.CreateAccountWithAddress(valAddr, sdk.NewCoins(initBalance))

				suite.CreateVestingAccountWithAddress(delAddr, sdk.NewCoins(initBalance), suite.NewBondCoins(initBalance.Amount.QuoRaw(2)))

				suite.CreateNewUnbondedValidator(sdk.ValAddress(valAddr), initBalance.Amount)
				suite.CreateDelegation(sdk.ValAddress(valAddr), delAddr, initBalance.Amount.QuoRaw(4))
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper) // bond the validator

				return delAddr.String()
			},
			expectedRes: &types.QueryDelegatedBalanceResponse{
				Vested:  suite.NewBondCoin(sdk.ZeroInt()),
				Vesting: suite.NewBondCoin(i(250e6)),
			},
		},
		{
			name: "vesting account with stake greater than vesting",
			setup: func() string {
				initBalance := suite.NewBondCoin(i(1e9))
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				valAddr, delAddr := addrs[0], addrs[1]

				suite.CreateAccountWithAddress(valAddr, sdk.NewCoins(initBalance))

				suite.CreateVestingAccountWithAddress(delAddr, sdk.NewCoins(initBalance), suite.NewBondCoins(initBalance.Amount.QuoRaw(2)))

				suite.CreateNewUnbondedValidator(sdk.ValAddress(valAddr), initBalance.Amount)
				threeQuarters := initBalance.Amount.QuoRaw(4).MulRaw(3)
				suite.CreateDelegation(sdk.ValAddress(valAddr), delAddr, threeQuarters)
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper) // bond the validator

				return delAddr.String()
			},
			expectedRes: &types.QueryDelegatedBalanceResponse{
				Vested:  suite.NewBondCoin(i(250e6)),
				Vesting: suite.NewBondCoin(i(500e6)),
			},
		},
		{
			name: "no account returns zeros",
			setup: func() string {
				return "kava10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t"
			},
			expectedRes: zeroResponse,
		},
		{
			name: "base account no delegations returns zeros",
			setup: func() string {
				acc := suite.CreateAccount(suite.NewBondCoins(i(1e9)), 0)
				return acc.GetAddress().String()
			},
			expectedRes: zeroResponse,
		},
		{
			name: "base account with delegations returns delegated",
			setup: func() string {
				initBalance := suite.NewBondCoin(i(1e9))
				val1Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 0)
				val2Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 1)
				delAcc := suite.CreateAccount(sdk.NewCoins(initBalance), 2)

				suite.CreateNewUnbondedValidator(val1Acc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateNewUnbondedValidator(val2Acc.GetAddress().Bytes(), initBalance.Amount)

				suite.CreateDelegation(val1Acc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount.QuoRaw(2))
				suite.CreateDelegation(val2Acc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount.QuoRaw(2))

				return delAcc.GetAddress().String()
			},
			expectedRes: &types.QueryDelegatedBalanceResponse{
				Vested:  suite.NewBondCoin(i(1e9)),
				Vesting: suite.NewBondCoin(sdk.ZeroInt()),
			},
		},
		{
			name: "base account with delegations and unbonding delegations returns only delegations",
			setup: func() string {
				initBalance := suite.NewBondCoin(i(1e9))
				valAcc := suite.CreateAccount(sdk.NewCoins(initBalance), 0)
				delAcc := suite.CreateAccount(sdk.NewCoins(initBalance), 1)

				suite.CreateNewUnbondedValidator(valAcc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateDelegation(valAcc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount)
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper) // bond the validator

				suite.CreateUnbondingDelegation(delAcc.GetAddress(), valAcc.GetAddress().Bytes(), initBalance.Amount.QuoRaw(2))

				return delAcc.GetAddress().String()
			},
			expectedRes: &types.QueryDelegatedBalanceResponse{
				Vested:  suite.NewBondCoin(i(500e6)),
				Vesting: suite.NewBondCoin(sdk.ZeroInt()),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			address := tc.setup()

			res, err := suite.queryClient.DelegatedBalance(
				context.Background(),
				&types.QueryDelegatedBalanceRequest{
					Delegator: address,
				},
			)
			suite.ErrorIs(err, tc.expectedErr)
			if err == nil {
				suite.Equal(tc.expectedRes, res)
			}
		})
	}
}

func (suite *grpcQueryTestSuite) TestQueryTotalSupply() {
	testCases := []struct {
		name          string
		setup         func()
		expectedTotal sdk.Int
		expectedErr   error
	}{
		{
			name:          "no liquid kava means no tvl",
			setup:         func() {},
			expectedTotal: sdk.ZeroInt(),
			expectedErr:   nil,
		},
		{
			name: "returns TVL from one bkava denom",
			setup: func() {
				initBalance := suite.NewBondCoin(i(1e9))
				valAcc := suite.CreateAccount(sdk.NewCoins(initBalance), 0)
				delAcc := suite.CreateAccount(sdk.NewCoins(initBalance), 1)

				suite.CreateNewUnbondedValidator(valAcc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateDelegation(valAcc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount)
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper) // bond the validator

				_, err := suite.Keeper.MintDerivative(
					suite.Ctx,
					delAcc.GetAddress(),
					valAcc.GetAddress().Bytes(),
					initBalance,
				)
				suite.Require().NoError(err)
			},
			expectedTotal: i(1e9),
			expectedErr:   nil,
		},
		{
			name: "returns TVL from multiple bkava denoms",
			setup: func() {
				initBalance := suite.NewBondCoin(i(1e9))
				val1Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 0)
				val2Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 1)
				delAcc := suite.CreateAccount(sdk.NewCoins(initBalance.Add(initBalance)), 2)

				suite.CreateNewUnbondedValidator(val1Acc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateDelegation(val1Acc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount)
				suite.CreateNewUnbondedValidator(val2Acc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateDelegation(val2Acc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount)
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper) // bond the validator

				_, err := suite.Keeper.MintDerivative(suite.Ctx, delAcc.GetAddress(), val1Acc.GetAddress().Bytes(), initBalance)
				suite.Require().NoError(err)
				_, err = suite.Keeper.MintDerivative(suite.Ctx, delAcc.GetAddress(), val2Acc.GetAddress().Bytes(), initBalance)
				suite.Require().NoError(err)
			},
			expectedTotal: i(2e9),
			expectedErr:   nil,
		},
		{
			name: "returns TVL from multiple delegators",
			setup: func() {
				initBalance := suite.NewBondCoin(i(1e9))
				valAcc := suite.CreateAccount(sdk.NewCoins(initBalance), 0)
				del1Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 1)
				del2Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 2)

				suite.CreateNewUnbondedValidator(valAcc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateDelegation(valAcc.GetAddress().Bytes(), del1Acc.GetAddress(), initBalance.Amount)
				suite.CreateDelegation(valAcc.GetAddress().Bytes(), del2Acc.GetAddress(), initBalance.Amount)
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper) // bond the validator

				_, err := suite.Keeper.MintDerivative(suite.Ctx, del1Acc.GetAddress(), valAcc.GetAddress().Bytes(), initBalance)
				suite.Require().NoError(err)
				_, err = suite.Keeper.MintDerivative(suite.Ctx, del2Acc.GetAddress(), valAcc.GetAddress().Bytes(), initBalance)
				suite.Require().NoError(err)
			},
			expectedTotal: i(2e9),
			expectedErr:   nil,
		},
		{
			name: "handles calculating tvl after slashing",
			setup: func() {
				initBalance := suite.NewBondCoin(i(1e9))
				val1Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 0)
				val2Acc := suite.CreateAccount(sdk.NewCoins(initBalance), 1)
				delAcc := suite.CreateAccount(sdk.NewCoins(initBalance.Add(initBalance)), 2)

				suite.CreateNewUnbondedValidator(val1Acc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateDelegation(val1Acc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount)
				suite.CreateNewUnbondedValidator(val2Acc.GetAddress().Bytes(), initBalance.Amount)
				suite.CreateDelegation(val2Acc.GetAddress().Bytes(), delAcc.GetAddress(), initBalance.Amount)
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper) // bond the validator

				_, err := suite.Keeper.MintDerivative(suite.Ctx, delAcc.GetAddress(), val1Acc.GetAddress().Bytes(), initBalance)
				suite.Require().NoError(err)
				_, err = suite.Keeper.MintDerivative(suite.Ctx, delAcc.GetAddress(), val2Acc.GetAddress().Bytes(), initBalance)
				suite.Require().NoError(err)

				suite.SlashValidator(val2Acc.GetAddress().Bytes(), d("0.1"))
			},
			// delegation + (delegation * 90%)
			expectedTotal: i(1e9).Add(i(1e9).MulRaw(90).QuoRaw(100)),
			expectedErr:   nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()

			res, err := suite.queryClient.TotalSupply(
				context.Background(),
				&types.QueryTotalSupplyRequest{},
			)

			suite.ErrorIs(err, tc.expectedErr)
			if err == nil {
				suite.Equal(tc.expectedTotal, res.Result[0].Amount)
			}
		})
	}
}
