package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// GET/SET/DELETE RewardPeriod can use default genesis state - in fact, if we had params.Rewards set, the reward periods would be created when tApp.InitializeFromGenesisState() was called, because it calls the begin blocker.

// GET/SET/DELETE ClaimPeriod default genesis state

// GET/SET/DELETE Claims default genesis state

// HandleRewardPeriodExpiry default genesis state, set a RewardPeriod

// IterateRewardPeriods default genesis state, set multiple RewardPeriods, iterate

// CreateNewRewardPeriod should use default genesis state

// CreateAndDeleteRewardPeriods default genesis state but needs to add Rewards to params.. Should set a period to inactive and make sure it gets deleted. Should delete a period from the store and make sure when gets created //TODO anything else?

// GetNextClaimPeriodID/GetNextClaimPeriodID default genesis state

// CreateClaimPeriod - default genesis state but need to set next claim period ID for that denom

// IterateClaimPeriodIDKeysAndValues default genesis state but needs a couple denoms with set next claim period ids

// IterateClaims default genesis state

// AddToClaim default genesis state

// ApplyRewardsToCdps - needs a params.Reward in genesis state, can create cdps using the cdp keeper. Needs to check that claims are created and that their values make sense.

// Suite:
// app, ctx, keeper

//  SetupTest - initialize empty app

type RewardsTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *RewardsTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	keeper := tApp.GetIncentiveKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
}

func (suite *RewardsTestSuite) TestGetSetDeleteRewardPeriod() {
	rp := types.NewRewardPeriod("bnb", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), time.Hour*8766)
	_, found := suite.keeper.GetRewardPeriod(suite.ctx, "bnb")
	suite.False(found)
	suite.NotPanics(func() {
		suite.keeper.SetRewardPeriod(suite.ctx, rp)
	})
	testRP, found := suite.keeper.GetRewardPeriod(suite.ctx, "bnb")
	suite.True(found)
	suite.Equal(rp, testRP)
	suite.NotPanics(func() {
		suite.keeper.DeleteRewardPeriod(suite.ctx, "bnb")
	})
	_, found = suite.keeper.GetRewardPeriod(suite.ctx, "bnb")
	suite.False(found)
}

func (suite *RewardsTestSuite) TestGetSetDeleteClaimPeriod() {
}

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func TestRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(RewardsTestSuite))
}
