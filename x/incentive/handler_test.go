package incentive_test

import (
	"errors"
	"fmt"
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

func (suite HandlerTestSuite) SwapRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.keeper.GetSwapClaim(suite.ctx, owner)
	suite.Require().Truef(found, "expected swap claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected swap claim reward to be %s, but got %s", expected, claim.Reward)
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
	failMsg := types.NewMsgClaimSwapRewardVVesting(userAddr, suite.addrs[2], "large", nil)
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	msg := types.NewMsgClaimSwapReward(userAddr, "large", nil)
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
	failMsg := types.NewMsgClaimSwapReward(valAddr, "large", nil)
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	msg := types.NewMsgClaimSwapRewardVVesting(valAddr, receiverAddr, "large", nil)
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
