package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type ParamsTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	addrs  []sdk.AccAddress
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *ParamsTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(10)
	tApp.InitializeFromGenesisStates(NewBep3GenStateMulti(addrs[0]))
	suite.keeper = tApp.GetBep3Keeper()
	suite.ctx = ctx
	suite.addrs = addrs
}

func (suite *ParamsTestSuite) TestGetSetBnbDeputyAddress() {
	params := suite.keeper.GetParams(suite.ctx)
	params.BnbDeputyAddress = suite.addrs[1]
	suite.NotPanics(func() { suite.keeper.SetParams(suite.ctx, params) })

	params = suite.keeper.GetParams(suite.ctx)
	suite.Equal(suite.addrs[1], params.BnbDeputyAddress)
	addr := suite.keeper.GetBnbDeputyAddress(suite.ctx)
	suite.Equal(suite.addrs[1], addr)
}

func (suite *ParamsTestSuite) TestGetMaxBlockLock() {
	params := suite.keeper.GetParams(suite.ctx)
	maxBlockLock := params.MaxBlockLock

	res := suite.keeper.GetMaxBlockLock(suite.ctx)
	suite.Equal(maxBlockLock, res)
}

func (suite *ParamsTestSuite) TestGetMinBlockLock() {
	params := suite.keeper.GetParams(suite.ctx)
	minBlockLock := params.MinBlockLock

	res := suite.keeper.GetMinBlockLock(suite.ctx)
	suite.Equal(minBlockLock, res)
}

func (suite *ParamsTestSuite) TestGetAssets() {
	params := suite.keeper.GetParams(suite.ctx)
	assets := params.SupportedAssets

	res, found := suite.keeper.GetAssets(suite.ctx)
	suite.True(found)
	suite.Equal(assets, res)
}

func (suite *ParamsTestSuite) TestGetAssetByDenom() {
	params := suite.keeper.GetParams(suite.ctx)
	asset := params.SupportedAssets[0]

	res, found := suite.keeper.GetAssetByDenom(suite.ctx, asset.Denom)
	suite.True(found)
	suite.Equal(asset, res)
}

func (suite *ParamsTestSuite) TestGetAssetByCoinID() {
	params := suite.keeper.GetParams(suite.ctx)
	asset := params.SupportedAssets[0]

	res, found := suite.keeper.GetAssetByCoinID(suite.ctx, asset.CoinID)
	suite.True(found)
	suite.Equal(asset, res)
}

func (suite *AssetTestSuite) TestValidateLiveAsset() {
	type args struct {
		coin sdk.Coin
	}
	testCases := []struct {
		name          string
		args          args
		expectedError sdk.CodeType
		expectPass    bool
	}{
		{
			"normal",
			args{
				coin: c("bnb", 1),
			},
			sdk.CodeType(0),
			true,
		},
		{
			"asset not supported",
			args{
				coin: c("bad", 1),
			},
			types.CodeAssetNotSupported,
			false,
		},
		{
			"asset not active",
			args{
				coin: c("inc", 1),
			},
			types.CodeAssetNotActive,
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			err := suite.keeper.ValidateLiveAsset(suite.ctx, tc.args.coin)

			if tc.expectPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
				suite.Equal(tc.expectedError, err.Result().Code)
			}
		})
	}
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
