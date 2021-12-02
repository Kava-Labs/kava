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
	_, addrs := app.GeneratePrivKeyAddressPairs(2)

	suite.addrs = addrs

	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	suite.keeper = suite.tApp.GetHardKeeper()
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)

	err := suite.tApp.FundModuleAccount(
		suite.ctx,
		types.ModuleAccountName,
		cs(c("usdx", 10000000000)),
	)
	suite.Require().NoError(err)

	suite.tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(suite.tApp.AppCodec()),
		NewHARDGenState(suite.tApp.AppCodec()),
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

func (suite *grpcQueryTestSuite) addBorrows() {
	borrows := []struct {
		Address sdk.AccAddress
		Coins   sdk.Coins
	}{
		{
			suite.addrs[0],
			cs(c("usdx", 1*10000000)),
		},
		{
			suite.addrs[1],
			cs(c("usdx", 2*10000000)),
		},
		{
			suite.addrs[0],
			cs(c("usdx", 4*10000000)),
		},
		{
			suite.addrs[0],
			cs(c("usdx", 8*10000000)),
		},
	}

	for _, dep := range borrows {
		suite.NotPanics(func() {
			err := suite.keeper.Borrow(suite.ctx, dep.Address, dep.Coins)
			suite.Require().NoErrorf(err, "borrow %s should not error", dep.Coins)
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

		// Unsynced deposits should be the same
		suite.Run(tt.giveName+"_unsynced", func() {
			res, err := suite.queryServer.UnsyncedDeposits(sdk.WrapSDKContext(suite.ctx), &types.QueryUnsyncedDepositsRequest{
				Denom:      tt.giveRequest.Denom,
				Owner:      tt.giveRequest.Owner,
				Pagination: tt.giveRequest.Pagination,
			})

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

func (suite *grpcQueryTestSuite) TestGrpcQueryBorrows() {
	suite.addDeposits()
	suite.addBorrows()

	tests := []struct {
		giveName          string
		giveRequest       *types.QueryBorrowsRequest
		wantDepositCounts int
		shouldError       bool
		errorSubstr       string
	}{
		{
			"empty query",
			&types.QueryBorrowsRequest{},
			2,
			false,
			"",
		},
		{
			"owner",
			&types.QueryBorrowsRequest{
				Owner: suite.addrs[0].String(),
			},
			// Excludes the second address
			1,
			false,
			"",
		},
		{
			"invalid owner",
			&types.QueryBorrowsRequest{
				Owner: "invalid address",
			},
			// No deposits
			0,
			true,
			"decoding bech32 failed",
		},
		{
			"owner and denom",
			&types.QueryBorrowsRequest{
				Owner: suite.addrs[0].String(),
				Denom: "usdx",
			},
			// Only the first one
			1,
			false,
			"",
		},
		{
			"owner and invalid denom empty response",
			&types.QueryBorrowsRequest{
				Owner: suite.addrs[0].String(),
				Denom: "invalid denom",
			},
			0,
			false,
			"",
		},
		{
			"denom",
			&types.QueryBorrowsRequest{
				Denom: "usdx",
			},
			2,
			false,
			"",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.giveName, func() {
			res, err := suite.queryServer.Borrows(sdk.WrapSDKContext(suite.ctx), tt.giveRequest)

			if tt.shouldError {
				suite.Error(err)
				suite.Contains(err.Error(), tt.errorSubstr)
			} else {
				suite.NoError(err)
				suite.Equal(tt.wantDepositCounts, len(res.Borrows))
			}
		})

		// Unsynced deposits should be the same
		suite.Run(tt.giveName+"_unsynced", func() {
			res, err := suite.queryServer.UnsyncedBorrows(sdk.WrapSDKContext(suite.ctx), &types.QueryUnsyncedBorrowsRequest{
				Denom:      tt.giveRequest.Denom,
				Owner:      tt.giveRequest.Owner,
				Pagination: tt.giveRequest.Pagination,
			})

			if tt.shouldError {
				suite.Error(err)
				suite.Contains(err.Error(), tt.errorSubstr)
			} else {
				suite.NoError(err)
				suite.Equal(tt.wantDepositCounts, len(res.Borrows))
			}
		})
	}
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
