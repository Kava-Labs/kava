package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Test suite used for all keeper tests
type DelegatorRewardsTestSuite struct {
	suite.Suite

	keeper        keeper.Keeper
	stakingKeeper stakingkeeper.Keeper

	app app.TestApp
	ctx sdk.Context

	genesisTime    time.Time
	addrs          []sdk.AccAddress
	validatorAddrs []sdk.ValAddress
}

// SetupTest is run automatically before each suite test
func (suite *DelegatorRewardsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, allAddrs := app.GeneratePrivKeyAddressPairs(10)
	suite.addrs = allAddrs[:5]
	for _, a := range allAddrs[5:] {
		suite.validatorAddrs = append(suite.validatorAddrs, sdk.ValAddress(a))
	}
	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *DelegatorRewardsTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.stakingKeeper = suite.app.GetStakingKeeper()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: suite.genesisTime})
}

func (suite *DelegatorRewardsTestSuite) SetupWithGenState(authBuilder app.AuthGenesisBuilder, incentBuilder IncentiveGenesisBuilder) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authBuilder.BuildMarshalled(),
		NewStakingGenesisState(),
		incentBuilder.BuildMarshalled(),
	)
}

func (suite *DelegatorRewardsTestSuite) TestAccumulateHardDelegatorRewards() {
	type args struct {
		delegation           sdk.Coin
		rewardsPerSecond     sdk.Coin
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
				delegation:           c("ukava", 1_000_000),
				rewardsPerSecond:     c("hard", 122354),
				timeElapsed:          7,
				expectedRewardFactor: d("0.428239000000000000"),
			},
		},
		{
			"1 day",
			args{
				delegation:           c("ukava", 1_000_000),
				rewardsPerSecond:     c("hard", 122354),
				timeElapsed:          86400,
				expectedRewardFactor: d("5285.692800000000000000"),
			},
		},
		{
			"0 seconds",
			args{
				delegation:           c("ukava", 1_000_000),
				rewardsPerSecond:     c("hard", 122354),
				timeElapsed:          0,
				expectedRewardFactor: d("0.0"),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			authBuilder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
				WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[0]), cs(c("ukava", 1e9)))

			incentBuilder := NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleDelegatorRewardPeriod(tc.args.delegation.Denom, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder)

			err := suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[0], tc.args.delegation)
			suite.Require().NoError(err)
			err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[0], tc.args.delegation)
			suite.Require().NoError(err)

			staking.EndBlocker(suite.ctx, suite.stakingKeeper)

			// Set up chain context at future time
			runAtTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			rewardPeriod, found := suite.keeper.GetHardDelegatorRewardPeriod(runCtx, tc.args.delegation.Denom)
			suite.Require().True(found)
			err = suite.keeper.AccumulateHardDelegatorRewards(runCtx, rewardPeriod)
			suite.Require().NoError(err)

			rewardFactor, _ := suite.keeper.GetHardDelegatorRewardFactor(runCtx, tc.args.delegation.Denom)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *DelegatorRewardsTestSuite) TestSynchronizeHardDelegatorReward() {
	type args struct {
		delegation           sdk.Coin
		rewardsPerSecond     sdk.Coin
		blockTimes           []int
		expectedRewardFactor sdk.Dec
		expectedRewards      sdk.Coins
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"10 blocks",
			args{
				delegation:           c("ukava", 1_000_000),
				rewardsPerSecond:     c("hard", 122354),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("6.117700000000000000"),
				expectedRewards:      cs(c("hard", 6117700)),
			},
		},
		{
			"10 blocks - long block time",
			args{
				delegation:           c("ukava", 1_000_000),
				rewardsPerSecond:     c("hard", 122354),
				blockTimes:           []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardFactor: d("52856.928000000000000000"),
				expectedRewards:      cs(c("hard", 52856928000)),
			},
		},
		{
			"delegator reward index updated when reward is zero",
			args{
				delegation:           c("ukava", 1),
				rewardsPerSecond:     c("hard", 1),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("0.000099999900000100"),
				expectedRewards:      nil,
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			authBuilder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
				WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[0]), cs(c("ukava", 1e9)))

			incentBuilder := NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleDelegatorRewardPeriod(tc.args.delegation.Denom, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder)

			// Create validator account
			staking.BeginBlocker(suite.ctx, suite.stakingKeeper)
			selfDelegationCoins := c("ukava", 1_000_000)
			err := suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[0], selfDelegationCoins)
			suite.Require().NoError(err)
			staking.EndBlocker(suite.ctx, suite.stakingKeeper)

			// Delegator delegates
			err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[0], tc.args.delegation)
			suite.Require().NoError(err)

			// Check that validator account has been created and delegation was successful
			valAcc, found := suite.stakingKeeper.GetValidator(suite.ctx, suite.validatorAddrs[0])
			suite.True(found)
			suite.Require().Equal(valAcc.Status, sdk.Bonded)
			suite.Require().Equal(valAcc.Tokens, tc.args.delegation.Amount.Add(selfDelegationCoins.Amount))

			// Check that Staking hooks initialized a HardLiquidityProviderClaim
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.DelegatorRewardIndexes[0].RewardFactor)

			// Run accumulator at several intervals
			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)

				rewardPeriod, found := suite.keeper.GetHardDelegatorRewardPeriod(blockCtx, tc.args.delegation.Denom)
				suite.Require().True(found)

				err := suite.keeper.AccumulateHardDelegatorRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// After we've accumulated, run synchronize
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeHardDelegatorRewards(suite.ctx, suite.addrs[0], nil, false)
			})

			// Check that reward factor and claim have been updated as expected
			rewardFactor, _ := suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, tc.args.delegation.Denom)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)

			claim, found = suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedRewardFactor, claim.DelegatorRewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, claim.Reward)
		})
	}
}

func (suite *DelegatorRewardsTestSuite) TestSimulateHardDelegatorRewardSynchronization() {
	type args struct {
		delegation            sdk.Coin
		rewardsPerSecond      sdk.Coin
		blockTimes            []int
		expectedRewardIndexes types.RewardIndexes
		expectedRewards       sdk.Coins
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"10 blocks",
			args{
				delegation:            c("ukava", 1_000_000),
				rewardsPerSecond:      c("hard", 122354),
				blockTimes:            []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("ukava", d("6.117700000000000000"))}, // Here the reward index stores data differently than inside a MultiRewardIndex
				expectedRewards:       cs(c("hard", 6117700)),
			},
		},
		{
			"10 blocks - long block time",
			args{
				delegation:            c("ukava", 1_000_000),
				rewardsPerSecond:      c("hard", 122354),
				blockTimes:            []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("ukava", d("52856.928000000000000000"))},
				expectedRewards:       cs(c("hard", 52856928000)),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			authBuilder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
				WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[0]), cs(c("ukava", 1e9)))

			incentBuilder := NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleDelegatorRewardPeriod(tc.args.delegation.Denom, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder)

			// Delegator delegates
			err := suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[0], tc.args.delegation)
			suite.Require().NoError(err)
			err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[0], tc.args.delegation)
			suite.Require().NoError(err)

			staking.EndBlocker(suite.ctx, suite.stakingKeeper)

			// Check that Staking hooks initialized a HardLiquidityProviderClaim
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.DelegatorRewardIndexes[0].RewardFactor)

			// Run accumulator at several intervals
			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)

				// Accumulate hard delegator rewards
				rewardPeriod, found := suite.keeper.GetHardDelegatorRewardPeriod(blockCtx, tc.args.delegation.Denom)
				suite.Require().True(found)
				err := suite.keeper.AccumulateHardDelegatorRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// Check that the synced claim held in memory has properly simulated syncing
			syncedClaim := suite.keeper.SimulateHardSynchronization(suite.ctx, claim)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				// Check that the user's claim's reward index matches the expected reward index
				rewardIndex, found := syncedClaim.DelegatorRewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(expectedRewardIndex, rewardIndex)

				// Check that the user's claim holds the expected amount of reward coins
				suite.Require().Equal(
					tc.args.expectedRewards.AmountOf(expectedRewardIndex.CollateralType),
					syncedClaim.Reward.AmountOf(expectedRewardIndex.CollateralType),
				)
			}
		})
	}
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

func (suite *DelegatorRewardsTestSuite) deliverMsgRedelegate(ctx sdk.Context, delegator sdk.AccAddress, sourceValidator, destinationValidator sdk.ValAddress, amount sdk.Coin) error {
	msg := staking.NewMsgBeginRedelegate(
		delegator,
		sourceValidator,
		destinationValidator,
		amount,
	)
	handleStakingMsg := staking.NewHandler(suite.stakingKeeper)
	_, err := handleStakingMsg(ctx, msg)
	return err
}

// given a user has a delegation to a bonded validator, when the validator starts unbonding, the user does not accumulate rewards
func (suite *DelegatorRewardsTestSuite) TestUnbondingValidatorSyncsClaim() {
	authBuilder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
		WithSimpleAccount(suite.addrs[2], cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[0]), cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[1]), cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[2]), cs(c("ukava", 1e9)))

	rewardsPerSecond := c("hard", 122354)
	bondDenom := "ukava"

	incentBuilder := NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithSimpleDelegatorRewardPeriod(bondDenom, rewardsPerSecond)

	suite.SetupWithGenState(authBuilder, incentBuilder)

	blockDuration := 10 * time.Second

	// Reduce the size of the validator set
	stakingParams := suite.app.GetStakingKeeper().GetParams(suite.ctx)
	stakingParams.MaxValidators = 2
	suite.app.GetStakingKeeper().SetParams(suite.ctx, stakingParams)

	// Create 3 validators
	err := suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[0], c(bondDenom, 10_000_000))
	suite.Require().NoError(err)
	err = suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[1], c(bondDenom, 5_000_000))
	suite.Require().NoError(err)
	err = suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[2], c(bondDenom, 1_000_000))
	suite.Require().NoError(err)

	// End the block so top validators become bonded
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})

	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(1 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	// Delegate to a bonded validator from the test user. This will initialize their incentive claim.
	err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[1], c(bondDenom, 1_000_000))
	suite.Require().NoError(err)

	// Start a new block to accumulate some delegation rewards for the user.
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(2 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	// Delegate to the unbonded validator to push it into the bonded validator set, pushing out the user's delegated validator
	err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[2], suite.validatorAddrs[2], c(bondDenom, 8_000_000))
	suite.Require().NoError(err)

	// End the block to start unbonding the user's validator
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	// but don't start the next block as it will accumulate delegator rewards and we won't be able to tell if the user's reward was synced.

	// Check that the user's claim has been synced. ie rewards added, index updated
	claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	globalIndex, found := suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, bondDenom)
	suite.Require().True(found)
	claimIndex, found := claim.DelegatorRewardIndexes.GetRewardIndex(bondDenom)
	suite.Require().True(found)
	suite.Require().Equal(globalIndex, claimIndex.RewardFactor)

	suite.Require().Equal(
		cs(c(rewardsPerSecond.Denom, 76471)),
		claim.Reward,
	)

	// Run another block and check the claim is not accumulating more rewards
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(3 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})

	suite.keeper.SynchronizeHardDelegatorRewards(suite.ctx, suite.addrs[0], nil, false)

	// rewards are the same as before
	laterClaim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().Equal(claim.Reward, laterClaim.Reward)

	// claim index has been updated to latest global value
	laterClaimIndex, found := laterClaim.DelegatorRewardIndexes.GetRewardIndex(bondDenom)
	suite.Require().True(found)
	globalIndex, found = suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, bondDenom)
	suite.Require().True(found)
	suite.Require().Equal(globalIndex, laterClaimIndex.RewardFactor)
}

// given a user has a delegation to an unbonded validator, when the validator becomes bonded, the user starts accumulating rewards
func (suite *DelegatorRewardsTestSuite) TestBondingValidatorSyncsClaim() {
	authBuilder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
		WithSimpleAccount(suite.addrs[2], cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[0]), cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[1]), cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[2]), cs(c("ukava", 1e9)))

	rewardsPerSecond := c("hard", 122354)
	bondDenom := "ukava"

	incentBuilder := NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithSimpleDelegatorRewardPeriod(bondDenom, rewardsPerSecond)

	suite.SetupWithGenState(authBuilder, incentBuilder)

	blockDuration := 10 * time.Second

	// Reduce the size of the validator set
	stakingParams := suite.app.GetStakingKeeper().GetParams(suite.ctx)
	stakingParams.MaxValidators = 2
	suite.app.GetStakingKeeper().SetParams(suite.ctx, stakingParams)

	// Create 3 validators
	err := suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[0], c(bondDenom, 10_000_000))
	suite.Require().NoError(err)
	err = suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[1], c(bondDenom, 5_000_000))
	suite.Require().NoError(err)
	err = suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[2], c(bondDenom, 1_000_000))
	suite.Require().NoError(err)

	// End the block so top validators become bonded
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})

	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(1 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	// Delegate to an unbonded validator from the test user. This will initialize their incentive claim.
	err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[2], c(bondDenom, 1_000_000))
	suite.Require().NoError(err)

	// Start a new block to accumulate some delegation rewards globally.
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(2 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})

	// Delegate to the user's unbonded validator to push it into the bonded validator set
	err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[2], suite.validatorAddrs[2], c(bondDenom, 4_000_000))
	suite.Require().NoError(err)

	// End the block to bond the user's validator
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	// but don't start the next block as it will accumulate delegator rewards and we won't be able to tell if the user's reward was synced.

	// Check that the user's claim has been synced. ie rewards added, index updated
	claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	globalIndex, found := suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, bondDenom)
	suite.Require().True(found)
	claimIndex, found := claim.DelegatorRewardIndexes.GetRewardIndex(bondDenom)
	suite.Require().True(found)
	suite.Require().Equal(globalIndex, claimIndex.RewardFactor)

	suite.Require().Equal(
		sdk.Coins(nil),
		claim.Reward,
	)

	// Run another block and check the claim is accumulating more rewards
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(3 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})

	suite.keeper.SynchronizeHardDelegatorRewards(suite.ctx, suite.addrs[0], nil, false)

	// rewards are greater than before
	laterClaim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().True(laterClaim.Reward.IsAllGT(claim.Reward))

	// claim index has been updated to latest global value
	laterClaimIndex, found := laterClaim.DelegatorRewardIndexes.GetRewardIndex(bondDenom)
	suite.Require().True(found)
	globalIndex, found = suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, bondDenom)
	suite.Require().True(found)
	suite.Require().Equal(globalIndex, laterClaimIndex.RewardFactor)
}

// If a validator is slashed delegators should have their claims synced
func (suite *DelegatorRewardsTestSuite) TestSlashingValidatorSyncsClaim() {
	authBuilder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[0]), cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[1]), cs(c("ukava", 1e9)))

	rewardsPerSecond := c("hard", 122354)
	bondDenom := "ukava"

	incentBuilder := NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithSimpleDelegatorRewardPeriod(bondDenom, rewardsPerSecond)

	suite.SetupWithGenState(authBuilder, incentBuilder)

	blockDuration := 10 * time.Second

	// Reduce the size of the validator set
	stakingParams := suite.app.GetStakingKeeper().GetParams(suite.ctx)
	stakingParams.MaxValidators = 2
	suite.app.GetStakingKeeper().SetParams(suite.ctx, stakingParams)

	// Create 2 validators
	err := suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[0], c(bondDenom, 10_000_000))
	suite.Require().NoError(err)
	err = suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[1], c(bondDenom, 10_000_000))
	suite.Require().NoError(err)

	// End the block so validators become bonded
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})

	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(1 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	// Delegate to a bonded validator from the test user. This will initialize their incentive claim.
	err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[1], c(bondDenom, 1_000_000))
	suite.Require().NoError(err)

	// Check that claim has been created with synced reward index but no reward coins
	initialClaim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
	suite.True(found)
	initialGlobalIndex, found := suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, bondDenom)
	suite.True(found)
	initialClaimIndex, found := initialClaim.DelegatorRewardIndexes.GetRewardIndex(bondDenom)
	suite.True(found)
	suite.Require().Equal(initialGlobalIndex, initialClaimIndex.RewardFactor)
	suite.True(initialClaim.Reward.Empty()) // Initial claim should not have any rewards

	// Start a new block to accumulate some delegation rewards for the user.
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(2 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	// Fetch validator and slash them
	stakingKeeper := suite.app.GetStakingKeeper()
	validator, found := stakingKeeper.GetValidator(suite.ctx, suite.validatorAddrs[1])
	suite.Require().True(found)
	suite.Require().True(validator.GetTokens().IsPositive())
	fraction := sdk.NewDecWithPrec(5, 1)
	stakingKeeper.Slash(suite.ctx, validator.ConsAddress(), suite.ctx.BlockHeight(), 10, fraction)

	// Check that the user's claim has been synced. ie rewards added, index updated
	claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)
	globalIndex, found := suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, bondDenom)
	suite.Require().True(found)
	claimIndex, found := claim.DelegatorRewardIndexes.GetRewardIndex(bondDenom)
	suite.Require().True(found)
	suite.Require().Equal(globalIndex, claimIndex.RewardFactor)

	// Check that rewards were added
	suite.Require().Equal(
		cs(c(rewardsPerSecond.Denom, 58264)),
		claim.Reward,
	)

	// Check that reward factor increased from initial value
	suite.True(claimIndex.RewardFactor.GT(initialClaimIndex.RewardFactor))
}

// Given a delegation to a bonded validator, when a user redelegates everything to another (bonded) validator, the user's claim is synced
func (suite *DelegatorRewardsTestSuite) TestRedelegationSyncsClaim() {
	authBuilder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[0]), cs(c("ukava", 1e9))).
		WithSimpleAccount(sdk.AccAddress(suite.validatorAddrs[1]), cs(c("ukava", 1e9)))

	rewardsPerSecond := c("hard", 122354)
	bondDenom := "ukava"

	incentBuilder := NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithSimpleDelegatorRewardPeriod(bondDenom, rewardsPerSecond)

	suite.SetupWithGenState(authBuilder, incentBuilder)

	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime)
	blockDuration := 10 * time.Second

	// Create 2 validators
	err := suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[0], c(bondDenom, 10_000_000))
	suite.Require().NoError(err)
	err = suite.deliverMsgCreateValidator(suite.ctx, suite.validatorAddrs[1], c(bondDenom, 5_000_000))
	suite.Require().NoError(err)

	// Delegatefrom the test user. This will initialize their incentive claim.
	err = suite.deliverMsgDelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[0], c(bondDenom, 1_000_000))
	suite.Require().NoError(err)

	// Start a new block to accumulate some delegation rewards globally.
	_ = suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(1 * blockDuration))
	_ = suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	// Redelegate the user's delegation between the two validators. This should trigger hooks that sync the user's claim.
	err = suite.deliverMsgRedelegate(suite.ctx, suite.addrs[0], suite.validatorAddrs[0], suite.validatorAddrs[1], c(bondDenom, 1_000_000))
	suite.Require().NoError(err)

	// Check that the user's claim has been synced. ie rewards added, index updated
	claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	globalIndex, found := suite.keeper.GetHardDelegatorRewardFactor(suite.ctx, bondDenom)
	suite.Require().True(found)
	claimIndex, found := claim.DelegatorRewardIndexes.GetRewardIndex(bondDenom)
	suite.Require().True(found)
	suite.Require().Equal(globalIndex, claimIndex.RewardFactor)
	suite.Require().Equal(
		cs(c(rewardsPerSecond.Denom, 76471)),
		claim.Reward,
	)
}

func TestDelegatorRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(DelegatorRewardsTestSuite))
}
