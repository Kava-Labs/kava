package accumulators_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/keeper/accumulators"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
)

type AccumulateEarnRewardsIntegrationTests struct {
	testutil.IntegrationTester

	keeper    testutil.TestKeeper
	userAddrs []sdk.AccAddress
	valAddrs  []sdk.ValAddress
}

func TestAccumulateEarnRewardsIntegrationTests(t *testing.T) {
	suite.Run(t, new(AccumulateEarnRewardsIntegrationTests))
}

func (suite *AccumulateEarnRewardsIntegrationTests) SetupTest() {
	suite.IntegrationTester.SetupTest()

	suite.keeper = testutil.TestKeeper{
		Keeper: suite.App.GetIncentiveKeeper(),
	}

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	suite.userAddrs = addrs[0:2]
	suite.valAddrs = []sdk.ValAddress{
		sdk.ValAddress(addrs[2]),
		sdk.ValAddress(addrs[3]),
	}

	// Setup app with test state
	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleAccount(addrs[0], cs(c("ukava", 1e12))).
		WithSimpleAccount(addrs[1], cs(c("ukava", 1e12))).
		WithSimpleAccount(addrs[2], cs(c("ukava", 1e12))).
		WithSimpleAccount(addrs[3], cs(c("ukava", 1e12)))

	incentiveBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.GenesisTime).
		WithSimpleRewardPeriod(types.CLAIM_TYPE_EARN, "bkava", cs())

	savingsBuilder := testutil.NewSavingsGenesisBuilder().
		WithSupportedDenoms("bkava")

	earnBuilder := testutil.NewEarnGenesisBuilder().
		WithAllowedVaults(earntypes.AllowedVault{
			Denom:             "bkava",
			Strategies:        earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
			IsPrivateVault:    false,
			AllowedDepositors: nil,
		})

	stakingBuilder := testutil.NewStakingGenesisBuilder()

	mintBuilder := testutil.NewMintGenesisBuilder().
		WithInflationMax(sdk.OneDec()).
		WithInflationMin(sdk.OneDec()).
		WithMinter(sdk.OneDec(), sdk.ZeroDec()).
		WithMintDenom("ukava")

	suite.StartChainWithBuilders(
		authBuilder,
		incentiveBuilder,
		savingsBuilder,
		earnBuilder,
		stakingBuilder,
		mintBuilder,
	)
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateUpdatedWhenBlockTimeHasIncreased() {
	suite.AddIncentiveMultiRewardPeriod(
		types.CLAIM_TYPE_EARN,
		types.NewMultiRewardPeriod(
			true,
			"bkava",         // reward period is set for "bkava" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
		),
	)

	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ukava", 800000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ukava", 200000))
	suite.NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], derivative0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], derivative1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}

	suite.keeper.StoreGlobalIndexes(suite.Ctx, types.CLAIM_TYPE_EARN, globalIndexes)
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	// Mint tokens, distribute to validators, claim staking rewards
	// 1 hour later
	_, resBeginBlock := suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	validatorRewards, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	suite.Require().Contains(validatorRewards, suite.valAddrs[1].String(), "there should be claim events for validator 0")
	suite.Require().Contains(validatorRewards, suite.valAddrs[0].String(), "there should be claim events for validator 1")

	// check time and factors

	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	stakingRewardIndexes0 := validatorRewards[suite.valAddrs[0].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative0.Amount.ToDec())

	stakingRewardIndexes1 := validatorRewards[suite.valAddrs[1].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative1.Amount.ToDec())

	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("3.64").Add(stakingRewardIndexes0),
		},
	})
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("3.64").Add(stakingRewardIndexes1),
		},
	})
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateUpdatedWhenBlockTimeHasIncreased_partialDeposit() {
	suite.AddIncentiveMultiRewardPeriod(
		types.CLAIM_TYPE_EARN,
		types.NewMultiRewardPeriod(
			true,
			"bkava",         // reward period is set for "bkava" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
		),
	)

	// 800000bkava0 minted, 700000 deposited
	// 200000bkava1 minted, 100000 deposited
	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ukava", 800000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ukava", 200000))
	suite.NoError(err)

	depositAmount0 := c(derivative0.Denom, 700000)
	depositAmount1 := c(derivative1.Denom, 100000)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], depositAmount0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], depositAmount1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}

	suite.keeper.StoreGlobalIndexes(suite.Ctx, types.CLAIM_TYPE_EARN, globalIndexes)

	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	// Mint tokens, distribute to validators, claim staking rewards
	// 1 hour later
	_, resBeginBlock := suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	validatorRewards, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	suite.Require().Contains(validatorRewards, suite.valAddrs[1].String(), "there should be claim events for validator 0")
	suite.Require().Contains(validatorRewards, suite.valAddrs[0].String(), "there should be claim events for validator 1")

	// check time and factors

	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	// Divided by deposit amounts, not bank supply amounts
	stakingRewardIndexes0 := validatorRewards[suite.valAddrs[0].String()].
		AmountOf("ukava").
		ToDec().
		Quo(depositAmount0.Amount.ToDec())

	stakingRewardIndexes1 := validatorRewards[suite.valAddrs[1].String()].
		AmountOf("ukava").
		ToDec().
		Quo(depositAmount1.Amount.ToDec())

	// Slightly increased rewards due to less bkava deposited
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("8.248571428571428571"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("4.154285714285714285").Add(stakingRewardIndexes0),
		},
	})

	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("14.42"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("7.24").Add(stakingRewardIndexes1),
		},
	})
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ukava", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ukava", 1000000))
	suite.NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], derivative0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], derivative1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.keeper.StoreGlobalIndexes(suite.Ctx, types.CLAIM_TYPE_EARN, previousIndexes)

	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	period := types.NewMultiRewardPeriod(
		true,
		"bkava",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	// Must manually accumulate rewards as BeginBlockers only run when the block time increases
	// This does not run any x/mint or x/distribution BeginBlockers
	earnKeeper := suite.App.GetEarnKeeper()
	err = accumulators.
		NewEarnAccumulator(suite.keeper.Store, suite.App.GetLiquidKeeper(), &earnKeeper, suite.keeper.Adapters).
		AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_EARN, period)
	suite.NoError(err)

	// check time and factors

	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	expected, f := previousIndexes.Get(derivative0.Denom)
	suite.True(f)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative0.Denom, expected)

	expected, f = previousIndexes.Get(derivative1.Denom)
	suite.True(f)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative1.Denom, expected)
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestNoAccumulationWhenSourceSharesAreZero() {
	suite.AddIncentiveMultiRewardPeriod(
		types.CLAIM_TYPE_EARN,
		types.NewMultiRewardPeriod(
			true,
			"bkava",         // reward period is set for "bkava" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
		),
	)

	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ukava", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ukava", 1000000))
	suite.NoError(err)

	// No earn deposits

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.keeper.StoreGlobalIndexes(suite.Ctx, types.CLAIM_TYPE_EARN, previousIndexes)

	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	// Mint tokens, distribute to validators, claim staking rewards
	// 1 hour later
	_, _ = suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)
	// check time and factors

	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	expected, f := previousIndexes.Get(derivative0.Denom)
	suite.True(f)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative0.Denom, expected)

	expected, f = previousIndexes.Get(derivative1.Denom)
	suite.True(f)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative1.Denom, expected)
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateAddedWhenStateDoesNotExist() {
	suite.AddIncentiveMultiRewardPeriod(
		types.CLAIM_TYPE_EARN,
		types.NewMultiRewardPeriod(
			true,
			"bkava",         // reward period is set for "bkava" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
		),
	)

	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ukava", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ukava", 1000000))
	suite.NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], derivative0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], derivative1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	_, resBeginBlock := suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	// After the second accumulation both current block time and indexes should be stored.
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	validatorRewards0, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	firstStakingRewardIndexes0 := validatorRewards0[suite.valAddrs[0].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative0.Amount.ToDec())

	firstStakingRewardIndexes1 := validatorRewards0[suite.valAddrs[1].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative1.Amount.ToDec())

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	// First accumulation can have staking rewards, but no other rewards
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "ukava",
			RewardFactor:   firstStakingRewardIndexes0,
		},
	})
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "ukava",
			RewardFactor:   firstStakingRewardIndexes1,
		},
	})

	_, resBeginBlock = suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	// After the second accumulation both current block time and indexes should be stored.
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, suite.Ctx.BlockTime())

	validatorRewards1, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	secondStakingRewardIndexes0 := validatorRewards1[suite.valAddrs[0].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative0.Amount.ToDec())

	secondStakingRewardIndexes1 := validatorRewards1[suite.valAddrs[1].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative1.Amount.ToDec())

	// Second accumulation has both staking rewards and incentive rewards
	// ukava incentive rewards: 3600 * 1000 / (2 * 1000000) == 1.8
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "ukava",
			// Incentive rewards + both staking rewards
			RewardFactor: d("1.8").Add(firstStakingRewardIndexes0).Add(secondStakingRewardIndexes0),
		},
		{
			CollateralType: "earn",
			RewardFactor:   d("3.6"),
		},
	})
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "ukava",
			// Incentive rewards + both staking rewards
			RewardFactor: d("1.8").Add(firstStakingRewardIndexes1).Add(secondStakingRewardIndexes1),
		},
		{
			CollateralType: "earn",
			RewardFactor:   d("3.6"),
		},
	})
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestNoPanicWhenStateDoesNotExist() {
	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ukava", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ukava", 1000000))
	suite.NoError(err)

	period := types.NewMultiRewardPeriod(
		true,
		"bkava",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(),
	)

	// Accumulate with no earn shares and no rewards per second will result in no increment to the indexes.
	// No increment and no previous indexes stored, results in an updated of nil. Setting this in the state panics.
	// Check there is no panic.
	suite.NotPanics(func() {
		// This does not update any state, as there are no bkava vaults
		// to iterate over, denoms are unknown
		err := suite.keeper.AccumulateEarnRewards(suite.Ctx, period)
		suite.NoError(err)
	})

	// Times are not stored for vaults with no state
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative0.Denom, time.Time{})
	suite.StoredTimeEquals(types.CLAIM_TYPE_EARN, derivative1.Denom, time.Time{})
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative0.Denom, nil)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_EARN, derivative1.Denom, nil)
}
