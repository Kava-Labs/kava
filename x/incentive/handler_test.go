package incentive_test

import (
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/hard"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }

// Test suite used for all keeper tests
type HandlerTestSuite struct {
	suite.Suite

	keeper     keeper.Keeper
	hardKeeper hardkeeper.Keeper
	cdpKeeper  cdpkeeper.Keeper
	handler    sdk.Handler

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

// SetupTest is run automatically before each suite test
func (suite *HandlerTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *HandlerTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.hardKeeper = suite.app.GetHardKeeper()
	suite.cdpKeeper = suite.app.GetCDPKeeper()
	suite.handler = incentive.NewHandler(suite.keeper)

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: suite.genesisTime})
}

func (suite *HandlerTestSuite) SetupWithGenState(authBuilder app.AuthGenesisBuilder, incentBuilder testutil.IncentiveGenesisBuilder, hardBuilder testutil.HardGenesisBuilder) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		hardBuilder.BuildMarshalled(),
		incentBuilder.BuildMarshalled(),
	)
}

func (suite *HandlerTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *HandlerTestSuite) getModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *HandlerTestSuite) TestPayoutUSDXMintingClaim() {
	type args struct {
		ctype                    string
		rewardsPerSecond         sdk.Coin
		initialCollateral        sdk.Coin
		initialPrincipal         sdk.Coin
		multipliers              types.Multipliers
		multiplier               string
		timeElapsed              int
		expectedBalance          sdk.Coins
		expectedPeriods          vesting.Periods
		isPeriodicVestingAccount bool
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type test struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []test{
		{
			"valid 1 day",
			args{
				ctype:                    "bnb-a",
				rewardsPerSecond:         c("ukava", 122354),
				initialCollateral:        c("bnb", 1000000000000),
				initialPrincipal:         c("usdx", 10000000000),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400,
				expectedBalance:          cs(c("usdx", 10000000000), c("ukava", 10571385600)),
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("ukava", 10571385600))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid zero rewards",
			args{
				ctype:                    "bnb-a",
				rewardsPerSecond:         c("ukava", 0),
				initialCollateral:        c("bnb", 1000000000000),
				initialPrincipal:         c("usdx", 10000000000),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400,
				expectedBalance:          cs(c("usdx", 10000000000)),
				expectedPeriods:          vesting.Periods{},
				isPeriodicVestingAccount: false,
			},
			errArgs{
				expectPass: false,
				contains:   "claim amount rounds to zero",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			userAddr := suite.addrs[0]
			authBulder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(userAddr, cs(tc.args.initialCollateral)).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("ukava", 1000000000000)))

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond).
				WithMultipliers(tc.args.multipliers)

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// setup cdp state
			err := suite.cdpKeeper.AddCdp(suite.ctx, userAddr, tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err = suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			msg := incentive.NewMsgClaimUSDXMintingReward(userAddr, tc.args.multiplier)
			_, err = suite.handler(suite.ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				acc := suite.getAccount(userAddr)
				suite.Require().Equal(tc.args.expectedBalance, acc.GetCoins())

				if tc.args.isPeriodicVestingAccount {
					vacc, ok := acc.(*vesting.PeriodicVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				}

				claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, userAddr)
				suite.Require().True(found)
				suite.Require().Equal(c("ukava", 0), claim.Reward)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *HandlerTestSuite) TestPayoutUSDXMintingClaimVVesting() {
	type args struct {
		ctype             string
		rewardsPerSecond  sdk.Coin
		initialCollateral sdk.Coin
		initialPrincipal  sdk.Coin
		multipliers       types.Multipliers
		multiplier        string
		timeElapsed       int
		expectedBalance   sdk.Coins
		expectedPeriods   vesting.Periods
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type test struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []test{
		{
			"valid 1 day",
			args{
				ctype:             "bnb-a",
				rewardsPerSecond:  c("ukava", 122354),
				initialCollateral: c("bnb", 1e12),
				initialPrincipal:  c("usdx", 1e10),
				multipliers:       types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:        "large",
				timeElapsed:       86400,
				expectedBalance:   cs(c("ukava", 10571385600)),
				expectedPeriods:   vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("ukava", 10571385600))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid zero rewards",
			args{
				ctype:             "bnb-a",
				rewardsPerSecond:  c("ukava", 0),
				initialCollateral: c("bnb", 1e12),
				initialPrincipal:  c("usdx", 1e10),
				multipliers:       types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:        "large",
				timeElapsed:       86400,
				expectedBalance:   cs(),
				expectedPeriods:   vesting.Periods{},
			},
			errArgs{
				expectPass: false,
				contains:   "claim amount rounds to zero",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			bacc := auth.NewBaseAccount(suite.addrs[2], cs(tc.args.initialCollateral, c("ukava", 400)), nil, 0, 0)
			bva, err := vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), suite.genesisTime.Unix()+16)
			suite.Require().NoError(err)
			periods := vesting.Periods{
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
			}
			vva := validatorvesting.NewValidatorVestingAccountRaw(bva, suite.genesisTime.Unix(), periods, sdk.ConsAddress{}, nil, 90)

			authBulder := app.NewAuthGenesisBuilder().
				WithAccounts(vva).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("ukava", 1e18))).
				WithSimpleAccount(suite.addrs[0], cs()) // the recipient address needs to be a instantiated account // TODO change?

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond).
				WithMultipliers(tc.args.multipliers)

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			err = cdpKeeper.AddCdp(suite.ctx, suite.addrs[2], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[2])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			// accumulate some usdx rewards
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err = suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			msg := incentive.NewMsgClaimUSDXMintingRewardVVesting(suite.addrs[2], suite.addrs[0], tc.args.multiplier)
			_, err = suite.handler(suite.ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				ak := suite.app.GetAccountKeeper()
				acc := ak.GetAccount(suite.ctx, suite.addrs[0])
				suite.Require().Equal(tc.args.expectedBalance, acc.GetCoins())

				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)

				claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[2])
				suite.Require().True(found)
				suite.Require().Equal(c("ukava", 0), claim.Reward)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *HandlerTestSuite) TestPayoutHardLiquidityProviderClaim() {
	type args struct {
		deposit                  sdk.Coins
		borrow                   sdk.Coins
		rewardsPerSecond         sdk.Coins
		multipliers              types.Multipliers
		multiplier               string
		timeElapsed              int64
		expectedRewards          sdk.Coins
		expectedPeriods          vesting.Periods
		isPeriodicVestingAccount bool
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type test struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []test{
		{
			"single reward denom: valid 1 day",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400,
				expectedRewards:          cs(c("hard", 21142771200)), // 10571385600 (deposit reward) + 10571385600 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771200))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"single reward denom: valid 10 days",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              864000,
				expectedRewards:          cs(c("hard", 211427712000)), // 105713856000 (deposit reward) + 105713856000 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712000))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 1 day",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400,
				expectedRewards:          cs(c("hard", 21142771200), c("ukava", 21142771200)), // 10571385600 (deposit reward) + 10571385600 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771200), c("ukava", 21142771200))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 10 days",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              864000,
				expectedRewards:          cs(c("hard", 211427712000), c("ukava", 211427712000)), // 105713856000 (deposit reward) + 105713856000 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712000), c("ukava", 211427712000))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms with different rewards per second: valid 1 day",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354), c("ukava", 222222)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400,
				expectedRewards:          cs(c("hard", 21142771200), c("ukava", 38399961600)),
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771200), c("ukava", 38399961600))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			userAddr := suite.addrs[3]
			authBulder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(userAddr, cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15))).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1000000000000000000), c("ukava", 1000000000000000000)))

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithMultipliers(tc.args.multipliers)
			for _, c := range tc.args.deposit {
				incentBuilder = incentBuilder.WithSimpleSupplyRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}
			for _, c := range tc.args.borrow {
				incentBuilder = incentBuilder.WithSimpleBorrowRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// User deposits and borrows
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)
			err = suite.hardKeeper.Borrow(suite.ctx, userAddr, tc.args.borrow)
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim that has 0 rewards
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			for _, coin := range tc.args.deposit {
				suite.Require().Equal(sdk.ZeroInt(), claim.Reward.AmountOf(coin.Denom))
			}

			// Set up future runtime context
			runAtTime := time.Unix(suite.ctx.BlockTime().Unix()+(tc.args.timeElapsed), 0)
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			// Accumulate supply rewards for each deposit denom
			for _, coin := range tc.args.deposit {
				rewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriods(runCtx, coin.Denom)
				suite.Require().True(found)
				err = suite.keeper.AccumulateHardSupplyRewards(runCtx, rewardPeriod)
				suite.Require().NoError(err)
			}

			// Accumulate borrow rewards for each deposit denom
			for _, coin := range tc.args.borrow {
				rewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriods(runCtx, coin.Denom)
				suite.Require().True(found)
				err = suite.keeper.AccumulateHardBorrowRewards(runCtx, rewardPeriod)
				suite.Require().NoError(err)
			}

			// Sync hard supply rewards
			deposit, found := suite.hardKeeper.GetDeposit(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

			// Sync hard borrow rewards
			borrow, found := suite.hardKeeper.GetBorrow(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

			// Fetch pre-claim balances
			ak := suite.app.GetAccountKeeper()
			preClaimAcc := ak.GetAccount(runCtx, userAddr)

			msg := types.NewMsgClaimHardReward(userAddr, tc.args.multiplier)
			_, err = suite.handler(runCtx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check that user's balance has increased by expected reward amount
				postClaimAcc := ak.GetAccount(suite.ctx, userAddr)
				suite.Require().Equal(preClaimAcc.GetCoins().Add(tc.args.expectedRewards...), postClaimAcc.GetCoins())

				if tc.args.isPeriodicVestingAccount {
					vacc, ok := postClaimAcc.(*vesting.PeriodicVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				}

				// Check that each claim reward coin's amount has been reset to 0
				claim, found := suite.keeper.GetHardLiquidityProviderClaim(runCtx, userAddr)
				suite.Require().True(found)
				for _, claimRewardCoin := range claim.Reward {
					suite.Require().Equal(c(claimRewardCoin.Denom, 0), claimRewardCoin)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *HandlerTestSuite) TestPayoutHardLiquidityProviderClaimVVesting() {
	type args struct {
		deposit          sdk.Coins
		borrow           sdk.Coins
		rewardsPerSecond sdk.Coins
		multipliers      types.Multipliers
		multiplier       string
		timeElapsed      int64
		expectedRewards  sdk.Coins
		expectedPeriods  vesting.Periods
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type test struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []test{
		{
			"single reward denom: valid 1 day",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      86400,
				expectedRewards:  cs(c("hard", 21142771202)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"single reward denom: valid 10 days",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      864000,
				expectedRewards:  cs(c("hard", 211427712008)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 1 day",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      86400,
				expectedRewards:  cs(c("hard", 21142771202), c("ukava", 21142771202)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 21142771202))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 10 days",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      864000,
				expectedRewards:  cs(c("hard", 211427712008), c("ukava", 211427712008)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008), c("ukava", 211427712008))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms with different rewards per second: valid 1 day",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 222222)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      86400,
				expectedRewards:  cs(c("hard", 21142771202), c("ukava", 38399961603)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 38399961603))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			userAddr := suite.addrs[3]

			bacc := auth.NewBaseAccount(userAddr, cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)), nil, 0, 0)
			bva, err := vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), suite.genesisTime.Unix()+16)
			suite.Require().NoError(err)
			periods := vesting.Periods{
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
			}
			vva := validatorvesting.NewValidatorVestingAccountRaw(bva, suite.genesisTime.Unix(), periods, sdk.ConsAddress{}, nil, 90)

			authBulder := app.NewAuthGenesisBuilder().
				WithAccounts(vva).
				WithSimpleAccount(suite.addrs[2], cs()).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1000000000000000000), c("ukava", 1000000000000000000)))

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithMultipliers(tc.args.multipliers)
			for _, c := range tc.args.deposit {
				incentBuilder = incentBuilder.WithSimpleSupplyRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}
			for _, c := range tc.args.borrow {
				incentBuilder = incentBuilder.WithSimpleBorrowRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			ak := suite.app.GetAccountKeeper()
			hardKeeper := suite.app.GetHardKeeper()

			// User deposits and borrows
			err = hardKeeper.Deposit(suite.ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)
			err = hardKeeper.Borrow(suite.ctx, userAddr, tc.args.borrow)
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim that has 0 rewards
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			for _, coin := range tc.args.deposit {
				suite.Require().Equal(sdk.ZeroInt(), claim.Reward.AmountOf(coin.Denom))
			}

			// Set up future runtime context
			runAtTime := time.Unix(suite.ctx.BlockTime().Unix()+(tc.args.timeElapsed), 0)
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			// Run Hard begin blocker
			hard.BeginBlocker(runCtx, suite.hardKeeper)

			// Accumulate supply rewards for each deposit denom
			for _, coin := range tc.args.deposit {
				rewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriods(runCtx, coin.Denom)
				suite.Require().True(found)
				err = suite.keeper.AccumulateHardSupplyRewards(runCtx, rewardPeriod)
				suite.Require().NoError(err)
			}

			// Accumulate borrow rewards for each deposit denom
			for _, coin := range tc.args.borrow {
				rewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriods(runCtx, coin.Denom)
				suite.Require().True(found)
				err = suite.keeper.AccumulateHardBorrowRewards(runCtx, rewardPeriod)
				suite.Require().NoError(err)
			}

			// Sync hard supply rewards
			deposit, found := suite.hardKeeper.GetDeposit(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

			// Sync hard borrow rewards
			borrow, found := suite.hardKeeper.GetBorrow(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

			// Fetch pre-claim balances
			preClaimAcc := ak.GetAccount(runCtx, suite.addrs[2])

			msg := types.NewMsgClaimHardRewardVVesting(userAddr, suite.addrs[2], tc.args.multiplier)
			_, err = suite.handler(runCtx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check that user's balance has increased by expected reward amount
				postClaimAcc := ak.GetAccount(suite.ctx, suite.addrs[2])
				suite.Require().Equal(preClaimAcc.GetCoins().Add(tc.args.expectedRewards...), postClaimAcc.GetCoins())

				vacc, ok := postClaimAcc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)

				// Check that each claim reward coin's amount has been reset to 0
				claim, found := suite.keeper.GetHardLiquidityProviderClaim(runCtx, suite.addrs[3])
				suite.Require().True(found)
				for _, claimRewardCoin := range claim.Reward {
					suite.Require().Equal(c(claimRewardCoin.Denom, 0), claimRewardCoin)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
