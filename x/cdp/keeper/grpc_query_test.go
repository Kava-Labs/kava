package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
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
	now         time.Time
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.tApp = app.NewTestApp()
	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{Height: 1, Time: time.Now().UTC(), ChainID: app.TestChainID})
	suite.tApp.InitDefaultGenesis(
		suite.ctx,
		NewPricefeedGenStateMulti(suite.tApp.AppCodec()),
		NewCDPGenStateMulti(suite.tApp.AppCodec()),
	)
	suite.keeper = suite.tApp.GetCDPKeeper()
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	suite.addrs = addrs

	suite.now = time.Now().UTC()
}

func (suite *grpcQueryTestSuite) addCdp() {
	ak := suite.tApp.GetAccountKeeper()
	pk := suite.tApp.GetPriceFeedKeeper()

	acc := ak.NewAccountWithAddress(suite.ctx, suite.addrs[0])
	err := suite.tApp.FundAccount(suite.ctx, acc.GetAddress(), cs(c("xrp", 200000000), c("btc", 500000000)))
	suite.NoError(err)

	ak.SetAccount(suite.ctx, acc)

	err = pk.SetCurrentPrices(suite.ctx, "xrp:usd")
	suite.NoError(err)

	ok := suite.keeper.UpdatePricefeedStatus(suite.ctx, "xrp:usd")
	suite.True(ok)

	err = suite.keeper.AddCdp(suite.ctx, suite.addrs[0], c("xrp", 100000000), c("usdx", 10000000), "xrp-a")
	suite.NoError(err)

	id := suite.keeper.GetNextCdpID(suite.ctx)
	suite.Equal(uint64(2), id)

	tp := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp-a", "usdx")
	suite.Equal(i(10000000), tp)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryParams() {
	res, err := suite.queryServer.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	var expected types.GenesisState
	defaultCdpState := NewCDPGenStateMulti(suite.tApp.AppCodec())
	suite.tApp.AppCodec().MustUnmarshalJSON(defaultCdpState[types.ModuleName], &expected)

	suite.Equal(expected.Params, res.Params, "params should equal test genesis state")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryParams_Default() {
	suite.keeper.SetParams(suite.ctx, types.DefaultParams())

	res, err := suite.queryServer.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Empty(res.Params.CollateralParams)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryAccounts() {
	res, err := suite.queryServer.Accounts(sdk.WrapSDKContext(suite.ctx), &types.QueryAccountsRequest{})
	suite.Require().NoError(err)

	ak := suite.tApp.GetAccountKeeper()
	acc := ak.GetModuleAccount(suite.ctx, types.ModuleName)
	liquidator := ak.GetModuleAccount(suite.ctx, types.LiquidatorMacc)

	suite.Len(res.Accounts, 2)
	suite.Equal(acc, &res.Accounts[0], "accounts should include module account")
	suite.Equal(liquidator, &res.Accounts[1], "accounts should include liquidator account")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryTotalPrincipal() {
	suite.addCdp()

	res, err := suite.queryServer.TotalPrincipal(sdk.WrapSDKContext(suite.ctx), &types.QueryTotalPrincipalRequest{})
	suite.Require().NoError(err)

	suite.Len(res.TotalPrincipal, 4, "total principal should include all collateral params")

	suite.Contains(res.TotalPrincipal, types.TotalPrincipal{
		CollateralType: "xrp-a",
		Amount:         sdk.NewCoin("usdx", sdkmath.NewInt(10000000)),
	}, "total principals should include added cdp")
	suite.Contains(res.TotalPrincipal, types.TotalPrincipal{
		CollateralType: "busd-a",
		Amount:         sdk.NewCoin("usdx", sdkmath.NewInt(0)),
	}, "total busd principal should be 0")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryTotalCollateral() {
	suite.addCdp()

	res, err := suite.queryServer.TotalCollateral(sdk.WrapSDKContext(suite.ctx), &types.QueryTotalCollateralRequest{})
	suite.Require().NoError(err)

	suite.Len(res.TotalCollateral, 4, "total collateral should include all collateral params")
	suite.Contains(res.TotalCollateral, types.TotalCollateral{
		CollateralType: "xrp-a",
		Amount:         sdk.NewCoin("xrp", sdkmath.NewInt(100000000)),
	}, "total collaterals should include added cdp")
	suite.Contains(res.TotalCollateral, types.TotalCollateral{
		CollateralType: "busd-a",
		Amount:         sdk.NewCoin("busd", sdkmath.NewInt(0)),
	}, "busd total collateral should be 0")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryCdps() {
	suite.addCdp()

	res, err := suite.queryServer.Cdps(sdk.WrapSDKContext(suite.ctx), &types.QueryCdpsRequest{
		CollateralType: "xrp-a",
		Pagination: &query.PageRequest{
			Limit: 100,
		},
	})
	suite.Require().NoError(err)

	suite.Len(res.Cdps, 1)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryCdps_InvalidCollateralType() {
	suite.addCdp()

	_, err := suite.queryServer.Cdps(sdk.WrapSDKContext(suite.ctx), &types.QueryCdpsRequest{
		CollateralType: "kava-a",
	})
	suite.Require().Error(err)
	suite.Require().Equal("rpc error: code = InvalidArgument desc = invalid collateral type", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcQueryCdp() {
	suite.addCdp()

	tests := []struct {
		giveName     string
		giveRequest  types.QueryCdpRequest
		wantAccepted bool
		wantErr      string
	}{
		{
			"valid",
			types.QueryCdpRequest{
				CollateralType: "xrp-a",
				Owner:          suite.addrs[0].String(),
			},
			true,
			"",
		},
		{
			"invalid collateral",
			types.QueryCdpRequest{
				CollateralType: "kava-a",
				Owner:          suite.addrs[0].String(),
			},
			false,
			"kava-a: invalid collateral for input collateral type",
		},
		{
			"missing owner",
			types.QueryCdpRequest{
				CollateralType: "xrp-a",
			},
			false,
			"rpc error: code = InvalidArgument desc = invalid address",
		},
		{
			"invalid owner",
			types.QueryCdpRequest{
				CollateralType: "xrp-a",
				Owner:          "invalid addr",
			},
			false,
			"rpc error: code = InvalidArgument desc = invalid address",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.giveName, func() {
			_, err := suite.queryServer.Cdp(sdk.WrapSDKContext(suite.ctx), &tt.giveRequest)

			if tt.wantAccepted {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().Equal(tt.wantErr, err.Error())
			}
		})
	}
}

func (suite *grpcQueryTestSuite) TestGrpcQueryDeposits() {
	suite.addCdp()

	tests := []struct {
		giveName            string
		giveRequest         *types.QueryDepositsRequest
		wantContainsDeposit *types.Deposit
		wantShouldErr       bool
		wantErr             string
	}{
		{
			"valid",
			&types.QueryDepositsRequest{
				CollateralType: "xrp-a",
				Owner:          suite.addrs[0].String(),
			},
			&types.Deposit{
				CdpID:     1,
				Depositor: suite.addrs[0],
				Amount:    sdk.NewCoin("xrp", sdkmath.NewInt(100000000)),
			},
			false,
			"",
		},
		{
			"invalid collateral type",
			&types.QueryDepositsRequest{
				CollateralType: "kava-a",
				Owner:          suite.addrs[0].String(),
			},
			nil,
			true,
			"kava-a: invalid collateral for input collateral type",
		},
		{
			"missing owner",
			&types.QueryDepositsRequest{
				CollateralType: "xrp-a",
			},
			nil,
			true,
			"rpc error: code = InvalidArgument desc = invalid address",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.giveName, func() {
			res, err := suite.queryServer.Deposits(sdk.WrapSDKContext(suite.ctx), tt.giveRequest)

			if tt.wantShouldErr {
				suite.Error(err)
				suite.Equal(tt.wantErr, err.Error())
			} else {
				suite.NoError(err)
				suite.Contains(res.Deposits, *tt.wantContainsDeposit)
			}
		})
	}
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
