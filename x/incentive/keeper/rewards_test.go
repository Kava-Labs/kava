package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *KeeperTestSuite) TestAccumulateRewards() {
	type args struct {
		ctype                 string
		rewardsPerSecond      sdk.Coin
		initialTime           time.Time
		initialTotalPrincipal sdk.Coin
		timeElapsed           int
		expectedRewardFactor  sdk.Dec
	}
	type test struct {
		name string
		args args
	}
	testCases := []test{
		{
			"7 seconds",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           7,
				expectedRewardFactor:  d("0.000000856478000000"),
			},
		},
		{
			"1 day",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           86400,
				expectedRewardFactor:  d("0.0105713856"),
			},
		},
		{
			"0 seconds",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           0,
				expectedRewardFactor:  d("0.0"),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithCDPGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			cdpKeeper.SetTotalPrincipal(suite.ctx, tc.args.ctype, cdptypes.DefaultStableDenom, tc.args.initialTotalPrincipal.Amount)

			// setup incentive state
			params := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousAccrualTime(suite.ctx, tc.args.ctype, tc.args.initialTime)
			suite.keeper.SetRewardFactor(suite.ctx, tc.args.ctype, sdk.ZeroDec())

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err := suite.keeper.AccumulateRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			rewardFactor, found := suite.keeper.GetRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *KeeperTestSuite) TestSyncRewards() {
	type args struct {
		ctype                string
		rewardsPerSecond     sdk.Coin
		initialTime          time.Time
		initialCollateral    sdk.Coin
		initialPrincipal     sdk.Coin
		blockTimes           []int
		expectedRewardFactor sdk.Dec
		expectedRewards      sdk.Coin
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"10 blocks",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("0.001223540000000000"),
				expectedRewards:      c("ukava", 12235400),
			},
		},
		{
			"10 blocks - long block time",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardFactor: d("10.57138560000000000"),
				expectedRewards:      c("ukava", 105713856000),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithCDPGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup incentive state
			params := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousAccrualTime(suite.ctx, tc.args.ctype, tc.args.initialTime)
			suite.keeper.SetRewardFactor(suite.ctx, tc.args.ctype, sdk.ZeroDec())

			// setup account state
			sk := suite.app.GetSupplyKeeper()
			sk.MintCoins(suite.ctx, cdptypes.ModuleName, sdk.NewCoins(tc.args.initialCollateral))
			sk.SendCoinsFromModuleToAccount(suite.ctx, cdptypes.ModuleName, suite.addrs[0], sdk.NewCoins(tc.args.initialCollateral))

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			err := cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)
				rewardPeriod, found := suite.keeper.GetRewardPeriod(blockCtx, tc.args.ctype)
				suite.Require().True(found)
				err := suite.keeper.AccumulateRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(suite.ctx, suite.addrs[0], tc.args.ctype)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeReward(suite.ctx, cdp)
			})

			rewardFactor, found := suite.keeper.GetRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)

			claim, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0])
			fmt.Println(claim)
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedRewardFactor, claim.RewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, claim.Reward)
		})
	}

}

func TestRewardCalculation(t *testing.T) {

	// Test Params
	ctype := "bnb-a"
	initialTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	rewardsPerSecond := c("ukava", 122_354)
	initialCollateral := c("bnb", 10_000_000_000)
	initialPrincipal := c("usdx", 100_000_000)
	oneYear := time.Hour * 24 * 365
	rewardPeriod := types.NewRewardPeriod(
		true,
		ctype,
		initialTime,
		initialTime.Add(4*oneYear),
		rewardsPerSecond,
	)

	// Setup app and module params
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: initialTime})
	tApp.InitializeFromGenesisStates(
		app.NewAuthGenState(addrs[:1], []sdk.Coins{cs(initialCollateral)}),
		NewPricefeedGenStateMulti(),
		NewCDPGenStateHighInterest(),
		NewIncentiveGenState(initialTime, initialTime.Add(oneYear), rewardPeriod),
	)
	fmt.Println("RUN INIT GENESIS")

	// Create a CDP
	cdpKeeper := tApp.GetCDPKeeper()
	err := cdpKeeper.AddCdp(
		ctx,
		addrs[0],
		initialCollateral,
		initialPrincipal,
		ctype,
	)
	require.NoError(t, err)

	// Calculate expected cdp reward using iteration

	// Use 10 blocks, each a very long 630720s, to total 6307200s or 1/5th of a year
	// The cdp stability fee is set to the max value 500%, so this time ensures the debt increases a significant amount (doubles)
	// High stability fees increase the chance of catching calculation bugs.
	blockTimes := newRepeatingSliceInt(630720, 10)
	expectedCDPReward := sdk.ZeroDec() //c(rewardPeriod.RewardsPerSecond.Denom, 0)
	for _, bt := range blockTimes {
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(int(time.Second) * bt)))

		// run cdp and incentive begin blockers to update factors
		tApp.BeginBlocker(ctx, abci.RequestBeginBlock{})

		// calculate expected cdp reward
		cdpBlockReward, err := calculateCDPBlockReward(ctx, cdpKeeper, addrs[0], ctype, sdk.NewInt(int64(bt)), rewardPeriod)
		require.NoError(t, err)
		expectedCDPReward = expectedCDPReward.Add(cdpBlockReward)
	}

	// calculate cdp reward using factor
	cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, addrs[0], ctype)
	require.True(t, found)
	incentiveKeeper := tApp.GetIncentiveKeeper()
	require.NotPanics(t, func() {
		incentiveKeeper.SynchronizeReward(ctx, cdp)
	})
	claim, found := incentiveKeeper.GetClaim(ctx, addrs[0])
	require.True(t, found)

	// Compare two methods of calculation
	relativeError := expectedCDPReward.Sub(claim.Reward.Amount.ToDec()).Quo(expectedCDPReward).Abs()
	maxError := d("0.0001")
	require.Truef(t, relativeError.LT(maxError),
		"percent diff %s > %s , expected: %s, actual %s,", relativeError, maxError, expectedCDPReward, claim.Reward.Amount,
	)
}

// calculateCDPBlockReward computes the reward that should be distributed to a cdp for the current block.
func calculateCDPBlockReward(ctx sdk.Context, cdpKeeper cdpkeeper.Keeper, owner sdk.AccAddress, ctype string, timeElapsed sdk.Int, rewardPeriod types.RewardPeriod) (sdk.Dec, error) {
	// Calculate total rewards to distribute this block
	newRewards := timeElapsed.Mul(rewardPeriod.RewardsPerSecond.Amount)

	// Calculate cdp's share of total debt
	totalPrincipal := cdpKeeper.GetTotalPrincipal(ctx, ctype, types.PrincipalDenom).ToDec()
	// cdpDebt
	cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, owner, ctype)
	if !found {
		return sdk.Dec{}, fmt.Errorf("couldn't find cdp for owner '%s' and collateral type '%s'", owner, ctype)
	}
	accumulatedInterest := cdpKeeper.CalculateNewInterest(ctx, cdp)
	cdpDebt := cdp.Principal.Add(cdp.AccumulatedFees).Add(accumulatedInterest).Amount

	// Calculate cdp's reward
	return newRewards.Mul(cdpDebt).ToDec().Quo(totalPrincipal), nil
}

func (suite *KeeperTestSuite) SetupWithCDPGenState() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	keeper := tApp.GetIncentiveKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
}

// newRepeatingSliceInt creates a slice of the specified length containing a single repeating element.
func newRepeatingSliceInt(element int, length int) []int {
	slice := make([]int, length)
	for i := 0; i < length; i++ {
		slice[i] = element
	}
	return slice
}
