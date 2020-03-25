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

func (suite *RewardsTestSuite) TestGetSetDeleteMethods() {
	// reward periods
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

	// claim periods
	cp := types.NewClaimPeriod("bnb", 1, suite.ctx.BlockTime().Add(time.Hour*168), time.Hour*8766)
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.False(found)
	suite.NotPanics(func() {
		suite.keeper.SetClaimPeriod(suite.ctx, cp)
	})
	testCP, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.True(found)
	suite.Equal(cp, testCP)
	suite.NotPanics(func() {
		suite.keeper.DeleteClaimPeriod(suite.ctx, 1, "bnb")
	})
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.False(found)

	// next claim period id
	suite.Panics(func() {
		suite.keeper.GetNextClaimPeriodID(suite.ctx, "bnb")
	})
	suite.NotPanics(func() {
		suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
	})
	testID := suite.keeper.GetNextClaimPeriodID(suite.ctx, "bnb")
	suite.Equal(uint64(1), testID)

	// claims
	addr, _ := sdk.AccAddressFromBech32("kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw")
	c := types.NewClaim(addr, c("ukava", 1000000), "bnb", 1)
	_, found = suite.keeper.GetClaim(suite.ctx, addr, "bnb", 1)
	suite.False(found)
	suite.NotPanics(func() {
		suite.keeper.SetClaim(suite.ctx, c)
	})
	testC, found := suite.keeper.GetClaim(suite.ctx, addr, "bnb", 1)
	suite.True(found)
	suite.Equal(c, testC)
	suite.NotPanics(func() {
		suite.keeper.DeleteClaim(suite.ctx, addr, "bnb", 1)
	})
	_, found = suite.keeper.GetClaim(suite.ctx, addr, "bnb", 1)
	suite.False(found)
}

func (suite *RewardsTestSuite) TestIterateMethods() {
	suite.addObjectsToStore() // adds 2 objects of each type to the store

	var rewardPeriods types.RewardPeriods
	suite.keeper.IterateRewardPeriods(suite.ctx, func(rp types.RewardPeriod) (stop bool) {
		rewardPeriods = append(rewardPeriods, rp)
		return false
	})
	suite.Equal(2, len(rewardPeriods))

	var claimPeriods types.ClaimPeriods
	suite.keeper.IterateClaimPeriods(suite.ctx, func(cp types.ClaimPeriod) (stop bool) {
		claimPeriods = append(claimPeriods, cp)
		return false
	})
	suite.Equal(2, len(claimPeriods))

	var claims types.Claims
	suite.keeper.IterateClaims(suite.ctx, func(c types.Claim) (stop bool) {
		claims = append(claims, c)
		return false
	})
	suite.Equal(2, len(claims))

	var genIDs types.GenesisClaimPeriodIDs
	suite.keeper.IterateClaimPeriodIDKeysAndValues(suite.ctx, func(denom string, id uint64) (stop bool) {
		genID := types.GenesisClaimPeriodID{Denom: denom, ID: id}
		genIDs = append(genIDs, genID)
		return false
	})
	suite.Equal(2, len(genIDs))
}

func (suite *RewardsTestSuite) TestHelperMethods() {
	rp := types.NewRewardPeriod("bnb", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), time.Hour*8766)
	suite.keeper.SetRewardPeriod(suite.ctx, rp)
	suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
	suite.NotPanics(func() {
		suite.keeper.HandleRewardPeriodExpiry(suite.ctx, rp)
	})
	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.True(found)

	addr, _ := sdk.AccAddressFromBech32("kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw")
	c1 := types.NewClaim(addr, c("ukava", 1000000), "bnb", 1)
	suite.keeper.SetClaim(suite.ctx, c1)
	suite.NotPanics(func() {
		suite.keeper.AddToClaim(suite.ctx, addr, "bnb", 1, c("ukava", 1000000))
	})
	testC, _ := suite.keeper.GetClaim(suite.ctx, addr, "bnb", 1)
	suite.Equal(c("ukava", 2000000), testC.Reward)

	suite.NotPanics(func() {
		suite.keeper.AddToClaim(suite.ctx, addr, "xpr", 1, c("ukava", 1000000))
	})

	suite.SetupTest()
	reward := types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)
	suite.NotPanics(func() {
		suite.keeper.CreateNewRewardPeriod(suite.ctx, reward)
	})
	_, found = suite.keeper.GetRewardPeriod(suite.ctx, "bnb")
	suite.True(found)

}

func (suite *RewardsTestSuite) TestCreateAndDeleteRewardsPeriods() {
	reward1 := types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)
	reward2 := types.NewReward(false, "xrp", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)
	params := types.NewParams(true, types.Rewards{reward1, reward2})
	suite.keeper.SetParams(suite.ctx, params)

	suite.NotPanics(func() {
		suite.keeper.CreateAndDeleteRewardPeriods(suite.ctx)
	})

	_, found := suite.keeper.GetRewardPeriod(suite.ctx, "bnb")
	suite.True(found)
	_, found = suite.keeper.GetRewardPeriod(suite.ctx, "xrp")
	suite.False(found)

}

func (suite *RewardsTestSuite) addObjectsToStore() {
	rp1 := types.NewRewardPeriod("bnb", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), time.Hour*8766)
	rp2 := types.NewRewardPeriod("xrp", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), time.Hour*8766)
	suite.keeper.SetRewardPeriod(suite.ctx, rp1)
	suite.keeper.SetRewardPeriod(suite.ctx, rp2)

	cp1 := types.NewClaimPeriod("bnb", 1, suite.ctx.BlockTime().Add(time.Hour*168), time.Hour*8766)
	cp2 := types.NewClaimPeriod("xrp", 1, suite.ctx.BlockTime().Add(time.Hour*168), time.Hour*8766)
	suite.keeper.SetClaimPeriod(suite.ctx, cp1)
	suite.keeper.SetClaimPeriod(suite.ctx, cp2)

	suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
	suite.keeper.SetNextClaimPeriodID(suite.ctx, "xrp", 1)

	addr, _ := sdk.AccAddressFromBech32("kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw")
	c1 := types.NewClaim(addr, c("ukava", 1000000), "bnb", 1)
	c2 := types.NewClaim(addr, c("ukava", 1000000), "xrp", 1)
	suite.keeper.SetClaim(suite.ctx, c1)
	suite.keeper.SetClaim(suite.ctx, c2)

}

func (suite *RewardsTestSuite) setupCdpChain() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	// need pricefeed and cdp gen state with one collateral
	// need kavadist module account with large balance and a helper to drain it's balance
	// need incentive params for one collateral

	// need to create 3 cdps

}

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func TestRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(RewardsTestSuite))
}
