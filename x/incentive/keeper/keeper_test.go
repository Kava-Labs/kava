package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	keeper := tApp.GetIncentiveKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) TestGetSetDeleteRewardPeriod() {
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

func (suite *KeeperTestSuite) TestGetSetDeleteClaimPeriod() {
	cp := types.NewClaimPeriod("bnb", 1, suite.ctx.BlockTime().Add(time.Hour*168), time.Hour*8766)
	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
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
}

func (suite *KeeperTestSuite) TestGetSetClaimPeriodID() {
	suite.Panics(func() {
		suite.keeper.GetNextClaimPeriodID(suite.ctx, "bnb")
	})
	suite.NotPanics(func() {
		suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
	})
	testID := suite.keeper.GetNextClaimPeriodID(suite.ctx, "bnb")
	suite.Equal(uint64(1), testID)
}

func (suite *KeeperTestSuite) TestGetSetDeleteClaim() {
	c := types.NewClaim(suite.addrs[0], c("ukava", 1000000), "bnb", 1)
	_, found := suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.False(found)
	suite.NotPanics(func() {
		suite.keeper.SetClaim(suite.ctx, c)
	})
	testC, found := suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.True(found)
	suite.Equal(c, testC)
	suite.NotPanics(func() {
		suite.keeper.DeleteClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	})
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestIterateMethods() {
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

func (suite *KeeperTestSuite) addObjectsToStore() {
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

	c1 := types.NewClaim(suite.addrs[0], c("ukava", 1000000), "bnb", 1)
	c2 := types.NewClaim(suite.addrs[0], c("ukava", 1000000), "xrp", 1)
	suite.keeper.SetClaim(suite.ctx, c1)
	suite.keeper.SetClaim(suite.ctx, c2)

	params := types.NewParams(
		true, types.Rewards{types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)},
	)
	suite.keeper.SetParams(suite.ctx, params)

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
