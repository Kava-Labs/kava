package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
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

func (suite *ParamsTestSuite) TestGetSetAsset() {
	asset, err := suite.keeper.GetAsset(suite.ctx, "bnb")
	suite.Require().NoError(err)
	suite.NotPanics(func() { suite.keeper.SetAsset(suite.ctx, asset) })
	_, err = suite.keeper.GetAsset(suite.ctx, "dne")
	suite.Require().Error(err)

	_, err = suite.keeper.GetAsset(suite.ctx, "inc")
	suite.Require().NoError(err)
}

func (suite *ParamsTestSuite) TestGetAssets() {
	assets, found := suite.keeper.GetAssets(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(2, len(assets))
}

func (suite *ParamsTestSuite) TestGetSetDeputyAddress() {
	asset, err := suite.keeper.GetAsset(suite.ctx, "bnb")
	suite.Require().NoError(err)
	asset.DeputyAddress = suite.addrs[1]
	suite.NotPanics(func() { suite.keeper.SetAsset(suite.ctx, asset) })

	asset, err = suite.keeper.GetAsset(suite.ctx, "bnb")
	suite.Require().NoError(err)
	suite.Equal(suite.addrs[1], asset.DeputyAddress)
	addr, err := suite.keeper.GetDeputyAddress(suite.ctx, "bnb")
	suite.Require().NoError(err)
	suite.Equal(suite.addrs[1], addr)

}

func (suite *ParamsTestSuite) TestGetDeputyFixedFee() {
	asset, err := suite.keeper.GetAsset(suite.ctx, "bnb")
	suite.Require().NoError(err)
	bnbDeputyFixedFee := asset.FixedFee

	res, err := suite.keeper.GetFixedFee(suite.ctx, asset.Denom)
	suite.Require().NoError(err)
	suite.Equal(bnbDeputyFixedFee, res)
}

func (suite *ParamsTestSuite) TestGetMinMaxSwapAmount() {
	asset, err := suite.keeper.GetAsset(suite.ctx, "bnb")
	suite.Require().NoError(err)
	minAmount := asset.MinSwapAmount

	res, err := suite.keeper.GetMinSwapAmount(suite.ctx, asset.Denom)
	suite.Require().NoError(err)
	suite.Equal(minAmount, res)

	maxAmount := asset.MaxSwapAmount
	res, err = suite.keeper.GetMaxSwapAmount(suite.ctx, asset.Denom)
	suite.Require().NoError(err)
	suite.Equal(maxAmount, res)
}

func (suite *ParamsTestSuite) TestGetMinMaxBlockLock() {
	asset, err := suite.keeper.GetAsset(suite.ctx, "bnb")
	suite.Require().NoError(err)
	minLock := asset.MinBlockLock

	res, err := suite.keeper.GetMinBlockLock(suite.ctx, asset.Denom)
	suite.Require().NoError(err)
	suite.Equal(minLock, res)

	maxLock := asset.MaxBlockLock
	res, err = suite.keeper.GetMaxBlockLock(suite.ctx, asset.Denom)
	suite.Require().NoError(err)
	suite.Equal(maxLock, res)
}

func (suite *ParamsTestSuite) TestGetAssetByCoinID() {
	asset, err := suite.keeper.GetAsset(suite.ctx, "bnb")
	suite.Require().NoError(err)

	res, found := suite.keeper.GetAssetByCoinID(suite.ctx, asset.CoinID)
	suite.True(found)
	suite.Equal(asset, res)
}

func (suite *ParamsTestSuite) TestGetAuthorizedAddresses() {
	deputyAddresses := suite.keeper.GetAuthorizedAddresses(suite.ctx)
	// the test params use the same deputy address for two assets
	expectedAddresses := []sdk.AccAddress{suite.addrs[0]}

	suite.Require().ElementsMatch(expectedAddresses, deputyAddresses)
}

func (suite *AssetTestSuite) TestValidateLiveAsset() {
	type args struct {
		coin sdk.Coin
	}
	testCases := []struct {
		name          string
		args          args
		expectedError error
		expectPass    bool
	}{
		{
			"normal",
			args{
				coin: c("bnb", 1),
			},
			nil,
			true,
		},
		{
			"asset not supported",
			args{
				coin: c("bad", 1),
			},
			types.ErrAssetNotSupported,
			false,
		},
		{
			"asset not active",
			args{
				coin: c("inc", 1),
			},
			types.ErrAssetNotActive,
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			err := suite.keeper.ValidateLiveAsset(suite.ctx, tc.args.coin)

			if tc.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expectedError))
			}
		})
	}
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
