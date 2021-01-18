package incentive_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
)

func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     app.TestApp
	handler sdk.Handler
	keeper  incentive.Keeper
	addrs   []sdk.AccAddress
}

func (suite *HandlerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetIncentiveKeeper()

	// Set up genesis state and initialize
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := []sdk.Coins{}
	for j := 0; j < 3; j++ {
		coins = append(coins, cs(c("bnb", 10000000000), c("ukava", 10000000000)))
	}
	authGS := app.NewAuthGenState(addrs, coins)
	incentiveGS := incentive.NewGenesisState(
		incentive.NewParams(
			incentive.RewardPeriods{incentive.NewRewardPeriod(true, "bnb-a", time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC), time.Date(2024, 12, 15, 14, 0, 0, 0, time.UTC), c("ukava", 122354))},
			incentive.Multipliers{incentive.NewMultiplier(incentive.MultiplierName("small"), 1, d("0.25")), incentive.NewMultiplier(incentive.MultiplierName("large"), 12, d("1.0"))},
			time.Date(2025, 12, 15, 14, 0, 0, 0, time.UTC),
		),
		incentive.DefaultGenesisAccumulationTimes,
		incentive.DefaultClaims,
	)
	tApp.InitializeFromGenesisStates(authGS, app.GenesisState{incentive.ModuleName: incentive.ModuleCdc.MustMarshalJSON(incentiveGS)})

	suite.addrs = addrs
	suite.handler = incentive.NewHandler(keeper)
	suite.keeper = keeper
	suite.app = tApp
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) addClaim() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, kavadist.ModuleName, cs(c("ukava", 1000000000000)))
	suite.Require().NoError(err)
	c1 := incentive.NewUSDXMintingClaim(suite.addrs[0], c("ukava", 1000000), types.RewardIndexes{types.NewRewardIndex("bnb-s", sdk.ZeroDec())})
	suite.NotPanics(func() {
		suite.keeper.SetClaim(suite.ctx, c1)
	})
}

func (suite *HandlerTestSuite) TestMsgClaimReward() {
	suite.addClaim()
	msg := incentive.NewMsgClaimUSDXMintingReward(suite.addrs[0], "small")
	res, err := suite.handler(suite.ctx, msg)
	suite.NoError(err)
	suite.Require().NotNil(res)
}
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int   { return sdk.NewInt(in) }
func d(str string) sdk.Dec { return sdk.MustNewDecFromStr(str) }
