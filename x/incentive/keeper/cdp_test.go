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
	// In the first begin block the cdp is synced, which triggers its claim to sync. But no global rewards have accumulated yet so the sync does nothing.
	// Global rewards accumulate immediately after during the incentive begin blocker.
	// Rewards are added to the cdp's claim in the next block when the cdp is synced.
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(10 * time.Minute))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(10 * time.Minute))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})

	// check cdp rewards
	cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(suite.ctx, suite.addrs[0], collateralType)
	suite.Require().True(found)
	// This additional sync adds the rewards accumulated at the end of the last begin block.
	// They weren't added during the begin blocker as the incentive BB runs after the CDP BB.
	suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)
	claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	// rewards are roughly rewardsPerSecond * secondsElapsed (10mins) * num blocks (2)
	suite.Require().Equal(c(types.USDXMintingRewardDenom, 1_200_001_671), claim.Reward)
}
