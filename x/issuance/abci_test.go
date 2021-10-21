package issuance_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/issuance"
)

// Test suite used for all keeper tests
type ABCITestSuite struct {
	suite.Suite

	keeper     issuance.Keeper
	app        app.TestApp
	ctx        sdk.Context
	addrs      []sdk.AccAddress
	modAccount sdk.AccAddress
	blockTime  time.Time
}

// The default state used by each test
func (suite *ABCITestSuite) SetupTest() {
	tApp := app.NewTestApp()
	blockTime := tmtime.Now()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: blockTime})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	keeper := tApp.GetIssuanceKeeper()
	modAccount, err := sdk.AccAddressFromBech32("kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w")
	suite.Require().NoError(err)
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
	suite.modAccount = modAccount
	suite.blockTime = blockTime
}

func (suite *ABCITestSuite) TestRateLimitingTimePassage() {
	type args struct {
		assets         issuance.Assets
		supplies       issuance.AssetSupplies
		blockTimes     []time.Duration
		expectedSupply issuance.AssetSupply
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			"time passage same period",
			args{
				assets: issuance.Assets{
					issuance.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, issuance.NewRateLimit(true, sdk.NewInt(10000000000), time.Hour*24)),
				},
				supplies: issuance.AssetSupplies{
					issuance.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour),
				},
				blockTimes:     []time.Duration{time.Hour},
				expectedSupply: issuance.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour*2),
			},
		},
		{
			"time passage new period",
			args{
				assets: issuance.Assets{
					issuance.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, issuance.NewRateLimit(true, sdk.NewInt(10000000000), time.Hour*24)),
				},
				supplies: issuance.AssetSupplies{
					issuance.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour),
				},
				blockTimes:     []time.Duration{time.Hour * 12, time.Hour * 12},
				expectedSupply: issuance.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Duration(0)),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := issuance.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)
			for _, supply := range tc.args.supplies {
				suite.keeper.SetAssetSupply(suite.ctx, supply, supply.GetDenom())
			}
			suite.keeper.SetPreviousBlockTime(suite.ctx, suite.blockTime)
			for _, bt := range tc.args.blockTimes {
				nextBlockTime := suite.ctx.BlockTime().Add(bt)
				suite.ctx = suite.ctx.WithBlockTime(nextBlockTime)
				suite.Require().NotPanics(func() {
					issuance.BeginBlocker(suite.ctx, suite.keeper)
				})
			}
			actualSupply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.expectedSupply.GetDenom())
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedSupply, actualSupply)
		})
	}
}

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}
