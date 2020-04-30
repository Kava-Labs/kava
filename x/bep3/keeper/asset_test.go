package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
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

	// Set asset supply with standard value for testing
	supply := types.AssetSupply{
		Denom:          "bnb",
		IncomingSupply: c("bnb", 5),
		OutgoingSupply: c("bnb", 5),
		CurrentSupply:  c("bnb", 40),
		Limit:          c("bnb", 50),
	}
	keeper.SetAssetSupply(ctx, supply, []byte(supply.Denom))

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
			supplyKeyPrefix := []byte(tc.args.coin.Denom)

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)
			err := suite.keeper.IncrementCurrentAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)

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
			supplyKeyPrefix := []byte(tc.args.coin.Denom)

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)
			err := suite.keeper.DecrementCurrentAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)

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
			supplyKeyPrefix := []byte(tc.args.coin.Denom)

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)
			err := suite.keeper.IncrementIncomingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)

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
			supplyKeyPrefix := []byte(tc.args.coin.Denom)

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)
			err := suite.keeper.DecrementIncomingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)

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
			supplyKeyPrefix := []byte(tc.args.coin.Denom)

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)
			err := suite.keeper.IncrementOutgoingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)

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
			supplyKeyPrefix := []byte(tc.args.coin.Denom)

			preSupply, found := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)
			err := suite.keeper.DecrementOutgoingAssetSupply(suite.ctx, tc.args.coin)
			postSupply, _ := suite.keeper.GetAssetSupply(suite.ctx, supplyKeyPrefix)

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

func TestAssetTestSuite(t *testing.T) {
	suite.Run(t, new(AssetTestSuite))
}
