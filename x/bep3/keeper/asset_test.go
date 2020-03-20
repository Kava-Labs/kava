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

type AssetTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

func (suite *AssetTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	// Initialize test app and set context
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	// Set up auth genesis state
	coins := []sdk.Coins{}
	for j := 0; j < 10; j++ {
		coins = append(coins, cs(c("bnb", STARING_BNB_BALANCE)))
	}
	_, addrs := app.GeneratePrivKeyAddressPairs(10)
	authGS := app.NewAuthGenState(addrs, coins)

	// Initialize genesis state
	tApp.InitializeFromGenesisStates(
		authGS,
		NewBep3GenStateMulti(),
	)
	// Load keeper
	keeper := tApp.GetBep3Keeper()

	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
	return
}

func (suite *AssetTestSuite) TestGetAssetSupplyInfo() {
	type args struct {
		denom        string
		inSwapSupply int64
		assetSupply  int64
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"normal",
			args{
				denom:        "bnb",
				inSwapSupply: 10000,
				assetSupply:  50000,
			},
			true,
		},
		{
			"unsupported asset",
			args{
				denom:        "xyz",
				inSwapSupply: 10000,
				assetSupply:  50000,
			},
			false,
		},
	}
	for _, tc := range testCases {
		_, _ = suite.keeper.LoadAssetSupply(suite.ctx, tc.args.denom)

		// Set in swap supply and asset supply
		suite.keeper.IncrementInSwapSupply(suite.ctx, c(tc.args.denom, tc.args.inSwapSupply))
		suite.keeper.IncrementAssetSupply(suite.ctx, c(tc.args.denom, tc.args.assetSupply))

		// Attempt to get information about this asset's supply
		assetSupplyInfo, err := suite.keeper.GetAssetSupplyInfo(suite.ctx, tc.args.denom)
		assetParam, found := suite.keeper.GetAssetByDenom(suite.ctx, tc.args.denom)
		if tc.expectPass {
			suite.NoError(err)
			suite.True(found)

			// Confirm contents
			expectedAssetSupplyInfo := types.AssetSupplyInfo{
				Denom:        tc.args.denom,
				InSwapSupply: tc.args.inSwapSupply,
				AssetSupply:  tc.args.assetSupply,
				SupplyLimit:  assetParam.Limit.Int64(),
			}
			suite.Equal(expectedAssetSupplyInfo, assetSupplyInfo)
		} else {
			suite.Error(err)
			suite.False(found)
		}
	}
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
		err := suite.keeper.ValidateLiveAsset(suite.ctx, tc.args.coin)

		if tc.expectPass {
			suite.NoError(err)
		} else {
			suite.Error(err)
			suite.Equal(tc.expectedError, err.Result().Code)
		}
	}
}

func (suite *AssetTestSuite) TestIncrementAssetSupply() {
	// Initialize asset 'bnb' asset supply
	_, _ = suite.keeper.LoadAssetSupply(suite.ctx, "bnb")

	type args struct {
		coin sdk.Coin
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"normal",
			args{
				coin: c("bnb", 10000),
			},
			true,
		},
	}

	for _, tc := range testCases {
		assetSupplyPre, _ := suite.keeper.GetAssetSupply(suite.ctx, []byte(tc.args.coin.Denom))
		suite.keeper.IncrementAssetSupply(suite.ctx, tc.args.coin)
		assetSupplyPost, _ := suite.keeper.GetAssetSupply(suite.ctx, []byte(tc.args.coin.Denom))

		if tc.expectPass {
			// Check asset supply changed
			suite.Equal(assetSupplyPre.Add(tc.args.coin), assetSupplyPost)
		} else {
			// Check asset supply hasn't changed
			suite.Equal(assetSupplyPre, assetSupplyPost)
		}
	}
}

func (suite *AssetTestSuite) TestDecrementAssetSupply() {
	// Initialize asset 'bnb' asset supply
	_, _ = suite.keeper.LoadAssetSupply(suite.ctx, "bnb")

	type args struct {
		coin sdk.Coin
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"normal",
			args{
				coin: c("bnb", 450),
			},
			true,
		},
		{
			"negative asset supply",
			args{
				coin: c("bnb", 1450),
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.keeper.IncrementAssetSupply(suite.ctx, c("bnb", 1000))
		assetSupplyPre, _ := suite.keeper.GetAssetSupply(suite.ctx, []byte(tc.args.coin.Denom))
		err := suite.keeper.DecrementAssetSupply(suite.ctx, tc.args.coin)
		assetSupplyPost, _ := suite.keeper.GetAssetSupply(suite.ctx, []byte(tc.args.coin.Denom))

		if tc.expectPass {
			suite.Nil(err)
			// Check asset supply changed
			suite.Equal(assetSupplyPre.Sub(tc.args.coin), assetSupplyPost)
		} else {
			suite.NotNil(err)
			// Check asset supply hasn't changed
			suite.Equal(assetSupplyPre, assetSupplyPost)
		}
	}
}

func (suite *AssetTestSuite) TestIncrementInSwapSupply() {
	// Initialize asset 'bnb' in swap supply
	_, _ = suite.keeper.LoadAssetSupply(suite.ctx, "bnb")

	type args struct {
		coin sdk.Coin
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"normal",
			args{
				coin: c("bnb", 10000),
			},
			true,
		},
	}

	for _, tc := range testCases {
		inSwapSupplyPre, _ := suite.keeper.GetInSwapSupply(suite.ctx, []byte(tc.args.coin.Denom))
		suite.keeper.IncrementInSwapSupply(suite.ctx, tc.args.coin)
		inSwapSupplyPost, _ := suite.keeper.GetInSwapSupply(suite.ctx, []byte(tc.args.coin.Denom))

		if tc.expectPass {
			// Check asset supply changed
			suite.Equal(inSwapSupplyPre.Add(tc.args.coin), inSwapSupplyPost)
		} else {
			// Check asset supply hasn't changed
			suite.Equal(inSwapSupplyPre, inSwapSupplyPost)
		}
	}
}

func (suite *AssetTestSuite) TestDecrementInSwapSupply() {
	// Initialize asset 'bnb' in swap supply
	_, _ = suite.keeper.LoadAssetSupply(suite.ctx, "bnb")

	type args struct {
		coin sdk.Coin
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"normal",
			args{
				coin: c("bnb", 500),
			},
			true,
		},
		{
			"negative in swap supply",
			args{
				coin: c("bnb", 5000),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.keeper.IncrementInSwapSupply(suite.ctx, c("bnb", 1000))
		inSwapSupplyPre, _ := suite.keeper.GetInSwapSupply(suite.ctx, []byte(tc.args.coin.Denom))
		err := suite.keeper.DecrementInSwapSupply(suite.ctx, tc.args.coin)
		inSwapSupplyPost, _ := suite.keeper.GetInSwapSupply(suite.ctx, []byte(tc.args.coin.Denom))

		if tc.expectPass {
			suite.Nil(err)
			// Check asset supply changed
			suite.Equal(inSwapSupplyPre.Sub(tc.args.coin), inSwapSupplyPost)
		} else {
			suite.NotNil(err)
			// Check asset supply hasn't changed
			suite.Equal(inSwapSupplyPre, inSwapSupplyPost)
		}
	}
}

func (suite *AssetTestSuite) TestValidateCreateSwapAgainstSupplyLimit() {
	// Initialize asset 'bnb' asset supply and in swap supply
	_, _ = suite.keeper.LoadAssetSupply(suite.ctx, "bnb")

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
				coin: c("bnb", 500),
			},
			sdk.CodeType(0),
			true,
		},
		{
			"above asset supply limit",
			args{
				coin: c("bnb", BNB_SUPPLY_LIMIT.Int64()+1),
			},
			types.CodeAboveTotalAssetSupplyLimit,
			false,
		},
	}

	for _, tc := range testCases {
		err := suite.keeper.ValidateCreateSwapAgainstSupplyLimit(suite.ctx, tc.args.coin)

		if tc.expectPass {
			suite.NoError(err)
		} else {
			suite.Error(err)
			suite.Equal(tc.expectedError, err.Result().Code)
		}
	}
}

func TestAssetTestSuite(t *testing.T) {
	suite.Run(t, new(AssetTestSuite))
}
