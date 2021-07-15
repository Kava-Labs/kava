package testutil

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/staking"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/swap"
)

type IntegrationTester struct {
	suite.Suite
	App app.TestApp
	Ctx sdk.Context
}

func (suite *IntegrationTester) NextBlockAt(blockTime time.Time) {
	if !suite.Ctx.BlockTime().Before(blockTime) {
		panic(fmt.Sprintf("new block time %s must be after current %s", blockTime, suite.Ctx.BlockTime()))
	}
	blockHeight := suite.Ctx.BlockHeight() + 1

	_ = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{})

	suite.Ctx = suite.Ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)

	_ = suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{}) // height and time in RequestBeginBlock are ignored by module begin blockers
}

func (suite *IntegrationTester) NextBlockAfter(blockDuration time.Duration) {
	suite.NextBlockAt(suite.Ctx.BlockTime().Add(blockDuration))
}

func (suite *IntegrationTester) DeliverIncentiveMsg(msg sdk.Msg) error {
	handler := incentive.NewHandler(suite.App.GetIncentiveKeeper())
	_, err := handler(suite.Ctx, msg)
	return err
}

func (suite *IntegrationTester) DeliverMsgCreateValidator(address sdk.ValAddress, selfDelegation sdk.Coin) error {
	msg := staking.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		staking.Description{},
		staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1_000_000),
	)
	handler := staking.NewHandler(suite.App.GetStakingKeeper())
	_, err := handler(suite.Ctx, msg)
	return err
}

func (suite *IntegrationTester) DeliverMsgDelegate(delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) error {
	msg := staking.NewMsgDelegate(
		delegator,
		validator,
		amount,
	)
	handleStakingMsg := staking.NewHandler(suite.App.GetStakingKeeper())
	_, err := handleStakingMsg(suite.Ctx, msg)
	return err
}

func (suite *IntegrationTester) DeliverSwapMsgDeposit(depositor sdk.AccAddress, tokenA, tokenB sdk.Coin, slippage sdk.Dec) error {
	msg := swap.NewMsgDeposit(
		depositor,
		tokenA,
		tokenB,
		slippage,
		suite.Ctx.BlockTime().Add(time.Hour).Unix(), // ensure msg will not fail due to short deadline
	)
	_, err := swap.NewHandler(suite.App.GetSwapKeeper())(suite.Ctx, msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgDeposit(depositor sdk.AccAddress, deposit sdk.Coins) error {
	msg := hard.NewMsgDeposit(depositor, deposit)
	_, err := hard.NewHandler(suite.App.GetHardKeeper())(suite.Ctx, msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgBorrow(depositor sdk.AccAddress, borrow sdk.Coins) error {
	msg := hard.NewMsgBorrow(depositor, borrow)
	_, err := hard.NewHandler(suite.App.GetHardKeeper())(suite.Ctx, msg)
	return err
}

func (suite *IntegrationTester) GetAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.App.GetAccountKeeper()
	return ak.GetAccount(suite.Ctx, addr)
}

func (suite *IntegrationTester) GetModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.App.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.Ctx, name)
}

func (suite *IntegrationTester) GetBalance(address sdk.AccAddress) sdk.Coins {
	acc := suite.App.GetAccountKeeper().GetAccount(suite.Ctx, address)
	if acc != nil {
		return acc.GetCoins()
	} else {
		return nil
	}
}

func (suite *IntegrationTester) ErrorIs(err, target error) bool {
	return suite.Truef(errors.Is(err, target), "err didn't match: %s, it was: %s", target, err)
}

func (suite *IntegrationTester) BalanceEquals(address sdk.AccAddress, expected sdk.Coins) {
	acc := suite.App.GetAccountKeeper().GetAccount(suite.Ctx, address)
	suite.Require().NotNil(acc, "expected account to not be nil")
	suite.Equalf(expected, acc.GetCoins(), "expected account balance to equal coins %s, but got %s", expected, acc.GetCoins())
}

func (suite *IntegrationTester) VestingPeriodsEqual(address sdk.AccAddress, expectedPeriods vesting.Periods) {
	acc := suite.App.GetAccountKeeper().GetAccount(suite.Ctx, address)
	suite.Require().NotNil(acc, "expected vesting account not to be nil")
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.Require().True(ok, "expected vesting account to be type PeriodicVestingAccount")
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
}

func (suite *IntegrationTester) SwapRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.App.GetIncentiveKeeper().GetSwapClaim(suite.Ctx, owner)
	suite.Require().Truef(found, "expected swap claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected swap claim reward to be %s, but got %s", expected, claim.Reward)
}

func (suite *IntegrationTester) DelegatorRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.App.GetIncentiveKeeper().GetDelegatorClaim(suite.Ctx, owner)
	suite.Require().Truef(found, "expected delegator claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected delegator claim reward to be %s, but got %s", expected, claim.Reward)
}

func (suite *IntegrationTester) HardRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.App.GetIncentiveKeeper().GetHardLiquidityProviderClaim(suite.Ctx, owner)
	suite.Require().Truef(found, "expected delegator claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected delegator claim reward to be %s, but got %s", expected, claim.Reward)
}
