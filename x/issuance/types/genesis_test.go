package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/issuance/types"
)

type GenesisTestSuite struct {
	suite.Suite

	addrs []string
}

func (suite *GenesisTestSuite) SetupTest() {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	var strAddrs []string
	for _, addr := range addrs {
		strAddrs = append(strAddrs, addr.String())
	}
	suite.addrs = strAddrs
}

func (suite *GenesisTestSuite) TestValidate() {
	type args struct {
		assets   []types.Asset
		supplies []types.AssetSupply
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
			"default",
			args{
				assets:   types.DefaultAssets,
				supplies: types.DefaultSupplies,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"with asset",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{types.NewAssetSupply(sdk.NewCoin("usdtoken", sdkmath.NewInt(1000000)), time.Hour)},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"with asset rate limit",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(true, sdkmath.NewInt(1000000000), time.Hour*24)),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"with multiple assets",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
					types.NewAsset(suite.addrs[0], "pegtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"blocked owner",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[0]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "asset owner cannot be blocked",
			},
		},
		{
			"empty owner",
			args{
				assets: []types.Asset{
					types.NewAsset("", "usdtoken", []string{suite.addrs[0]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "owner must not be empty",
			},
		},
		{
			"empty blocked address",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{""}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "blocked address must not be empty",
			},
		},
		{
			"invalid denom",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "USD2T ", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
		{
			"duplicate denom",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
					types.NewAsset(suite.addrs[1], "usdtoken", []string{}, true, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "duplicate asset denoms",
			},
		},
		{
			"duplicate asset",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "duplicate asset denoms",
			},
		},
		{
			"invalid block list",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, false, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				supplies: []types.AssetSupply{types.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour)},
			},
			errArgs{
				expectPass: false,
				contains:   "blocked-list should be empty",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			gs := types.NewGenesisState(types.NewParams(tc.args.assets), tc.args.supplies)
			err := gs.Validate()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
