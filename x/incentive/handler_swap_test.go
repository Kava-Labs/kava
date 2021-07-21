package incentive_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

const secondsPerDay = 24 * 60 * 60

// Test suite used for all keeper tests
type HandlerTestSuite struct {
	testutil.IntegrationTester

	// TODO remove these
	keeper     keeper.Keeper
	hardKeeper hardkeeper.Keeper
	cdpKeeper  cdpkeeper.Keeper
	handler    sdk.Handler

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
	suite.App = app.NewTestApp()

	suite.keeper = suite.App.GetIncentiveKeeper()
	suite.hardKeeper = suite.App.GetHardKeeper()
	suite.cdpKeeper = suite.App.GetCDPKeeper()
	suite.handler = incentive.NewHandler(suite.keeper)

	suite.Ctx = suite.App.NewContext(true, abci.Header{Height: 1, Time: suite.genesisTime})
}

type genesisBuilder interface {
	BuildMarshalled() app.GenesisState
}

func (suite *HandlerTestSuite) SetupWithGenState(builders ...genesisBuilder) {
	suite.SetupApp()

	builtGenStates := []app.GenesisState{
		NewStakingGenesisState(),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		NewHardGenStateMulti(suite.genesisTime).BuildMarshalled(),
		NewSwapGenesisState(),
	}
	for _, builder := range builders {
		builtGenStates = append(builtGenStates, builder.BuildMarshalled())
	}

	suite.App.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		builtGenStates...,
	)
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

// authBuilder returns a new auth genesis builder with a full kavadist module account.
func (suite *HandlerTestSuite) authBuilder() app.AuthGenesisBuilder {
	return app.NewAuthGenesisBuilder().
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c(types.USDXMintingRewardDenom, 1e18), c("hard", 1e18), c("swap", 1e18)))
}

// incentiveBuilder returns a new incentive genesis builder with a genesis time and multipliers set
func (suite *HandlerTestSuite) incentiveBuilder() testutil.IncentiveGenesisBuilder {
	return testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("small"), 1, d("0.2")),
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
		}).
		WithMultipliersTODO(types.MultipliersPerDenom{
			{
				Denom: "hard",
				Multipliers: types.Multipliers{
					types.NewMultiplier(types.MultiplierName("small"), 1, d("0.2")),
					types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
				},
			},
			{
				Denom: "swap",
				Multipliers: types.Multipliers{
					types.NewMultiplier(types.MultiplierName("medium"), 6, d("0.5")),
					types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
				},
			},
			{
				Denom: "ukava",
				Multipliers: types.Multipliers{
					types.NewMultiplier(types.MultiplierName("small"), 1, d("0.2")),
					types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
				},
			},
		})
}

func (suite *HandlerTestSuite) TestPayoutSwapClaim() {
	userAddr := suite.addrs[0]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12), c("busd", 1e12)))

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSwapRewardPeriod("busd:ukava", cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// deposit into a swap pool
	suite.NoError(
		suite.DeliverSwapMsgDeposit(userAddr, c("ukava", 1e9), c("busd", 1e9), d("1.0")),
	)
	// accumulate some swap rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapRewardVVesting(userAddr, suite.addrs[1], "large", nil),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapReward(userAddr, "large", nil),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := cs(c("swap", 7*1e6), c("hard", 7*1e6))
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards...))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004793, Amount: expectedRewards},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.SwapRewardEquals(userAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutSwapClaimSingleDenom() {
	userAddr := suite.addrs[0]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12), c("busd", 1e12)))

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSwapRewardPeriod("busd:ukava", cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// deposit into a swap pool
	suite.NoError(
		suite.DeliverSwapMsgDeposit(userAddr, c("ukava", 1e9), c("busd", 1e9), d("1.0")),
	)

	// accumulate some swap rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapRewardVVesting(userAddr, suite.addrs[1], "large", nil),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapReward(userAddr, "large", []string{"swap"}),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.SwapRewardEquals(userAddr, cs(c("hard", 7*1e6)))
}

func (suite *HandlerTestSuite) TestPayoutSwapClaimVVesting() {
	valAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("ukava", 1e12), c("busd", 1e12)))

	authBulder := suite.authBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSwapRewardPeriod("busd:ukava", cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// deposit into a swap pool
	suite.NoError(
		suite.DeliverSwapMsgDeposit(valAddr, c("ukava", 1e9), c("busd", 1e9), d("1.0")),
	)

	// accumulate some swap rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapReward(valAddr, "large", nil),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapRewardVVesting(valAddr, receiverAddr, "large", nil),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := cs(c("hard", 7*1e6), c("swap", 7*1e6))
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards...))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004793, Amount: expectedRewards},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.SwapRewardEquals(valAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutSwapClaimVVestingSingleDenom() {
	valAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("ukava", 1e12), c("busd", 1e12)))

	authBulder := suite.authBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSwapRewardPeriod("busd:ukava", cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// deposit into a swap pool
	suite.NoError(
		suite.DeliverSwapMsgDeposit(valAddr, c("ukava", 1e9), c("busd", 1e9), d("1.0")),
	)

	// accumulate some swap rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapReward(valAddr, "large", []string{"swap"}),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimSwapRewardVVesting(valAddr, receiverAddr, "large", []string{"swap"}),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.SwapRewardEquals(valAddr, cs(c("hard", 7*1e6)))
}
