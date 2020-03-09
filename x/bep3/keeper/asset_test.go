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

func (suite *AssetTestSuite) TestValidateActiveAsset() {
	// Make an atomic swap in order to set the asset supply
	timestamp := tmtime.Now().Unix()
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	_ = suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, int64(360), suite.addrs[0],
		suite.addrs[1], binanceAddrs[0].String(), binanceAddrs[1].String(), cs(c("bnb", 1)), "1bnb")

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
		err := suite.keeper.ValidateActiveAsset(suite.ctx, tc.args.coin)

		if tc.expectPass {
			suite.NoError(err)
		} else {
			suite.Error(err)
			suite.Equal(tc.expectedError, err.Result().Code)
		}
	}
}

func (suite *AssetTestSuite) TestValidateProposedIncrease() {
	// Make an atomic swap in order to set the asset supply
	timestamp := tmtime.Now().Unix()
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	_ = suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, int64(360), suite.addrs[0],
		suite.addrs[1], binanceAddrs[0].String(), binanceAddrs[1].String(), cs(c("bnb", 1)), "1bnb")

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
			"amount too small",
			args{
				coin: c("bnb", 0),
			},
			types.CodeAmountTooSmall,
			false,
		},
		{
			"asset supply not set",
			args{
				coin: c("inc", 500),
			},
			types.CodeAssetSupplyNotSet,
			false,
		},
		{
			"above asset supply limit",
			args{
				coin: c("bnb", BNB_SUPPLY_LIMIT.Int64()+1),
			},
			types.CodeAboveAssetSupplyLimit,
			false,
		},
	}

	for _, tc := range testCases {
		err := suite.keeper.ValidateProposedIncrease(suite.ctx, tc.args.coin)

		if tc.expectPass {
			suite.NoError(err)
		} else {
			suite.Error(err)
			suite.Equal(tc.expectedError, err.Result().Code)
		}
	}
}

func (suite *AssetTestSuite) TestIncrementAssetSupply() {
	// Make an atomic swap in order to set the asset supply
	timestamp := tmtime.Now().Unix()
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	_ = suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, int64(360), suite.addrs[0],
		suite.addrs[1], binanceAddrs[0].String(), binanceAddrs[1].String(), cs(c("bnb", 1)), "1bnb")

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
	}

	for _, tc := range testCases {
		assetSupplyPre, _ := suite.keeper.GetAssetSupply(suite.ctx, []byte(tc.args.coin.Denom))
		err := suite.keeper.IncrementAssetSupply(suite.ctx, tc.args.coin)
		assetSupplyPost, _ := suite.keeper.GetAssetSupply(suite.ctx, []byte(tc.args.coin.Denom))

		if tc.expectPass {
			suite.NoError(err)
			// Check asset supply changed
			suite.Equal(assetSupplyPre.Add(tc.args.coin), assetSupplyPost)
		} else {
			suite.Error(err)
			// Check asset supply hasn't changed
			suite.Equal(assetSupplyPre, assetSupplyPost)
		}
	}
}

func TestAssetTestSuite(t *testing.T) {
	suite.Run(t, new(AssetTestSuite))
}
