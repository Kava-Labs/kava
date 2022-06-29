package testutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	committeekeeper "github.com/kava-labs/kava/x/committee/keeper"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	earnkeeper "github.com/kava-labs/kava/x/earn/keeper"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	incentivekeeper "github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	liquidkeeper "github.com/kava-labs/kava/x/liquid/keeper"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
	routerkeeper "github.com/kava-labs/kava/x/router/keeper"
	routertypes "github.com/kava-labs/kava/x/router/types"
	swapkeeper "github.com/kava-labs/kava/x/swap/keeper"
	swaptypes "github.com/kava-labs/kava/x/swap/types"
)

var testChainID = "kavatest_1-1"

type IntegrationTester struct {
	suite.Suite
	App app.TestApp
	Ctx sdk.Context

	GenesisTime time.Time
}

func (suite *IntegrationTester) SetupSuite() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	// Default genesis time, can be overridden with WithGenesisTime
	suite.GenesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *IntegrationTester) SetApp() {
	suite.App = app.NewTestApp()
}

func (suite *IntegrationTester) SetupTest() {
	suite.SetApp()
}

func (suite *IntegrationTester) WithGenesisTime(genesisTime time.Time) {
	suite.GenesisTime = genesisTime
}

func (suite *IntegrationTester) StartChainWithBuilders(builders ...GenesisBuilder) {
	var builtGenStates []app.GenesisState
	for _, builder := range builders {
		builtGenStates = append(builtGenStates, builder.BuildMarshalled(suite.App.AppCodec()))
	}

	suite.StartChain(builtGenStates...)
}

func (suite *IntegrationTester) StartChain(genesisStates ...app.GenesisState) {
	suite.App.InitializeFromGenesisStatesWithTimeAndChainID(
		suite.GenesisTime,
		testChainID,
		genesisStates...,
	)

	suite.Ctx = suite.App.NewContext(false, tmproto.Header{
		Height:  1,
		Time:    suite.GenesisTime,
		ChainID: testChainID,
	})
}

func (suite *IntegrationTester) NextBlockAfter(blockDuration time.Duration) {
	suite.NextBlockAfterWithReq(
		blockDuration,
		abcitypes.RequestEndBlock{},
		abcitypes.RequestBeginBlock{},
	)
}

func (suite *IntegrationTester) NextBlockAfterWithReq(
	blockDuration time.Duration,
	reqEnd abcitypes.RequestEndBlock,
	reqBegin abcitypes.RequestBeginBlock,
) (abcitypes.ResponseEndBlock, abcitypes.ResponseBeginBlock) {
	return suite.NextBlockAtWithRequest(
		suite.Ctx.BlockTime().Add(blockDuration),
		reqEnd,
		reqBegin,
	)
}

func (suite *IntegrationTester) NextBlockAt(
	blockTime time.Time,
) (abcitypes.ResponseEndBlock, abcitypes.ResponseBeginBlock) {
	return suite.NextBlockAtWithRequest(
		blockTime,
		abcitypes.RequestEndBlock{},
		abcitypes.RequestBeginBlock{},
	)
}

func (suite *IntegrationTester) NextBlockAtWithRequest(
	blockTime time.Time,
	reqEnd abcitypes.RequestEndBlock,
	reqBegin abcitypes.RequestBeginBlock,
) (abcitypes.ResponseEndBlock, abcitypes.ResponseBeginBlock) {
	if !suite.Ctx.BlockTime().Before(blockTime) {
		panic(fmt.Sprintf("new block time %s must be after current %s", blockTime, suite.Ctx.BlockTime()))
	}
	blockHeight := suite.Ctx.BlockHeight() + 1

	responseEndBlock := suite.App.EndBlocker(suite.Ctx, reqEnd)
	suite.Ctx = suite.Ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight).WithChainID(testChainID)
	responseBeginBlock := suite.App.BeginBlocker(suite.Ctx, reqBegin) // height and time in RequestBeginBlock are ignored by module begin blockers

	return responseEndBlock, responseBeginBlock
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
	case *types.MsgClaimEarnReward:
		_, err = msgServer.ClaimEarnReward(sdk.WrapSDKContext(suite.Ctx), msg)
	default:
		panic("unhandled incentive msg")
	}

	return err
}

// MintLiquidAnyValAddr mints liquid tokens with the given validator address,
// creating the validator if it does not already exist.
// **Note:** This will increment the block height/time and run the End and Begin
// blockers!
func (suite *IntegrationTester) MintLiquidAnyValAddr(
	owner sdk.AccAddress,
	validator sdk.ValAddress,
	amount sdk.Coin,
) (sdk.Coin, error) {
	// Check if validator already created
	_, found := suite.App.GetStakingKeeper().GetValidator(suite.Ctx, validator)
	if !found {
		// Create validator
		if err := suite.DeliverMsgCreateValidator(validator, sdk.NewCoin("ukava", sdk.NewInt(1e9))); err != nil {
			return sdk.Coin{}, err
		}

		// new block required to bond validator
		suite.NextBlockAfter(7 * time.Second)
	}

	// Delegate and mint liquid tokens
	return suite.DeliverMsgDelegateMint(owner, validator, amount)
}

func (suite *IntegrationTester) GetAbciValidator(valAddr sdk.ValAddress) abcitypes.Validator {
	sk := suite.App.GetStakingKeeper()

	val, found := sk.GetValidator(suite.Ctx, valAddr)
	suite.Require().True(found)

	pk, err := val.ConsPubKey()
	suite.Require().NoError(err)

	return abcitypes.Validator{
		Address: pk.Address(),
		Power:   val.GetConsensusPower(sk.PowerReduction(suite.Ctx)),
	}
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
	msgServer := swapkeeper.NewMsgServerImpl(suite.App.GetSwapKeeper())
	_, err := msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), msg)

	return err
}

func (suite *IntegrationTester) DeliverHardMsgDeposit(owner sdk.AccAddress, deposit sdk.Coins) error {
	msg := hardtypes.NewMsgDeposit(owner, deposit)
	msgServer := hardkeeper.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgBorrow(owner sdk.AccAddress, borrow sdk.Coins) error {
	msg := hardtypes.NewMsgBorrow(owner, borrow)
	msgServer := hardkeeper.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Borrow(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgRepay(owner sdk.AccAddress, repay sdk.Coins) error {
	msg := hardtypes.NewMsgRepay(owner, owner, repay)
	msgServer := hardkeeper.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Repay(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverHardMsgWithdraw(owner sdk.AccAddress, withdraw sdk.Coins) error {
	msg := hardtypes.NewMsgWithdraw(owner, withdraw)
	msgServer := hardkeeper.NewMsgServerImpl(suite.App.GetHardKeeper())

	_, err := msgServer.Withdraw(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverMsgCreateCDP(owner sdk.AccAddress, collateral, principal sdk.Coin, collateralType string) error {
	msg := cdptypes.NewMsgCreateCDP(owner, collateral, principal, collateralType)
	msgServer := cdpkeeper.NewMsgServerImpl(suite.App.GetCDPKeeper())

	_, err := msgServer.CreateCDP(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverCDPMsgRepay(owner sdk.AccAddress, collateralType string, payment sdk.Coin) error {
	msg := cdptypes.NewMsgRepayDebt(owner, collateralType, payment)
	msgServer := cdpkeeper.NewMsgServerImpl(suite.App.GetCDPKeeper())

	_, err := msgServer.RepayDebt(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverCDPMsgBorrow(owner sdk.AccAddress, collateralType string, draw sdk.Coin) error {
	msg := cdptypes.NewMsgDrawDebt(owner, collateralType, draw)
	msgServer := cdpkeeper.NewMsgServerImpl(suite.App.GetCDPKeeper())

	_, err := msgServer.DrawDebt(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverMsgMintDerivative(
	sender sdk.AccAddress,
	validator sdk.ValAddress,
	amount sdk.Coin,
) (sdk.Coin, error) {
	msg := liquidtypes.NewMsgMintDerivative(sender, validator, amount)
	msgServer := liquidkeeper.NewMsgServerImpl(suite.App.GetLiquidKeeper())

	res, err := msgServer.MintDerivative(sdk.WrapSDKContext(suite.Ctx), &msg)
	if err != nil {
		// Instead of returning res.Received, as res will be nil if there is an error
		return sdk.Coin{}, err
	}

	return res.Received, err
}

func (suite *IntegrationTester) DeliverEarnMsgDeposit(
	depositor sdk.AccAddress,
	amount sdk.Coin,
	strategy earntypes.StrategyType,
) error {
	msg := earntypes.NewMsgDeposit(depositor.String(), amount, strategy)
	msgServer := earnkeeper.NewMsgServerImpl(suite.App.GetEarnKeeper())

	_, err := msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), msg)
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

	msgServer := committeekeeper.NewMsgServerImpl(suite.App.GetCommitteeKeeper())

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

func (suite *IntegrationTester) VestingPeriodsEqual(address sdk.AccAddress, expectedPeriods []vestingtypes.Period) {
	acc := suite.App.GetAccountKeeper().GetAccount(suite.Ctx, address)
	suite.Require().NotNil(acc, "expected vesting account not to be nil")
	vacc, ok := acc.(*vestingtypes.PeriodicVestingAccount)
	suite.Require().True(ok, "expected vesting account to be type PeriodicVestingAccount")
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
}

// -----------------------------------------------------------------------------
// x/incentive

func (suite *IntegrationTester) SwapRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.App.GetIncentiveKeeper().GetClaim(suite.Ctx, types.RewardTypeSwap, owner)
	suite.Require().Truef(found, "expected swap claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected swap claim reward to be %s, but got %s", expected, claim.Reward)
}

func (suite *IntegrationTester) DelegatorRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.App.GetIncentiveKeeper().GetClaim(suite.Ctx, types.RewardTypeDelegator, owner)
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

func (suite *IntegrationTester) EarnRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.App.GetIncentiveKeeper().GetEarnClaim(suite.Ctx, owner)
	suite.Require().Truef(found, "expected earn claim to be found for %s", owner)
	suite.Truef(expected.IsEqual(claim.Reward), "expected earn claim reward to be %s, but got %s", expected, claim.Reward)
}

// AddTestAddrsFromPubKeys adds the addresses into the SimApp providing only the public keys.
func (suite *IntegrationTester) AddTestAddrsFromPubKeys(ctx sdk.Context, pubKeys []cryptotypes.PubKey, accAmt sdk.Int) {
	initCoins := sdk.NewCoins(sdk.NewCoin(suite.App.GetStakingKeeper().BondDenom(ctx), accAmt))

	for _, pk := range pubKeys {
		suite.App.FundAccount(ctx, sdk.AccAddress(pk.Address()), initCoins)
	}
}

func (suite *IntegrationTester) StoredEarnTimeEquals(denom string, expected time.Time) {
	storedTime, found := suite.App.GetIncentiveKeeper().GetEarnRewardAccrualTime(suite.Ctx, denom)
	suite.Equal(found, expected != time.Time{}, "expected time is %v but time found = %v", expected, found)
	if found {
		suite.Equal(expected, storedTime)
	} else {
		suite.Empty(storedTime)
	}
}

func (suite *IntegrationTester) StoredEarnIndexesEqual(denom string, expected types.RewardIndexes) {
	storedIndexes, found := suite.App.GetIncentiveKeeper().GetEarnRewardIndexes(suite.Ctx, denom)
	suite.Equal(found, expected != nil)

	if found {
		suite.Equal(expected, storedIndexes)
	} else {
		// Can't compare Equal for types.RewardIndexes(nil) vs types.RewardIndexes{}
		suite.Empty(storedIndexes)
	}
}

func (suite *IntegrationTester) AddIncentiveEarnMultiRewardPeriod(period types.MultiRewardPeriod) {
	ik := suite.App.GetIncentiveKeeper()
	params := ik.GetParams(suite.Ctx)

	for i, reward := range params.EarnRewardPeriods {
		if reward.CollateralType == period.CollateralType {
			// Replace existing reward period if the collateralType exists.
			// Params are invalid if there are multiple reward periods for the
			// same collateral type.
			params.EarnRewardPeriods[i] = period
			ik.SetParams(suite.Ctx, params)
			return
		}
	}

	params.EarnRewardPeriods = append(params.EarnRewardPeriods, period)

	suite.NoError(params.Validate())
	ik.SetParams(suite.Ctx, params)
}

// -----------------------------------------------------------------------------
// x/router

func (suite *IntegrationTester) DeliverRouterMsgDelegateMintDeposit(
	depositor sdk.AccAddress,
	validator sdk.ValAddress,
	amount sdk.Coin,
) error {
	msg := routertypes.MsgDelegateMintDeposit{
		Depositor: depositor.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
	msgServer := routerkeeper.NewMsgServerImpl(suite.App.GetRouterKeeper())

	_, err := msgServer.DelegateMintDeposit(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverRouterMsgMintDeposit(
	depositor sdk.AccAddress,
	validator sdk.ValAddress,
	amount sdk.Coin,
) error {
	msg := routertypes.MsgMintDeposit{
		Depositor: depositor.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
	msgServer := routerkeeper.NewMsgServerImpl(suite.App.GetRouterKeeper())

	_, err := msgServer.MintDeposit(sdk.WrapSDKContext(suite.Ctx), &msg)
	return err
}

func (suite *IntegrationTester) DeliverMsgDelegateMint(
	delegator sdk.AccAddress,
	validator sdk.ValAddress,
	amount sdk.Coin,
) (sdk.Coin, error) {
	if err := suite.DeliverMsgDelegate(delegator, validator, amount); err != nil {
		return sdk.Coin{}, err
	}

	return suite.DeliverMsgMintDerivative(delegator, validator, amount)
}

// -----------------------------------------------------------------------------
// x/distribution

func (suite *IntegrationTester) GetBeginBlockClaimedStakingRewards(
	resBeginBlock abcitypes.ResponseBeginBlock,
) (validatorRewards map[string]sdk.Coins, totalRewards sdk.Coins) {
	// Events emitted in BeginBlocker are in the ResponseBeginBlock, not in
	// ctx.EventManager().Events() as BeginBlock is called with a NewEventManager()
	// cosmos-sdk/types/module/module.go: func(m *Manager) BeginBlock(...)

	// We also need to parse the events to get the rewards as querying state will
	// always contain 0 rewards -- rewards are always claimed right after
	// mint+distribution in BeginBlocker which resets distribution state back to
	// 0 for reward amounts
	blockRewardsClaimed := make(map[string]sdk.Coins)
	for _, event := range resBeginBlock.Events {
		if event.Type != distributiontypes.EventTypeWithdrawRewards {
			continue
		}

		// Example event attributes, amount can be empty for no rewards
		//
		// Event: withdraw_rewards
		// - amount:
		// - validator: kavavaloper1em2mlkrkx0qsa6327tgvl3g0fh8a95hjnqvrwh
		// Event: withdraw_rewards
		// - amount: 523909ukava
		// - validator: kavavaloper1nmgpgr8l4t8pw9zqx9cltuymvz85wmw9sy8kjy
		attrsMap := attrsToMap(event.Attributes)

		validator, found := attrsMap[distributiontypes.AttributeKeyValidator]
		suite.Require().Truef(found, "expected validator attribute to be found in event %s", event)

		amountStr, found := attrsMap[sdk.AttributeKeyAmount]
		suite.Require().Truef(found, "expected amount attribute to be found in event %s", event)

		amount := sdk.NewCoins()

		// Only parse amount if it is not empty
		if len(amountStr) > 0 {
			parsedAmt, err := sdk.ParseCoinNormalized(amountStr)
			suite.Require().NoError(err)
			amount = amount.Add(parsedAmt)
		}

		blockRewardsClaimed[validator] = amount
	}

	totalClaimedRewards := sdk.NewCoins()
	for _, amount := range blockRewardsClaimed {
		totalClaimedRewards = totalClaimedRewards.Add(amount...)
	}

	return blockRewardsClaimed, totalClaimedRewards
}

func attrsToMap(attrs []abcitypes.EventAttribute) map[string]string {
	out := make(map[string]string)

	for _, attr := range attrs {
		out[string(attr.Key)] = string(attr.Value)
	}

	return out
}
