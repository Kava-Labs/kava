package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
)

type AccumulateEarnRewardsIntegrationTests struct {
	testutil.IntegrationTester

	keeper    TestKeeper
	userAddrs []sdk.AccAddress
	valAddrs  []sdk.ValAddress
}

func TestAccumulateEarnRewardsIntegrationTests(t *testing.T) {
	suite.Run(t, new(AccumulateEarnRewardsIntegrationTests))
}

func (suite *AccumulateEarnRewardsIntegrationTests) SetupTest() {
	suite.IntegrationTester.SetupTest()

	suite.keeper = TestKeeper{
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
		WithSimpleEarnRewardPeriod("bkava", cs())

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
	suite.AddIncentiveEarnMultiRewardPeriod(
		types.NewMultiRewardPeriod(
			true,
			"bkava",         // reward period is set for "bkava" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
		),
	)

	derivative0, err := suite.MintLiquidAnyValAddr(
		suite.userAddrs[0],
		suite.valAddrs[0],
		c("ukava", 800000),
	)
	suite.NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], derivative0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ukava", 200000))
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

	suite.keeper.storeGlobalEarnIndexes(suite.Ctx, globalIndexes)
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative1.Denom, suite.Ctx.BlockTime())

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

	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	stakingRewardIndexes0 := validatorRewards[suite.valAddrs[0].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative0.Amount.ToDec())

	stakingRewardIndexes1 := validatorRewards[suite.valAddrs[1].String()].
		AmountOf("ukava").
		ToDec().
		Quo(derivative1.Amount.ToDec())

	suite.StoredEarnIndexesEqual(derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("3.64").Add(stakingRewardIndexes0),
		},
	})
	suite.StoredEarnIndexesEqual(derivative1.Denom, types.RewardIndexes{
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
