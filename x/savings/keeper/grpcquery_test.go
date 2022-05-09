package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/savings/keeper"
	"github.com/kava-labs/kava/x/savings/types"
)

type grpcQueryTestSuite struct {
	suite.Suite

	tApp        app.TestApp
	ctx         sdk.Context
	keeper      keeper.Keeper
	queryServer types.QueryServer
	addrs       []sdk.AccAddress
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.tApp = app.NewTestApp()
	_, addrs := app.GeneratePrivKeyAddressPairs(2)

	suite.addrs = addrs

	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	suite.keeper = suite.tApp.GetSavingsKeeper()
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)

	err := suite.tApp.FundModuleAccount(
		suite.ctx,
		types.ModuleAccountName,
		cs(
			c("usdx", 10000000000),
			c("busd", 10000000000),
		),
	)
	suite.Require().NoError(err)

	savingsGenesis := types.GenesisState{
		Params: types.NewParams([]string{"bnb", "busd"}),
	}
	savingsGenState := app.GenesisState{types.ModuleName: suite.tApp.AppCodec().MustMarshalJSON(&savingsGenesis)}

	suite.tApp.InitializeFromGenesisStates(
		savingsGenState,
		app.NewFundedGenStateWithSameCoins(
			suite.tApp.AppCodec(),
			cs(
				c("bnb", 10000000000),
				c("busd", 20000000000),
			),
			addrs,
		),
	)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryParams() {
	res, err := suite.queryServer.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	var expected types.GenesisState
	savingsGenesis := types.GenesisState{
		Params: types.NewParams([]string{"bnb", "busd"}),
	}
	savingsGenState := app.GenesisState{types.ModuleName: suite.tApp.AppCodec().MustMarshalJSON(&savingsGenesis)}
	suite.tApp.AppCodec().MustUnmarshalJSON(savingsGenState[types.ModuleName], &expected)

	suite.Equal(expected.Params, res.Params, "params should equal test genesis state")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryDeposits() {
	suite.addDeposits()

	tests := []struct {
		giveName          string
		giveRequest       *types.QueryDepositsRequest
		wantDepositCounts int
		shouldError       bool
		errorSubstr       string
	}{
		{
			"empty query",
			&types.QueryDepositsRequest{},
			2,
			false,
			"",
		},
		{
			"owner",
			&types.QueryDepositsRequest{
				Owner: suite.addrs[0].String(),
			},
			// Excludes the second address
			1,
			false,
			"",
		},
		{
			"invalid owner",
			&types.QueryDepositsRequest{
				Owner: "invalid address",
			},
			// No deposits
			0,
			true,
			"decoding bech32 failed",
		},
		{
			"owner and denom",
			&types.QueryDepositsRequest{
				Owner: suite.addrs[0].String(),
				Denom: "bnb",
			},
			// Only the first one
			1,
			false,
			"",
		},
		{
			"owner and invalid denom empty response",
			&types.QueryDepositsRequest{
				Owner: suite.addrs[0].String(),
				Denom: "invalid denom",
			},
			0,
			false,
			"",
		},
		{
			"denom",
			&types.QueryDepositsRequest{
				Denom: "bnb",
			},
			2,
			false,
			"",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.giveName, func() {
			res, err := suite.queryServer.Deposits(sdk.WrapSDKContext(suite.ctx), tt.giveRequest)

			if tt.shouldError {
				suite.Error(err)
				suite.Contains(err.Error(), tt.errorSubstr)
			} else {
				suite.NoError(err)
				suite.Equal(tt.wantDepositCounts, len(res.Deposits))
			}
		})
	}
}

func (suite *grpcQueryTestSuite) addDeposits() {
	deposits := []struct {
		Address sdk.AccAddress
		Coins   sdk.Coins
	}{
		{
			suite.addrs[0],
			cs(c("bnb", 100000000)),
		},
		{
			suite.addrs[1],
			cs(c("bnb", 20000000)),
		},
		{
			suite.addrs[0],
			cs(c("busd", 20000000)),
		},
		{
			suite.addrs[0],
			cs(c("busd", 8000000)),
		},
	}

	for _, dep := range deposits {
		suite.NotPanics(func() {
			err := suite.keeper.Deposit(suite.ctx, dep.Address, dep.Coins)
			suite.Require().NoError(err)
		})
	}
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
