package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/issuance/types"
)

func (suite *KeeperTestSuite) TestIncrementCurrentAssetSupply() {
	type args struct {
		assets   types.Assets
		supplies types.AssetSupplies
		coin     sdk.Coin
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"valid supply increase",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(true, sdk.NewInt(10000000000), time.Hour*24)),
				},
				supplies: types.AssetSupplies{
					types.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour),
				},
				coin: sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"over limit increase",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(true, sdk.NewInt(10000000000), time.Hour*24)),
				},
				supplies: types.AssetSupplies{
					types.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour),
				},
				coin: sdk.NewCoin("usdtoken", sdk.NewInt(10000000001)),
			},
			errArgs{
				expectPass: false,
				contains:   "asset supply over limit",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)
			for _, supply := range tc.args.supplies {
				suite.keeper.SetAssetSupply(suite.ctx, supply, supply.GetDenom())
			}
			err := suite.keeper.IncrementCurrentAssetSupply(suite.ctx, tc.args.coin)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				for _, expectedSupply := range tc.args.supplies {
					expectedSupply.CurrentSupply = expectedSupply.CurrentSupply.Add(tc.args.coin)
					actualSupply, found := suite.keeper.GetAssetSupply(suite.ctx, expectedSupply.GetDenom())
					suite.Require().True(found)
					suite.Require().Equal(expectedSupply, actualSupply, tc.name)
				}
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
