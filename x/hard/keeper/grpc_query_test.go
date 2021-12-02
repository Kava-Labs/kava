package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
	"github.com/stretchr/testify/suite"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
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
	suite.tApp.InitializeFromGenesisStates(
		NewHARDGenState(suite.tApp.AppCodec()),
	)
	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	suite.keeper = suite.tApp.GetHardKeeper()
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)

	_, addrs := app.GeneratePrivKeyAddressPairs(5)

	app.NewFundedGenStateWithSameCoins(
		suite.tApp.AppCodec(),
		cs(
			c("bnb", 1000*100000000),
			c("busd", 1000*100000000),
		),
		addrs,
	)
	suite.addrs = addrs
}

func (suite *grpcQueryTestSuite) TestGrpcQueryParams() {
	res, err := suite.queryServer.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	var expected types.GenesisState
	defaultHARDState := NewHARDGenState(suite.tApp.AppCodec())
	suite.tApp.AppCodec().MustUnmarshalJSON(defaultHARDState[types.ModuleName], &expected)

	suite.Equal(expected.Params, res.Params, "params should equal test genesis state")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryAccounts() {
	res, err := suite.queryServer.Accounts(sdk.WrapSDKContext(suite.ctx), &types.QueryAccountsRequest{})
	suite.Require().NoError(err)

	ak := suite.tApp.GetAccountKeeper()
	acc := ak.GetModuleAccount(suite.ctx, types.ModuleName)

	suite.Len(res.Accounts, 1)
	suite.Equal(acc, &res.Accounts[0], "accounts should include module account")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryAccounts_InvalidName() {
	_, err := suite.queryServer.Accounts(sdk.WrapSDKContext(suite.ctx), &types.QueryAccountsRequest{
		Name: "boo",
	})
	suite.Require().Error(err)
	suite.Require().Equal("rpc error: code = InvalidArgument desc = invalid account name", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcQueryDeposits_EmptyResponse() {
	res, err := suite.queryServer.Deposits(sdk.WrapSDKContext(suite.ctx), &types.QueryDepositsRequest{})
	suite.Require().NoError(err)
	suite.Require().Empty(res)
}

func (suite *grpcQueryTestSuite) addDeposits() {
	deposits := []struct {
		Address sdk.AccAddress
		Coins   sdk.Coins
	}{
		{
			suite.addrs[0],
			sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
		},
		{
			suite.addrs[1],
			sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
		},
		{
			suite.addrs[0],
			sdk.NewCoins(sdk.NewCoin("busd", sdk.NewInt(400))),
		},
		{
			suite.addrs[0],
			sdk.NewCoins(sdk.NewCoin("bun", sdk.NewInt(800))),
		},
	}

	for _, dep := range deposits {
		suite.NotPanics(func() {
			err := suite.keeper.Deposit(suite.ctx, dep.Address, dep.Coins)
			suite.Require().NoError(err)
		})
	}
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
			4,
			false,
			"",
		},
		{
			"owner",
			&types.QueryDepositsRequest{
				Owner: sdk.AccAddress("test").String(),
			},
			// Excludes the second address
			3,
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
				Owner: sdk.AccAddress("test").String(),
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
				Owner: sdk.AccAddress("test").String(),
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
func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
