package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *KeeperTestSuite) TestAccumulateUSDXMintingRewards() {
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
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			cdpKeeper.SetTotalPrincipal(suite.ctx, tc.args.ctype, cdptypes.DefaultStableDenom, tc.args.initialTotalPrincipal.Amount)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, tc.args.ctype, tc.args.initialTime)
			suite.keeper.SetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype, sdk.ZeroDec())

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err := suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			rewardFactor, found := suite.keeper.GetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *KeeperTestSuite) TestSynchronizeUSDXMintingReward() {
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
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, tc.args.ctype, tc.args.initialTime)
			suite.keeper.SetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype, sdk.ZeroDec())

			// setup account state
			sk := suite.app.GetSupplyKeeper()
			sk.MintCoins(suite.ctx, cdptypes.ModuleName, sdk.NewCoins(tc.args.initialCollateral))
			sk.SendCoinsFromModuleToAccount(suite.ctx, cdptypes.ModuleName, suite.addrs[0], sdk.NewCoins(tc.args.initialCollateral))

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			err := cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)
				rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(blockCtx, tc.args.ctype)
				suite.Require().True(found)
				err := suite.keeper.AccumulateUSDXMintingRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(suite.ctx, suite.addrs[0], tc.args.ctype)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)
			})

			rewardFactor, found := suite.keeper.GetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)

			claim, found = suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedRewardFactor, claim.RewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, claim.Reward)
		})
	}
}

func (suite *KeeperTestSuite) TestAccumulateHardBorrowRewards() {
	type args struct {
		borrow               sdk.Coin
		rewardsPerSecond     sdk.Coin
		initialTime          time.Time
		timeElapsed          int
		expectedRewardFactor sdk.Dec
	}
	type test struct {
		name string
		args args
	}
	testCases := []test{
		{
			"7 seconds",
			args{
				borrow:               c("bnb", 1000000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:          7,
				expectedRewardFactor: d("0.000000856478000000"),
			},
		},
		{
			"1 day",
			args{
				borrow:               c("bnb", 1000000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:          86400,
				expectedRewardFactor: d("0.0105713856"),
			},
		},
		{
			"0 seconds",
			args{
				borrow:               c("bnb", 1000000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:          0,
				expectedRewardFactor: d("0.0"),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardBorrowRewardAccrualTime(suite.ctx, tc.args.borrow.Denom, tc.args.initialTime)
			suite.keeper.SetHardBorrowRewardFactor(suite.ctx, tc.args.borrow.Denom, sdk.ZeroDec())

			// User deposits and borrows to increase total borrowed amount
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
			suite.Require().NoError(err)

			// Set up chain context at future time
			runAtTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			rewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriod(runCtx, tc.args.borrow.Denom)
			suite.Require().True(found)
			err = suite.keeper.AccumulateHardBorrowRewards(runCtx, rewardPeriod)
			suite.Require().NoError(err)

			rewardFactor, found := suite.keeper.GetHardBorrowRewardFactor(runCtx, tc.args.borrow.Denom)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *KeeperTestSuite) TestSynchronizeHardBorrowReward() {
	type args struct {
		borrow               sdk.Coin
		rewardsPerSecond     sdk.Coin
		initialTime          time.Time
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
				borrow:               c("bnb", 10000000000), // TODO: 2 decimal diff from TestAccumulateHardBorrowRewards's borrow
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("0.001223540000000000"),
				expectedRewards:      c("hard", 12235400),
			},
		},
		{
			"10 blocks - long block time",
			args{
				borrow:               c("bnb", 10000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:           []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardFactor: d("10.57138560000000000"),
				expectedRewards:      c("hard", 105713856000),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardBorrowRewardAccrualTime(suite.ctx, tc.args.borrow.Denom, tc.args.initialTime)
			suite.keeper.SetHardBorrowRewardFactor(suite.ctx, tc.args.borrow.Denom, sdk.ZeroDec())

			// User deposits and borrows to increase total borrowed amount
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.BorrowRewardIndexes[0].RewardFactor)

			// Run accumulator at several intervals
			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)

				rewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriod(blockCtx, tc.args.borrow.Denom)
				suite.Require().True(found)

				err := suite.keeper.AccumulateHardBorrowRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// After we've accumulated, run synchronize
			borrow, found := hardKeeper.GetBorrow(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)
			})

			// Check that reward factor and claim have been updated as expected
			rewardFactor, found := suite.keeper.GetHardBorrowRewardFactor(suite.ctx, tc.args.borrow.Denom)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)

			claim, found = suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedRewardFactor, claim.BorrowRewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, claim.Reward)
		})
	}
}

func (suite *KeeperTestSuite) TestAccumulateHardSupplyRewards() {
	type args struct {
		deposit              sdk.Coin
		rewardsPerSecond     sdk.Coin
		initialTime          time.Time
		timeElapsed          int
		expectedRewardFactor sdk.Dec
	}
	type test struct {
		name string
		args args
	}
	testCases := []test{
		{
			"7 seconds",
			args{
				deposit:              c("bnb", 1000000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:          7,
				expectedRewardFactor: d("0.000000856478000000"),
			},
		},
		{
			"1 day",
			args{
				deposit:              c("bnb", 1000000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:          86400,
				expectedRewardFactor: d("0.0105713856"),
			},
		},
		{
			"0 seconds",
			args{
				deposit:              c("bnb", 1000000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:          0,
				expectedRewardFactor: d("0.0"),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, tc.args.deposit.Denom, tc.args.initialTime)
			suite.keeper.SetHardSupplyRewardFactor(suite.ctx, tc.args.deposit.Denom, sdk.ZeroDec())

			// User deposits to increase total supplied amount
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(tc.args.deposit))
			suite.Require().NoError(err)

			// Set up chain context at future time
			runAtTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			rewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriod(runCtx, tc.args.deposit.Denom)
			suite.Require().True(found)
			err = suite.keeper.AccumulateHardSupplyRewards(runCtx, rewardPeriod)
			suite.Require().NoError(err)

			rewardFactor, found := suite.keeper.GetHardSupplyRewardFactor(runCtx, tc.args.deposit.Denom)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *KeeperTestSuite) TestSynchronizeHardSupplyReward() {
	type args struct {
		deposit              sdk.Coin
		rewardsPerSecond     sdk.Coin
		initialTime          time.Time
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
				deposit:              c("bnb", 10000000000), // TODO: 2 decimal diff
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("0.001223540000000000"),
				expectedRewards:      c("hard", 12235400),
			},
		},
		{
			"10 blocks - long block time",
			args{
				deposit:              c("bnb", 10000000000),
				rewardsPerSecond:     c("hard", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:           []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardFactor: d("10.57138560000000000"),
				expectedRewards:      c("hard", 105713856000),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, tc.args.deposit.Denom, tc.args.initialTime)
			suite.keeper.SetHardSupplyRewardFactor(suite.ctx, tc.args.deposit.Denom, sdk.ZeroDec())

			// User deposits and borrows to increase total borrowed amount
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(tc.args.deposit))
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.SupplyRewardIndexes[0].RewardFactor)

			// Run accumulator at several intervals
			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)

				rewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriod(blockCtx, tc.args.deposit.Denom)
				suite.Require().True(found)

				err := suite.keeper.AccumulateHardSupplyRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// After we've accumulated, run synchronize
			deposit, found := hardKeeper.GetDeposit(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)
			})

			// Check that reward factor and claim have been updated as expected
			rewardFactor, found := suite.keeper.GetHardSupplyRewardFactor(suite.ctx, tc.args.deposit.Denom)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)

			claim, found = suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedRewardFactor, claim.SupplyRewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, claim.Reward)
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateHardSupplyIndexDenoms() {
	type args struct {
		firstDeposit              sdk.Coins
		secondDeposit             sdk.Coins
		rewardsPerSecond          sdk.Coin
		initialTime               time.Time
		expectedSupplyIndexDenoms []string
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"update adds one supply reward index",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				secondDeposit:             cs(c("ukava", 10000000000)),
				rewardsPerSecond:          c("hard", 122354),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb", "ukava"},
			},
		},
		{
			"update adds multiple supply reward indexes",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				secondDeposit:             cs(c("ukava", 10000000000), c("btcb", 10000000000), c("xrp", 10000000000)),
				rewardsPerSecond:          c("hard", 122354),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb", "ukava", "btcb", "xrp"},
			},
		},
		{
			"update doesn't add supply reward index for same denom",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				secondDeposit:             cs(c("bnb", 5000000000)),
				rewardsPerSecond:          c("hard", 122354),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb"},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// Set up generic reward periods
			var rewardPeriods types.RewardPeriods
			for _, denom := range tc.args.expectedSupplyIndexDenoms {
				rewardPeriod := types.NewRewardPeriod(true, denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)
				rewardPeriods = append(rewardPeriods, rewardPeriod)
			}

			// Setup incentive state
			params := types.NewParams(
				true,
				rewardPeriods, rewardPeriods, rewardPeriods, rewardPeriods,
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)

			// Set each denom's previous accrual time and supply reward factor
			for _, denom := range tc.args.expectedSupplyIndexDenoms {
				suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, denom, tc.args.initialTime)
				suite.keeper.SetHardSupplyRewardFactor(suite.ctx, denom, sdk.ZeroDec())
			}

			// User deposits (first time)
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, tc.args.firstDeposit)
			suite.Require().NoError(err)

			// Confirm that a claim was created and populated with the correct supply indexes
			claimAfterFirstDeposit, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			for _, coin := range tc.args.firstDeposit {
				_, hasIndex := claimAfterFirstDeposit.HasSupplyRewardIndex(coin.Denom)
				suite.Require().True(hasIndex)
			}
			suite.Require().True(len(claimAfterFirstDeposit.SupplyRewardIndexes) == len(tc.args.firstDeposit))

			// User deposits (second time)
			err = hardKeeper.Deposit(suite.ctx, userAddr, tc.args.secondDeposit)
			suite.Require().NoError(err)

			// Confirm that the claim contains all expected supply indexes
			claimAfterSecondDeposit, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			for _, denom := range tc.args.expectedSupplyIndexDenoms {
				_, hasIndex := claimAfterSecondDeposit.HasSupplyRewardIndex(denom)
				suite.Require().True(hasIndex)
			}
			suite.Require().True(len(claimAfterSecondDeposit.SupplyRewardIndexes) == len(tc.args.expectedSupplyIndexDenoms))
		})
	}
}

func (suite *KeeperTestSuite) SetupWithGenState() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	_, addrs := app.GeneratePrivKeyAddressPairs(5)

	authGS := app.NewAuthGenState(
		[]sdk.AccAddress{addrs[3]},
		[]sdk.Coins{
			sdk.NewCoins(
				sdk.NewCoin("bnb", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("ukava", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("btcb", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("xrp", sdk.NewInt(1000000000000000)),
			),
		},
	)

	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
		NewHardGenStateMulti(),
	)

	keeper := tApp.GetIncentiveKeeper()
	hardKeeper := tApp.GetHardKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.hardKeeper = hardKeeper
	suite.addrs = addrs
}
