package keeper_test

import (
	"context"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/kava-labs/kava/x/precisebank/types"
)

type grpcQueryTestSuite struct {
	testutil.Suite

	queryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.Suite.SetupTest()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestQueryTotalFractionalBalance() {
	testCases := []struct {
		name         string
		giveBalances []sdkmath.Int
	}{
		{
			"empty",
			[]sdkmath.Int{},
		},
		{
			"min amount",
			[]sdkmath.Int{
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"exceeds conversion factor",
			[]sdkmath.Int{
				// 4 accounts * 0.5 == 2
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			total := sdk.NewCoin(types.ExtendedCoinDenom, sdkmath.ZeroInt())
			for i, balance := range tc.giveBalances {
				addr := sdk.AccAddress([]byte(strconv.Itoa(i)))
				suite.Keeper.SetFractionalBalance(suite.Ctx, addr, balance)

				total.Amount = total.Amount.Add(balance)
			}

			res, err := suite.queryClient.TotalFractionalBalances(
				context.Background(),
				&types.QueryTotalFractionalBalancesRequest{},
			)
			suite.Require().NoError(err)

			suite.Require().Equal(total, res.Total)
		})
	}
}

func (suite *grpcQueryTestSuite) TestQueryRemainder() {
	res, err := suite.queryClient.Remainder(
		context.Background(),
		&types.QueryRemainderRequest{},
	)
	suite.Require().NoError(err)

	expRemainder := sdk.NewCoin(types.ExtendedCoinDenom, sdkmath.ZeroInt())
	suite.Require().Equal(expRemainder, res.Remainder)

	// Mint fractional coins to create non-zero remainder

	pbk := suite.App.GetPrecisebankKeeper()

	coin := sdk.NewCoin(types.ExtendedCoinDenom, sdkmath.OneInt())
	err = pbk.MintCoins(
		suite.Ctx,
		minttypes.ModuleName,
		sdk.NewCoins(coin),
	)
	suite.Require().NoError(err)

	res, err = suite.queryClient.Remainder(
		context.Background(),
		&types.QueryRemainderRequest{},
	)
	suite.Require().NoError(err)

	expRemainder.Amount = types.ConversionFactor().Sub(coin.Amount)
	suite.Require().Equal(expRemainder, res.Remainder)
}

func (suite *grpcQueryTestSuite) TestQueryFractionalBalance() {
	testCases := []struct {
		name        string
		giveBalance sdkmath.Int
	}{
		{
			"zero",
			sdkmath.ZeroInt(),
		},
		{
			"min amount",
			sdkmath.OneInt(),
		},
		{
			"max amount",
			types.ConversionFactor().SubRaw(1),
		},
		{
			"multiple integer amounts, 0 fractional",
			types.ConversionFactor().MulRaw(5),
		},
		{
			"multiple integer amounts, non-zero fractional",
			types.ConversionFactor().MulRaw(5).Add(types.ConversionFactor().QuoRaw(2)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			addr := sdk.AccAddress([]byte("test"))

			coin := sdk.NewCoin(types.ExtendedCoinDenom, tc.giveBalance)
			suite.MintToAccount(addr, sdk.NewCoins(coin))

			res, err := suite.queryClient.FractionalBalance(
				context.Background(),
				&types.QueryFractionalBalanceRequest{
					Address: addr.String(),
				},
			)
			suite.Require().NoError(err)

			// Only fractional amount, even if minted more than conversion factor
			expAmount := tc.giveBalance.Mod(types.ConversionFactor())
			expFractionalBalance := sdk.NewCoin(types.ExtendedCoinDenom, expAmount)
			suite.Require().Equal(expFractionalBalance, res.FractionalBalance)
		})
	}
}
