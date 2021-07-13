package incentive_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/staking"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	"github.com/kava-labs/kava/x/swap"
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

type GenesisBuilder interface {
	BuildMarshalled() app.GenesisState
}

func (suite *HandlerTestSuite) SetupWithGenState(builders ...GenesisBuilder) {
	suite.SetupApp()

	builtGenStates := []app.GenesisState{
		NewStakingGenesisState(),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		NewSwapGenesisState(),
	}
	for _, builder := range builders {
		builtGenStates = append(builtGenStates, builder.BuildMarshalled())
	}

	suite.app.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		builtGenStates...,
	)
}

func (suite *HandlerTestSuite) NextBlockAt(blockTime time.Time) {
	if !suite.ctx.BlockTime().Before(blockTime) {
		panic(fmt.Sprintf("new block time %s must be after current %s", blockTime, suite.ctx.BlockTime()))
	}
	blockHeight := suite.ctx.BlockHeight() + 1

	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})

	suite.ctx = suite.ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)

	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in RequestBeginBlock are ignored by module begin blockers
}

func (suite *HandlerTestSuite) NextBlockAfter(blockDuration time.Duration) {
	suite.NextBlockAt(suite.ctx.BlockTime().Add(blockDuration))
}

func (suite *HandlerTestSuite) DeliverMsgCreateValidator(address sdk.ValAddress, selfDelegation sdk.Coin) error {
	msg := staking.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		staking.Description{},
		staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1_000_000),
	)
	handleStakingMsg := staking.NewHandler(suite.app.GetStakingKeeper())
	_, err := handleStakingMsg(suite.ctx, msg)
	return err
}

func (suite *HandlerTestSuite) DeliverMsgDelegate(delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) error {
	msg := staking.NewMsgDelegate(
		delegator,
		validator,
		amount,
	)
	handleStakingMsg := staking.NewHandler(suite.app.GetStakingKeeper())
	_, err := handleStakingMsg(suite.ctx, msg)
	return err
}

func (suite *HandlerTestSuite) DeliverSwapMsgDeposit(depositor sdk.AccAddress, tokenA, tokenB sdk.Coin, slippage sdk.Dec) error {
	msg := swap.NewMsgDeposit(
		depositor,
		tokenA,
		tokenB,
		slippage,
		suite.ctx.BlockTime().Add(time.Hour).Unix(), // ensure msg will not fail due to short deadline
	)
	_, err := swap.NewHandler(suite.app.GetSwapKeeper())(suite.ctx, msg)
	return err
}

func (suite *HandlerTestSuite) GetAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *HandlerTestSuite) GetModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *HandlerTestSuite) GetBalance(address sdk.AccAddress) sdk.Coins {
	acc := suite.app.GetAccountKeeper().GetAccount(suite.ctx, address)
	if acc != nil {
		return acc.GetCoins()
	} else {
		return nil
	}
}

// for the purposes of incentive module. A validator vesting account only needs to exist, and have enough balance to delegate/or supply.
func (suite *HandlerTestSuite) NewValidatorVestingAccountWithBalance(address sdk.AccAddress, spendableBalance sdk.Coins) *validatorvesting.ValidatorVestingAccount {
	bacc := auth.NewBaseAccount(address, spendableBalance, nil, 0, 0)
	// vesting coins set to nil and vesting end time set to genesis full base account balance should be spendable
	bva, err := vesting.NewBaseVestingAccount(bacc, nil, suite.genesisTime.Unix())
	if err != nil {
		panic(err.Error())
	}
	// vesting start time set to genesis and no vesting periods
	return validatorvesting.NewValidatorVestingAccountRaw(bva, suite.genesisTime.Unix(), nil, sdk.ConsAddress{}, nil, 90)
}

func (suite *HandlerTestSuite) ErrorIs(err, target error) bool {
	return suite.Truef(errors.Is(err, target), "err didn't match: %s, it was: %s", target, err)
}

func (suite HandlerTestSuite) BalanceEquals(address sdk.AccAddress, expected sdk.Coins) {
	acc := suite.app.GetAccountKeeper().GetAccount(suite.ctx, address)
	suite.Require().NotNil(acc, "expected account to not be nil")
	suite.Equalf(expected, acc.GetCoins(), "expected account balance to equal coins %s, but got %s", expected, acc.GetCoins())
}

func (suite *HandlerTestSuite) VestingPeriodsEqual(address sdk.AccAddress, expectedPeriods vesting.Periods) {
	acc := suite.app.GetAccountKeeper().GetAccount(suite.ctx, address)
	suite.Require().NotNil(acc, "expected vesting account not to be nil")
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.Require().True(ok, "expected vesting account to be type PeriodicVestingAccount")
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
}

func (suite HandlerTestSuite) DelegatorRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.keeper.GetDelegatorClaim(suite.ctx, owner)
	suite.Require().Truef(found, "expected delegator claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected delegator claim reward to be %s, but got %s", expected, claim.Reward)
}

func (suite HandlerTestSuite) SwapRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.keeper.GetSwapClaim(suite.ctx, owner)
	suite.Require().Truef(found, "expected swap claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected swap claim reward to be %s, but got %s", expected, claim.Reward)
}

func (suite *HandlerTestSuite) TestPayoutUSDXMintingClaim() {
	type args struct {
		ctype                    string
		rewardsPerSecond         sdk.Coin
		initialCollateral        sdk.Coin
		initialPrincipal         sdk.Coin
		multipliers              types.Multipliers
		multiplier               string
		timeElapsed              time.Duration
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
				timeElapsed:              86400 * time.Second,
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
				timeElapsed:              86400 * time.Second,
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

			suite.NextBlockAfter(tc.args.timeElapsed)

			msg := incentive.NewMsgClaimUSDXMintingReward(userAddr, tc.args.multiplier)
			_, err = suite.handler(suite.ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				acc := suite.GetAccount(userAddr)
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
		timeElapsed       time.Duration
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
				timeElapsed:       86400 * time.Second,
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
				timeElapsed:       86400 * time.Second,
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
			suite.NextBlockAfter(tc.args.timeElapsed)

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
		timeElapsed              time.Duration
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
				timeElapsed:              86400 * time.Second,
				expectedRewards:          cs(c("hard", 21142771202)), // 10571385600 (deposit reward) + 10571385600 (borrow reward) + 2 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202))}},
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
				timeElapsed:              864000 * time.Second,
				expectedRewards:          cs(c("hard", 211427712008)), // 105713856000 (deposit reward) + 105713856000 (borrow reward) + 8 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008))}},
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
				timeElapsed:              86400 * time.Second,
				expectedRewards:          cs(c("hard", 21142771202), c("ukava", 21142771202)), // 10571385600 (deposit reward) + 10571385600 (borrow reward) + 2 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 21142771202))}},
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
				timeElapsed:              864000 * time.Second,
				expectedRewards:          cs(c("hard", 211427712008), c("ukava", 211427712008)), // 105713856000 (deposit reward) + 105713856000 (borrow reward) + 8 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008), c("ukava", 211427712008))}},
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
				timeElapsed:              86400 * time.Second,
				expectedRewards:          cs(c("hard", 21142771202), c("ukava", 38399961603)),
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 38399961603))}},
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

			// Accumulate supply and borrow rewards
			suite.NextBlockAfter(tc.args.timeElapsed)

			// Fetch pre-claim balances
			ak := suite.app.GetAccountKeeper()
			preClaimAcc := ak.GetAccount(suite.ctx, userAddr)

			msg := types.NewMsgClaimHardReward(userAddr, tc.args.multiplier)
			_, err = suite.handler(suite.ctx, msg)

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
				claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
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
		timeElapsed      time.Duration
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
				timeElapsed:      86400 * time.Second,
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
				timeElapsed:      864000 * time.Second,
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
				timeElapsed:      86400 * time.Second,
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
				timeElapsed:      864000 * time.Second,
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
				timeElapsed:      86400 * time.Second,
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

			// Accumulate supply and borrow rewards
			suite.NextBlockAfter(tc.args.timeElapsed)

			// Fetch pre-claim balances
			preClaimAcc := ak.GetAccount(suite.ctx, suite.addrs[2])

			msg := types.NewMsgClaimHardRewardVVesting(userAddr, suite.addrs[2], tc.args.multiplier)
			_, err = suite.handler(suite.ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check that user's balance has increased by expected reward amount
				postClaimAcc := ak.GetAccount(suite.ctx, suite.addrs[2])
				suite.Require().Equal(preClaimAcc.GetCoins().Add(tc.args.expectedRewards...), postClaimAcc.GetCoins())

				vacc, ok := postClaimAcc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)

				// Check that each claim reward coin's amount has been reset to 0
				claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
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

func (suite *HandlerTestSuite) TestPayoutDelegatorClaim() {
	userAddr := suite.addrs[0]
	receiverAddr := suite.addrs[1]

	authBulder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12))).
		WithSimpleAccount(receiverAddr, cs(c("ukava", 1e12))).
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1e18), c("swap", 1e18)))

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
		}).
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(userAddr), c("ukava", 1e9)),
	)
	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	// preClaimAcc := suite.app.GetAccountKeeper().GetAccount(suite.ctx, valAddr)
	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	failMsg := types.NewMsgClaimDelegatorRewardVVesting(userAddr, receiverAddr, "large")
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	msg := types.NewMsgClaimDelegatorReward(userAddr, "large", nil)
	_, err = suite.handler(suite.ctx, msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := cs(c("hard", 2*7*1e6), c("swap", 2*7*1e6))
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards...))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004786, Amount: expectedRewards},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.DelegatorRewardEquals(userAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutDelegatorClaimSingleDenom() {
	userAddr := suite.addrs[0]

	suite.SetupWithGenState(

		app.NewAuthGenesisBuilder().
			WithSimpleAccount(userAddr, cs(c("ukava", 1e12))).
			WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1e18), c("swap", 1e18))),

		testutil.NewIncentiveGenesisBuilder().
			WithGenesisTime(suite.genesisTime).
			WithMultipliers(types.Multipliers{
				types.NewMultiplier("large", 12, d("1.0")),
			}).
			WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6), c("swap", 1e6))),
	)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(userAddr), c("ukava", 1e9)),
	)
	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Claim rewards
	msg := types.NewMsgClaimDelegatorReward(userAddr, "large", []string{"swap"})
	_, err := suite.handler(suite.ctx, msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 2*7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004786, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.DelegatorRewardEquals(userAddr, cs(c("hard", 2*7*1e6)))
}

func (suite *HandlerTestSuite) TestPayoutDelegatorClaimVVesting() {
	valAddr := suite.addrs[0]
	receiverAddr := suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("ukava", 1e12)))

	authBulder := app.NewAuthGenesisBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil).
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1e18)))

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
		}).
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(valAddr), c("ukava", 1e9)),
	)
	suite.NextBlockAfter(7 * time.Second) // new block required to bond validator

	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	failMsg := types.NewMsgClaimDelegatorReward(valAddr, "large", nil)
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim the delegation rewards
	msg := types.NewMsgClaimDelegatorRewardVVesting(valAddr, receiverAddr, "large")
	_, err = suite.handler(suite.ctx, msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("hard", 2*7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004786, Amount: cs(expectedRewards)},
	})
	// Check that each claim reward coin's amount has been reset to 0
	suite.DelegatorRewardEquals(valAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutSwapClaim() {
	userAddr := suite.addrs[0]

	authBulder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12), c("busd", 1e12))).
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("swap", 1e18)))

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
		}).
		WithSimpleSwapRewardPeriod("busd/ukava", cs(c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// deposit into a swap pool
	suite.NoError(
		suite.DeliverSwapMsgDeposit(userAddr, c("ukava", 1e9), c("busd", 1e9), d("1.0")),
	)

	// accumulate some swap rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	failMsg := types.NewMsgClaimSwapRewardVVesting(userAddr, suite.addrs[2], "large")
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	msg := types.NewMsgClaimSwapReward(userAddr, "large")
	_, err = suite.handler(suite.ctx, msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.SwapRewardEquals(userAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutSwapClaimVVesting() {
	valAddr := suite.addrs[0]
	receiverAddr := suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("ukava", 1e12), c("busd", 1e12)))

	authBulder := app.NewAuthGenesisBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil).
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("swap", 1e18)))

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
		}).
		WithSimpleSwapRewardPeriod("busd/ukava", cs(c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// deposit into a swap pool
	suite.NoError(
		suite.DeliverSwapMsgDeposit(valAddr, c("ukava", 1e9), c("busd", 1e9), d("1.0")),
	)

	// accumulate some swap rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	failMsg := types.NewMsgClaimSwapReward(valAddr, "large")
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	msg := types.NewMsgClaimSwapRewardVVesting(valAddr, receiverAddr, "large")
	_, err = suite.handler(suite.ctx, msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.SwapRewardEquals(valAddr, nil)
}
