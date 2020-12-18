package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest/keeper"
	"github.com/kava-labs/kava/x/harvest/types"
)

func (suite *KeeperTestSuite) TestApplyDepositRewards() {
	type args struct {
		depositor            sdk.AccAddress
		denom                string
		depositAmount        sdk.Coins
		totalDeposits        sdk.Coin
		rewardRate           sdk.Coin
		claimType            types.ClaimType
		previousBlockTime    time.Time
		blockTime            time.Time
		expectedClaimBalance sdk.Coin
	}
	type errArgs struct {
		expectPanic bool
		contains    string
	}
	type testCase struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []testCase{
		{
			name: "distribute rewards",
			args: args{
				depositor:            sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                "bnb",
				rewardRate:           c("hard", 500),
				depositAmount:        cs(c("bnb", 100)),
				totalDeposits:        c("bnb", 1000),
				claimType:            types.LP,
				previousBlockTime:    time.Date(2020, 11, 1, 13, 59, 50, 0, time.UTC),
				blockTime:            time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				expectedClaimBalance: c("hard", 500),
			},
			errArgs: errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tc.args.blockTime})
			loanToValue, _ := sdk.NewDecFromStr("0.6")
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), tc.args.rewardRate, time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), tc.args.rewardRate, time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "usdx:usd", sdk.NewInt(1000000), sdk.NewInt(USDX_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "kava:usd", sdk.NewInt(1000000), sdk.NewInt(KAVA_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				},
				0, // LTV counter
			), tc.args.previousBlockTime, types.DefaultDistributionTimes)
			tApp.InitializeFromGenesisStates(app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, cs(tc.args.totalDeposits))
			keeper := tApp.GetHarvestKeeper()
			deposit := types.NewDeposit(tc.args.depositor, tc.args.depositAmount)
			keeper.SetDeposit(ctx, deposit)
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			if tc.errArgs.expectPanic {
				suite.Require().Panics(func() { suite.keeper.ApplyDepositRewards(suite.ctx) })
			} else {
				suite.Require().NotPanics(func() { suite.keeper.ApplyDepositRewards(suite.ctx) })
				claim, f := suite.keeper.GetClaim(suite.ctx, tc.args.depositor, tc.args.denom, tc.args.claimType)
				suite.Require().True(f)
				suite.Require().Equal(tc.args.expectedClaimBalance, claim.Amount)
			}
		})
	}
}

func TestApplyDelegatorRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(DelegatorRewardsTestSuite))
}

type DelegatorRewardsTestSuite struct {
	suite.Suite

	validatorAddrs []sdk.ValAddress
	delegatorAddrs []sdk.AccAddress

	keeper        keeper.Keeper
	stakingKeeper staking.Keeper
	app           app.TestApp
	rewardRate    int64
}

// The default state used by each test
func (suite *DelegatorRewardsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, allAddrs := app.GeneratePrivKeyAddressPairs(10)
	suite.delegatorAddrs = allAddrs[:5]
	for _, a := range allAddrs[5:] {
		suite.validatorAddrs = append(suite.validatorAddrs, sdk.ValAddress(a))
	}

	suite.app = app.NewTestApp()

	suite.rewardRate = 500

	suite.app.InitializeFromGenesisStates(
		equalCoinsAuthGenState(allAddrs, cs(c("ukava", 5_000_000))),
		stakingGenesisState(),
		harvestGenesisState(c("hard", suite.rewardRate)),
	)

	suite.keeper = suite.app.GetHarvestKeeper()
	suite.stakingKeeper = suite.app.GetStakingKeeper()
}

func (suite *DelegatorRewardsTestSuite) TestSlash() {

	blockTime := time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC)
	ctx := suite.app.NewContext(true, abci.Header{Height: 1, Time: blockTime})
	const rewardDuration = 5
	suite.keeper.SetPreviousDelegationDistribution(ctx, blockTime.Add(-1*rewardDuration*time.Second), "ukava")

	suite.Require().NoError(
		suite.deliverMsgCreateValidator(ctx, suite.validatorAddrs[0], c("ukava", 5_000_000)),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	suite.Require().NoError(
		suite.slashValidator(ctx, suite.validatorAddrs[0], "0.05"),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	// Run function under test
	suite.keeper.ApplyDelegationRewards(ctx, "ukava")

	// Check claim amounts
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, sdk.AccAddress(suite.validatorAddrs[0]), c("hard", suite.rewardRate*rewardDuration)),
	)
}

func (suite *DelegatorRewardsTestSuite) TestUndelegation() {

	blockTime := time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC)
	ctx := suite.app.NewContext(true, abci.Header{Height: 1, Time: blockTime})
	const rewardDuration = 5
	suite.keeper.SetPreviousDelegationDistribution(ctx, blockTime.Add(-1*rewardDuration*time.Second), "ukava")

	suite.Require().NoError(
		suite.deliverMsgCreateValidator(ctx, suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgDelegate(ctx, suite.delegatorAddrs[0], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	suite.Require().NoError(
		suite.deliverMsgUndelegate(ctx, suite.delegatorAddrs[0], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	// Run function under test
	suite.keeper.ApplyDelegationRewards(ctx, "ukava")

	// Check claim amounts
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, sdk.AccAddress(suite.validatorAddrs[0]), c("hard", suite.rewardRate*rewardDuration)),
	)
	suite.Require().False(
		suite.kavaClaimExists(ctx, suite.delegatorAddrs[0]),
	)
}

func (suite *DelegatorRewardsTestSuite) TestUnevenNumberDelegations() {

	// Setup a context
	blockTime := time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC)
	ctx := suite.app.NewContext(true, abci.Header{Height: 1, Time: blockTime})
	const rewardDuration = 5
	suite.keeper.SetPreviousDelegationDistribution(ctx, blockTime.Add(-1*rewardDuration*time.Second), "ukava")

	suite.Require().NoError(
		suite.deliverMsgCreateValidator(ctx, suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgDelegate(ctx, suite.delegatorAddrs[0], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgDelegate(ctx, suite.delegatorAddrs[1], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	// Run function under test
	suite.keeper.ApplyDelegationRewards(ctx, "ukava")

	// Check claim amounts
	expectedReward := suite.rewardRate * rewardDuration / 3 // floor division
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, sdk.AccAddress(suite.validatorAddrs[0]), c("hard", expectedReward)),
	)
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, suite.delegatorAddrs[0], c("hard", expectedReward)),
	)
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, suite.delegatorAddrs[1], c("hard", expectedReward)),
	)
}

func (suite *DelegatorRewardsTestSuite) TestSlashWithUndelegated() {

	// Setup a context
	blockTime := time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC)
	ctx := suite.app.NewContext(true, abci.Header{Height: 1, Time: blockTime})
	const rewardDuration = 5
	suite.keeper.SetPreviousDelegationDistribution(ctx, blockTime.Add(-1*rewardDuration*time.Second), "ukava")

	suite.Require().NoError(
		suite.deliverMsgCreateValidator(ctx, suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgDelegate(ctx, suite.delegatorAddrs[0], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgDelegate(ctx, suite.delegatorAddrs[1], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	suite.Require().NoError(
		suite.deliverMsgUndelegate(ctx, suite.delegatorAddrs[0], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	suite.Require().NoError(
		suite.slashValidator(ctx, suite.validatorAddrs[0], "0.05"),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	// Run function under test
	suite.keeper.ApplyDelegationRewards(ctx, "ukava")

	// Check claim amounts
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, sdk.AccAddress(suite.validatorAddrs[0]), c("hard", suite.rewardRate*rewardDuration/2)),
	)
	suite.Require().False(
		suite.kavaClaimExists(ctx, suite.delegatorAddrs[0]),
	)
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, suite.delegatorAddrs[1], c("hard", suite.rewardRate*rewardDuration/2)),
	)
}
func (suite *DelegatorRewardsTestSuite) TestUnbondingValidator() {

	// Setup a context
	blockTime := time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC)
	ctx := suite.app.NewContext(true, abci.Header{Height: 1, Time: blockTime})
	const rewardDuration = 5
	suite.keeper.SetPreviousDelegationDistribution(ctx, blockTime.Add(-1*rewardDuration*time.Second), "ukava")

	suite.Require().NoError(
		suite.deliverMsgCreateValidator(ctx, suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgCreateValidator(ctx, suite.validatorAddrs[1], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgDelegate(ctx, suite.delegatorAddrs[0], suite.validatorAddrs[0], c("ukava", 1_000_000)),
	)
	suite.Require().NoError(
		suite.deliverMsgDelegate(ctx, suite.delegatorAddrs[1], suite.validatorAddrs[1], c("ukava", 1_000_000)),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	suite.Require().NoError(
		// jail the validator to put it into an unbonding state
		suite.jailValidator(ctx, suite.validatorAddrs[0]),
	)
	staking.EndBlocker(ctx, suite.stakingKeeper)

	// Run function under test
	suite.keeper.ApplyDelegationRewards(ctx, "ukava")

	// Check claim amounts
	suite.Require().False(
		// validator 0 will be unbonding and should not receive rewards
		suite.kavaClaimExists(ctx, sdk.AccAddress(suite.validatorAddrs[0])),
	)
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, sdk.AccAddress(suite.validatorAddrs[1]), c("hard", suite.rewardRate*rewardDuration/2)),
	)
	suite.Require().False(
		// delegations to unbonding validators and should not receive rewards
		suite.kavaClaimExists(ctx, suite.delegatorAddrs[0]),
	)
	suite.Require().NoError(
		suite.verifyKavaClaimAmount(ctx, suite.delegatorAddrs[1], c("hard", suite.rewardRate*rewardDuration/2)),
	)
}

func (suite *DelegatorRewardsTestSuite) deliverMsgCreateValidator(ctx sdk.Context, address sdk.ValAddress, selfDelegation sdk.Coin) error {
	msg := staking.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		staking.Description{},
		staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1_000_000),
	)
	handleStakingMsg := staking.NewHandler(suite.stakingKeeper)
	_, err := handleStakingMsg(ctx, msg)
	return err
}
func (suite *DelegatorRewardsTestSuite) deliverMsgDelegate(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) error {
	msg := staking.NewMsgDelegate(
		delegator,
		validator,
		amount,
	)
	handleStakingMsg := staking.NewHandler(suite.stakingKeeper)
	_, err := handleStakingMsg(ctx, msg)
	return err
}
func (suite *DelegatorRewardsTestSuite) deliverMsgUndelegate(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) error {
	msg := staking.NewMsgUndelegate(
		delegator,
		validator,
		amount,
	)
	handleStakingMsg := staking.NewHandler(suite.stakingKeeper)
	_, err := handleStakingMsg(ctx, msg)
	return err
}

func (suite *DelegatorRewardsTestSuite) slashValidator(ctx sdk.Context, validator sdk.ValAddress, slashPercent string) error {
	// Assume slashable offence occurred at block 1. Note this might cause problems if tests are running at a block height higher than the unbonding period (default 3 weeks)
	const infractionHeight int64 = 1

	val, found := suite.stakingKeeper.GetValidator(ctx, validator)
	if !found {
		return fmt.Errorf("can't find validator in state")
	}
	suite.stakingKeeper.Slash(
		ctx,
		sdk.GetConsAddress(val.ConsPubKey),
		infractionHeight,
		val.GetConsensusPower(),
		sdk.MustNewDecFromStr(slashPercent),
	)
	return nil
}
func (suite *DelegatorRewardsTestSuite) jailValidator(ctx sdk.Context, validator sdk.ValAddress) error {
	val, found := suite.stakingKeeper.GetValidator(ctx, validator)
	if !found {
		return fmt.Errorf("can't find validator in state")
	}
	suite.stakingKeeper.Jail(ctx, sdk.GetConsAddress(val.ConsPubKey))
	return nil
}

// verifyKavaClaimAmount looks up a ukava claim and checks the claim amount is equal to an expected value
func (suite *DelegatorRewardsTestSuite) verifyKavaClaimAmount(ctx sdk.Context, owner sdk.AccAddress, expectedAmount sdk.Coin) error {
	claim, found := suite.keeper.GetClaim(ctx, owner, "ukava", types.Stake)
	if !found {
		return fmt.Errorf("could not find claim")
	}
	if !expectedAmount.IsEqual(claim.Amount) {
		return fmt.Errorf("expected claim amount (%s) != actual claim amount (%s)", expectedAmount, claim.Amount)
	}
	return nil
}

// kavaClaimExists checks the store for a ukava claim
func (suite *DelegatorRewardsTestSuite) kavaClaimExists(ctx sdk.Context, owner sdk.AccAddress) bool {
	_, found := suite.keeper.GetClaim(ctx, owner, "ukava", types.Stake)
	return found
}

func harvestGenesisState(rewardRate sdk.Coin) app.GenesisState {
	loanToValue := sdk.MustNewDecFromStr("0.6")
	genState := types.NewGenesisState(
		types.NewParams(
			true,
			types.DistributionSchedules{
				types.NewDistributionSchedule(
					true,
					"bnb",
					time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC),
					time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC),
					rewardRate,
					time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC),
					types.Multipliers{
						types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")),
						types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")),
						types.NewMultiplier(types.Large, 24, sdk.OneDec()),
					},
				),
			},
			types.DelegatorDistributionSchedules{
				types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(
						true,
						"ukava",
						time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC),
						time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC),
						rewardRate,
						time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC),
						types.Multipliers{
							types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")),
							types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")),
							types.NewMultiplier(types.Large, 24, sdk.OneDec()),
						},
					),
					time.Hour*24,
				),
			},
			types.MoneyMarkets{
				types.NewMoneyMarket("usdx", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "usdx:usd", sdk.NewInt(1000000), sdk.NewInt(USDX_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "kava:usd", sdk.NewInt(1000000), sdk.NewInt(KAVA_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			},
			0, // LTV counter
		),
		types.DefaultPreviousBlockTime,
		types.DefaultDistributionTimes,
	)
	return app.GenesisState{
		types.ModuleName: types.ModuleCdc.MustMarshalJSON(genState),
	}

}

func stakingGenesisState() app.GenesisState {
	genState := staking.DefaultGenesisState()
	genState.Params.BondDenom = "ukava"
	return app.GenesisState{
		staking.ModuleName: staking.ModuleCdc.MustMarshalJSON(genState),
	}
}

// equalCoinsAuthGenState returns an auth genesis state with the same coins for each account
func equalCoinsAuthGenState(addresses []sdk.AccAddress, coins sdk.Coins) app.GenesisState {
	coinsList := []sdk.Coins{}
	for range addresses {
		coinsList = append(coinsList, coins)
	}
	return app.NewAuthGenState(addresses, coinsList)
}
