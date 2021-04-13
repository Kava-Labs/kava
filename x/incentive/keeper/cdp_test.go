package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *KeeperTestSuite) TestRiskyCDPsAccumulateRewards() {
	suite.SetupWithGenState()
	initialTime := suite.ctx.BlockTime()

	// Setup incentive state
	collateralType := "bnb-a"
	rewardsPerSecond := c(types.USDXMintingRewardDenom, 1_000_000)
	params := types.NewParams(
		types.RewardPeriods{types.NewRewardPeriod(true, collateralType, initialTime, initialTime.Add(4*oneYear), rewardsPerSecond)},
		nil, // hard rewards not needed
		nil,
		nil, // delegator rewards not needed
		types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
		initialTime.Add(5*oneYear),
	)
	suite.keeper.SetParams(suite.ctx, params)
	suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, collateralType, initialTime)
	suite.keeper.SetUSDXMintingRewardFactor(suite.ctx, collateralType, sdk.ZeroDec())

	// Setup cdp state containing one CDP
	cdpKeeper := suite.app.GetCDPKeeper()
	initialCollateral := c("bnb", 1_000_000_000)
	initialPrincipal := c("usdx", 100_000_000)
	cdpKeeper.SetPreviousAccrualTime(suite.ctx, collateralType, suite.ctx.BlockTime())
	cdpKeeper.SetInterestFactor(suite.ctx, collateralType, sdk.OneDec())
	// add coins to user's address // TODO move this to auth genesis setup
	sk := suite.app.GetSupplyKeeper()
	sk.MintCoins(suite.ctx, cdptypes.ModuleName, cs(initialCollateral))
	sk.SendCoinsFromModuleToAccount(suite.ctx, cdptypes.ModuleName, suite.addrs[0], cs(initialCollateral))

	err := cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], initialCollateral, initialPrincipal, collateralType)
	suite.Require().NoError(err)

	// Skip ahead two blocks to accumulate both interest and usdx reward for the cdp
	// Two blocks are required because the cdp begin blocker runs before incentive begin blocker.
	// So in the first begin block, the cdp is synced, which syncs rewards, but no rewards have accumulated yet. Rewards accumulate immediately after when the incentive begin block runs.
	// Rewards are added to the cdp in the next cdp begin blocker (where it is synced).
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(10 * time.Minute))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(10 * time.Minute))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})

	// check cdp rewards
	cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(suite.ctx, suite.addrs[0], collateralType)
	suite.Require().True(found)
	suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)
	claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	// rewards are roughly rewardsPerSecond * secondsElapsed (10mins) * num blocks (2)
	suite.Require().Equal(c(types.USDXMintingRewardDenom, 1_200_001_671), claim.Reward)
}
