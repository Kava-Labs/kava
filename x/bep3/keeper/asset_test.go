package keeper_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

type AssetTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *AssetTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	// Initialize test app and set context
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	// Initialize genesis state
	deputy, _ := sdk.AccAddressFromBech32(TestDeputy)
	tApp.InitializeFromGenesisStates(NewBep3GenStateMulti(deputy))

	keeper := tApp.GetBep3Keeper()
	params := keeper.GetParams(ctx)
	params.AssetParams[0].SupplyLimit.Limit = sdk.NewInt(50)
	params.AssetParams[1].SupplyLimit.Limit = sdk.NewInt(100)
	params.AssetParams[1].SupplyLimit.TimeBasedLimit = sdk.NewInt(15)
	keeper.SetParams(ctx, params)
	// Set asset supply with standard value for testing
	supply := types.NewAssetSupply(c("bnb", 5), c("bnb", 5), c("bnb", 40), c("bnb", 0), time.Duration(0))
	keeper.SetAssetSupply(ctx, supply, supply.IncomingSupply.Denom)

	supply = types.NewAssetSupply(c("inc", 10), c("inc", 5), c("inc", 5), c("inc", 0), time.Duration(0))
	keeper.SetAssetSupply(ctx, supply, supply.IncomingSupply.Denom)
	keeper.SetPreviousBlockTime(ctx, ctx.BlockTime())

	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	return
}

func (suite *AssetTestSuite) TestIncrementCurrentAssetSupply() {
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
				coin: c("bnb", 5),
			},
			true,
		},
		{
			"equal limit",
			args{
				coin: c("bnb", 10),
			},
			true,
		},
		{
			"exceeds limit",
			args{
				coin: c("bnb", 11),
			},
			false,
		},
		{
			"unsupported asset",
			args{
				coin: c("xyz", 5),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
			err := suite.keeper.IncrementCurrentAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)

			if tc.expectPass {
				suite.True(found)
				suite.NoError(err)
				suite.Equal(preSupply.CurrentSupply.Add(tc.args.coin), postSupply.CurrentSupply)
			} else {
				suite.Error(err)
				suite.Equal(preSupply.CurrentSupply, postSupply.CurrentSupply)
			}
		})
	}
}

func (suite *AssetTestSuite) TestIncrementTimeLimitedCurrentAssetSupply() {
	type args struct {
		coin           sdk.Coin
		expectedSupply types.AssetSupply
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
			"normal",
			args{
				coin: c("inc", 5),
				expectedSupply: types.AssetSupply{
					IncomingSupply:           c("inc", 10),
					OutgoingSupply:           c("inc", 5),
					CurrentSupply:            c("inc", 10),
					TimeLimitedCurrentSupply: c("inc", 5),
					TimeElapsed:              time.Duration(0)},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"over limit",
			args{
				coin:           c("inc", 16),
				expectedSupply: types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "asset supply over limit for current time period",
			},
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			err := suite.keeper.IncrementCurrentAssetSupply(suite.ctx, tc.args.coin)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				supply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
				suite.Require().Equal(tc.args.expectedSupply, supply)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *AssetTestSuite) TestDecrementCurrentAssetSupply() {
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
				coin: c("bnb", 30),
			},
			true,
		},
		{
			"equal current",
			args{
				coin: c("bnb", 40),
			},
			true,
		},
		{
			"exceeds current",
			args{
				coin: c("bnb", 41),
			},
			false,
		},
		{
			"unsupported asset",
			args{
				coin: c("xyz", 30),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
			err := suite.keeper.DecrementCurrentAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)

			if tc.expectPass {
				suite.True(found)
				suite.NoError(err)
				suite.True(preSupply.CurrentSupply.Sub(tc.args.coin).IsEqual(postSupply.CurrentSupply))
			} else {
				suite.Error(err)
				suite.Equal(preSupply.CurrentSupply, postSupply.CurrentSupply)
			}
		})
	}
}

func (suite *AssetTestSuite) TestIncrementIncomingAssetSupply() {
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
				coin: c("bnb", 2),
			},
			true,
		},
		{
			"incoming + current = limit",
			args{
				coin: c("bnb", 5),
			},
			true,
		},
		{
			"incoming + current > limit",
			args{
				coin: c("bnb", 6),
			},
			false,
		},
		{
			"unsupported asset",
			args{
				coin: c("xyz", 2),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
			err := suite.keeper.IncrementIncomingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)

			if tc.expectPass {
				suite.True(found)
				suite.NoError(err)
				suite.Equal(preSupply.IncomingSupply.Add(tc.args.coin), postSupply.IncomingSupply)
			} else {
				suite.Error(err)
				suite.Equal(preSupply.IncomingSupply, postSupply.IncomingSupply)
			}
		})
	}
}

func (suite *AssetTestSuite) TestIncrementTimeLimitedIncomingAssetSupply() {
	type args struct {
		coin           sdk.Coin
		expectedSupply types.AssetSupply
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
			"normal",
			args{
				coin: c("inc", 5),
				expectedSupply: types.AssetSupply{
					IncomingSupply:           c("inc", 15),
					OutgoingSupply:           c("inc", 5),
					CurrentSupply:            c("inc", 5),
					TimeLimitedCurrentSupply: c("inc", 0),
					TimeElapsed:              time.Duration(0)},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"over limit",
			args{
				coin:           c("inc", 6),
				expectedSupply: types.AssetSupply{},
			},
			errArgs{
				expectPass: false,
				contains:   "asset supply over limit for current time period",
			},
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			err := suite.keeper.IncrementIncomingAssetSupply(suite.ctx, tc.args.coin)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				supply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
				suite.Require().Equal(tc.args.expectedSupply, supply)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *AssetTestSuite) TestDecrementIncomingAssetSupply() {
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
				coin: c("bnb", 4),
			},
			true,
		},
		{
			"equal incoming",
			args{
				coin: c("bnb", 5),
			},
			true,
		},
		{
			"exceeds incoming",
			args{
				coin: c("bnb", 6),
			},
			false,
		},
		{
			"unsupported asset",
			args{
				coin: c("xyz", 4),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
			err := suite.keeper.DecrementIncomingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)

			if tc.expectPass {
				suite.True(found)
				suite.NoError(err)
				suite.True(preSupply.IncomingSupply.Sub(tc.args.coin).IsEqual(postSupply.IncomingSupply))
			} else {
				suite.Error(err)
				suite.Equal(preSupply.IncomingSupply, postSupply.IncomingSupply)
			}
		})
	}
}

func (suite *AssetTestSuite) TestIncrementOutgoingAssetSupply() {
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
				coin: c("bnb", 30),
			},
			true,
		},
		{
			"outgoing + amount = current",
			args{
				coin: c("bnb", 35),
			},
			true,
		},
		{
			"outoing + amount > current",
			args{
				coin: c("bnb", 36),
			},
			false,
		},
		{
			"unsupported asset",
			args{
				coin: c("xyz", 30),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
			err := suite.keeper.IncrementOutgoingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)

			if tc.expectPass {
				suite.True(found)
				suite.NoError(err)
				suite.Equal(preSupply.OutgoingSupply.Add(tc.args.coin), postSupply.OutgoingSupply)
			} else {
				suite.Error(err)
				suite.Equal(preSupply.OutgoingSupply, postSupply.OutgoingSupply)
			}
		})
	}
}

func (suite *AssetTestSuite) TestDecrementOutgoingAssetSupply() {
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
				coin: c("bnb", 4),
			},
			true,
		},
		{
			"equal outgoing",
			args{
				coin: c("bnb", 5),
			},
			true,
		},
		{
			"exceeds outgoing",
			args{
				coin: c("bnb", 6),
			},
			false,
		},
		{
			"unsupported asset",
			args{
				coin: c("xyz", 4),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)
			err := suite.keeper.DecrementOutgoingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, tc.args.coin.Denom)

			if tc.expectPass {
				suite.True(found)
				suite.NoError(err)
				suite.True(preSupply.OutgoingSupply.Sub(tc.args.coin).IsEqual(postSupply.OutgoingSupply))
			} else {
				suite.Error(err)
				suite.Equal(preSupply.OutgoingSupply, postSupply.OutgoingSupply)
			}
		})
	}
}

func (suite *AssetTestSuite) TestUpdateTimeBasedSupplyLimits() {
	type args struct {
		asset          string
		duration       time.Duration
		expectedSupply types.AssetSupply
	}
	type errArgs struct {
		expectPanic bool
		contains    string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"rate-limited increment time",
			args{
				asset:          "inc",
				duration:       time.Second,
				expectedSupply: types.NewAssetSupply(c("inc", 10), c("inc", 5), c("inc", 5), c("inc", 0), time.Second),
			},
			errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
		{
			"rate-limited increment time half",
			args{
				asset:          "inc",
				duration:       time.Minute * 30,
				expectedSupply: types.NewAssetSupply(c("inc", 10), c("inc", 5), c("inc", 5), c("inc", 0), time.Minute*30),
			},
			errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
		{
			"rate-limited period change",
			args{
				asset:          "inc",
				duration:       time.Hour + time.Second,
				expectedSupply: types.NewAssetSupply(c("inc", 10), c("inc", 5), c("inc", 5), c("inc", 0), time.Duration(0)),
			},
			errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
		{
			"rate-limited period change exact",
			args{
				asset:          "inc",
				duration:       time.Hour,
				expectedSupply: types.NewAssetSupply(c("inc", 10), c("inc", 5), c("inc", 5), c("inc", 0), time.Duration(0)),
			},
			errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
		{
			"rate-limited period change big",
			args{
				asset:          "inc",
				duration:       time.Hour * 4,
				expectedSupply: types.NewAssetSupply(c("inc", 10), c("inc", 5), c("inc", 5), c("inc", 0), time.Duration(0)),
			},
			errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
		{
			"non rate-limited increment time",
			args{
				asset:          "bnb",
				duration:       time.Second,
				expectedSupply: types.NewAssetSupply(c("bnb", 5), c("bnb", 5), c("bnb", 40), c("bnb", 0), time.Duration(0)),
			},
			errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
		{
			"new asset increment time",
			args{
				asset:          "lol",
				duration:       time.Second,
				expectedSupply: types.NewAssetSupply(c("lol", 0), c("lol", 0), c("lol", 0), c("lol", 0), time.Second),
			},
			errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			deputy, _ := sdk.AccAddressFromBech32(TestDeputy)
			newParams := types.Params{
				AssetParams: types.AssetParams{
					types.AssetParam{
						Denom:  "bnb",
						CoinID: 714,
						SupplyLimit: types.SupplyLimit{
							Limit:          sdk.NewInt(350000000000000),
							TimeLimited:    false,
							TimeBasedLimit: sdk.ZeroInt(),
							TimePeriod:     time.Hour,
						},
						Active:        true,
						DeputyAddress: deputy,
						FixedFee:      sdk.NewInt(1000),
						MinSwapAmount: sdk.OneInt(),
						MaxSwapAmount: sdk.NewInt(1000000000000),
						MinBlockLock:  types.DefaultMinBlockLock,
						MaxBlockLock:  types.DefaultMaxBlockLock,
					},
					types.AssetParam{
						Denom:  "inc",
						CoinID: 9999,
						SupplyLimit: types.SupplyLimit{
							Limit:          sdk.NewInt(100),
							TimeLimited:    true,
							TimeBasedLimit: sdk.NewInt(10),
							TimePeriod:     time.Hour,
						},
						Active:        false,
						DeputyAddress: deputy,
						FixedFee:      sdk.NewInt(1000),
						MinSwapAmount: sdk.OneInt(),
						MaxSwapAmount: sdk.NewInt(1000000000000),
						MinBlockLock:  types.DefaultMinBlockLock,
						MaxBlockLock:  types.DefaultMaxBlockLock,
					},
					types.AssetParam{
						Denom:  "lol",
						CoinID: 9999,
						SupplyLimit: types.SupplyLimit{
							Limit:          sdk.NewInt(100),
							TimeLimited:    true,
							TimeBasedLimit: sdk.NewInt(10),
							TimePeriod:     time.Hour,
						},
						Active:        false,
						DeputyAddress: deputy,
						FixedFee:      sdk.NewInt(1000),
						MinSwapAmount: sdk.OneInt(),
						MaxSwapAmount: sdk.NewInt(1000000000000),
						MinBlockLock:  types.DefaultMinBlockLock,
						MaxBlockLock:  types.DefaultMaxBlockLock,
					},
				},
			}
			suite.keeper.SetParams(suite.ctx, newParams)
			suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(tc.args.duration))
			suite.NotPanics(
				func() {
					suite.keeper.UpdateTimeBasedSupplyLimits(suite.ctx)
				},
			)
			if !tc.errArgs.expectPanic {
				supply, found := suite.keeper.GetAssetSupply(suite.ctx, tc.args.asset)
				suite.Require().True(found)
				suite.Require().Equal(tc.args.expectedSupply, supply)
			}
		})
	}
}

func TestAssetTestSuite(t *testing.T) {
	suite.Run(t, new(AssetTestSuite))
}
