package testutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/committee"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	"github.com/kava-labs/kava/x/hard"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	incentivekeeper "github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/swap"
	swaptypes "github.com/kava-labs/kava/x/swap/types"
)

type IntegrationTester struct {
	suite.Suite
	App app.TestApp
	Ctx sdk.Context
}

func (suite *IntegrationTester) SetupSuite() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
}

func (suite *IntegrationTester) SetApp() {
	suite.App = app.NewTestApp()
}

func (suite *IntegrationTester) StartChain(genesisTime time.Time, genesisStates ...app.GenesisState) {
	suite.App.InitializeFromGenesisStatesWithTime(
		genesisTime,
		genesisStates...,
	)

	suite.Ctx = suite.App.NewContext(false, tmproto.Header{Height: 1, Time: genesisTime})
}

func (suite *IntegrationTester) NextBlockAt(blockTime time.Time) {
	if !suite.Ctx.BlockTime().Before(blockTime) {
		panic(fmt.Sprintf("new block time %s must be after current %s", blockTime, suite.Ctx.BlockTime()))
	}
	blockHeight := suite.Ctx.BlockHeight() + 1

	_ = suite.App.EndBlocker(suite.Ctx, abcitypes.RequestEndBlock{})

	suite.Ctx = suite.Ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)

	_ = suite.App.BeginBlocker(suite.Ctx, abcitypes.RequestBeginBlock{}) // height and time in RequestBeginBlock are ignored by module begin blockers
}

func (suite *IntegrationTester) NextBlockAfter(blockDuration time.Duration) {
	suite.NextBlockAt(suite.Ctx.BlockTime().Add(blockDuration))
}

func (suite *IntegrationTester) DeliverIncentiveMsg(msg sdk.Msg) error {
	msgServer := incentivekeeper.NewMsgServerImpl(suite.App.GetIncentiveKeeper())

	var err error

	switch msg := msg.(type) {
	case *types.MsgClaimHardReward:
		_, err = msgServer.ClaimHardReward(sdk.WrapSDKContext(suite.Ctx), msg)
	case *types.MsgClaimSwapReward:
		_, err = msgServer.ClaimSwapReward(sdk.WrapSDKContext(suite.Ctx), msg)
	case *types.MsgClaimUSDXMintingReward:
		_, err = msgServer.ClaimUSDXMintingReward(sdk.WrapSDKContext(suite.Ctx), msg)
	case *types.MsgClaimDelegatorReward:
		_, err = msgServer.ClaimDelegatorReward(sdk.WrapSDKContext(suite.Ctx), msg)
	default:
		panic("unhandled incentive msg")
	}

	return err
}

func (suite *IntegrationTester) DeliverMsgCreateValidator(address sdk.ValAddress, selfDelegation sdk.Coin) error {
	msg, err := stakingtypes.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1_000_000),
	)
	if err != nil {
		return err
	}

	msgServer := stakingkeeper.NewMsgServerImpl(suite.App.GetStakingKeeper())
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(suite.Ctx), msg)

	return err
}

func (suite *IntegrationTester) DeliverMsgDelegate(delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) error {
	msg := stakingtypes.NewMsgDelegate(
		delegator,
		validator,
		amount,
	)
	msgServer := stakingkeeper.NewMsgServerImpl(suite.App.GetStakingKeeper())
	_, err := msgServer.Delegate(sdk.WrapSDKContext(suite.Ctx), msg)
	return err
}

func (suite *IntegrationTester) DeliverSwapMsgDeposit(depositor sdk.AccAddress, tokenA, tokenB sdk.Coin, slippage sdk.Dec) error {
	msg := swaptypes.NewMsgDeposit(
		depositor.String(),
		tokenA,
		tokenB,
		slippage,
		suite.Ctx.BlockTime().Add(time.Hour).Unix(), // ensure msg will not fail due to short deadline
	)
	msgServer := swap.NewMsgServerImpl(suite.App.GetSwapKeeper())
	_, err := msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), msg)

	return err
}

func (suite *IntegrationTester) DeliverHardMsgDeposit(owner sdk.AccAddress, deposit sdk.Coins) error {
	msg := hardtypes.NewMsgDeposit(owner, deposit)
	msgServer := hard.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgBorrow(owner sdk.AccAddress, borrow sdk.Coins) error {
	msg := hardtypes.NewMsgBorrow(owner, borrow)
	msgServer := hard.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Borrow(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgRepay(owner sdk.AccAddress, repay sdk.Coins) error {
	msg := hardtypes.NewMsgRepay(owner, owner, repay)
	msgServer := hard.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Repay(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgWithdraw(owner sdk.AccAddress, withdraw sdk.Coins) error {
	msg := hardtypes.NewMsgWithdraw(owner, withdraw)
	msgServer := hard.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Withdraw(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverMsgCreateCDP(owner sdk.AccAddress, collateral, principal sdk.Coin, collateralType string) error {
	msg := cdptypes.NewMsgCreateCDP(owner, collateral, principal, collateralType)
	msgServer := cdp.NewMsgServerImpl(suite.App.GetCDPKeeper())

	_, err := msgServer.CreateCDP(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverCDPMsgRepay(owner sdk.AccAddress, collateralType string, payment sdk.Coin) error {
	msg := cdptypes.NewMsgRepayDebt(owner, collateralType, payment)
	msgServer := cdp.NewMsgServerImpl(suite.App.GetCDPKeeper())

	_, err := msgServer.RepayDebt(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverCDPMsgBorrow(owner sdk.AccAddress, collateralType string, draw sdk.Coin) error {
	msg := cdptypes.NewMsgDrawDebt(owner, collateralType, draw)
	msgServer := cdp.NewMsgServerImpl(suite.App.GetCDPKeeper())

	_, err := msgServer.DrawDebt(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) ProposeAndVoteOnNewParams(voter sdk.AccAddress, committeeID uint64, changes []proposaltypes.ParamChange) {
	propose, err := committeetypes.NewMsgSubmitProposal(
		proposaltypes.NewParameterChangeProposal(
			"test title",
			"test description",
			changes,
		),
		voter,
		committeeID,
	)
	suite.NoError(err)

	msgServer := committee.NewMsgServerImpl(suite.App.GetCommitteeKeeper())

	res, err := msgServer.SubmitProposal(sdk.WrapSDKContext(suite.Ctx), propose)
	suite.NoError(err)

	proposalID := res.ProposalID
	vote := committeetypes.NewMsgVote(voter, proposalID, committeetypes.VOTE_TYPE_YES)
	_, err = msgServer.Vote(sdk.WrapSDKContext(suite.Ctx), vote)
	suite.NoError(err)
}

func (suite *IntegrationTester) GetAccount(addr sdk.AccAddress) authtypes.AccountI {
	ak := suite.App.GetAccountKeeper()
	return ak.GetAccount(suite.Ctx, addr)
}

func (suite *IntegrationTester) GetModuleAccount(name string) authtypes.ModuleAccountI {
	ak := suite.App.GetAccountKeeper()
	return ak.GetModuleAccount(suite.Ctx, name)
}

func (suite *IntegrationTester) GetBalance(address sdk.AccAddress) sdk.Coins {
	bk := suite.App.GetBankKeeper()
	return bk.GetAllBalances(suite.Ctx, address)
}

func (suite *IntegrationTester) ErrorIs(err, target error) bool {
	return suite.Truef(errors.Is(err, target), "err didn't match: %s, it was: %s", target, err)
}

func (suite *IntegrationTester) BalanceEquals(address sdk.AccAddress, expected sdk.Coins) {
	bk := suite.App.GetBankKeeper()
	suite.Equalf(
		expected,
		bk.GetAllBalances(suite.Ctx, address),
		"expected account balance to equal coins %s, but got %s",
		expected,
		bk.GetAllBalances(suite.Ctx, address),
	)
}

func (suite *IntegrationTester) BalanceInEpsilon(address sdk.AccAddress, expected sdk.Coins, epsilon float64) {
	actual := suite.GetBalance(address)

	allDenoms := expected.Add(actual...)
	for _, coin := range allDenoms {
		suite.InEpsilonf(
			expected.AmountOf(coin.Denom).Int64(),
			actual.AmountOf(coin.Denom).Int64(),
			epsilon,
			"expected balance to be within %f%% of coins %s, but got %s", epsilon*100, expected, actual,
		)
	}
}

func (suite *IntegrationTester) VestingPeriodsEqual(address sdk.AccAddress, expectedPeriods vestingtypes.Periods) {
	acc := suite.App.GetAccountKeeper().GetAccount(suite.Ctx, address)
	suite.Require().NotNil(acc, "expected vesting account not to be nil")
	vacc, ok := acc.(*vestingtypes.PeriodicVestingAccount)
	suite.Require().True(ok, "expected vesting account to be type PeriodicVestingAccount")
	suite.ElementsMatch(expectedPeriods, vacc.VestingPeriods)
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

func (suite *IntegrationTester) USDXRewardEquals(owner sdk.AccAddress, expected sdk.Coin) {
	claim, found := suite.App.GetIncentiveKeeper().GetUSDXMintingClaim(suite.Ctx, owner)
	suite.Require().Truef(found, "expected delegator claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected delegator claim reward to be %s, but got %s", expected, claim.Reward)
}
