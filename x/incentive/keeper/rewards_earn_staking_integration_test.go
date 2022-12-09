package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

type EarnStakingRewardsIntegrationTestSuite struct {
	testutil.IntegrationTester

	keeper    TestKeeper
	userAddrs []sdk.AccAddress
	valAddrs  []sdk.ValAddress
}

func TestEarnStakingRewardsIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(EarnStakingRewardsIntegrationTestSuite))
}

func (suite *EarnStakingRewardsIntegrationTestSuite) SetupTest() {
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

	kavamintBuilder := testutil.NewKavamintGenesisBuilder()

	suite.StartChainWithBuilders(
		authBuilder,
		incentiveBuilder,
		savingsBuilder,
		earnBuilder,
		stakingBuilder,
		mintBuilder,
		kavamintBuilder,
	)
}

func (suite *EarnStakingRewardsIntegrationTestSuite) TestStakingRewardsDistributed() {
	// derivative 1: 8 total staked, 7 to earn, 1 not in earn
	// derivative 2: 2 total staked, 1 to earn, 1 not in earn
	userMintAmount0 := c("ukava", 8e9)
	userMintAmount1 := c("ukava", 2e9)

	userDepositAmount0 := i(7e9)
	userDepositAmount1 := i(1e9)

	// Create two validators
	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], userMintAmount0)
	suite.Require().NoError(err)

	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[1], userMintAmount1)
	suite.Require().NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], sdk.NewCoin(derivative0.Denom, userDepositAmount0), earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], sdk.NewCoin(derivative1.Denom, userDepositAmount1), earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	// Get derivative denoms
	lq := suite.App.GetLiquidKeeper()
	vaultDenom1 := lq.GetLiquidStakingTokenDenom(suite.valAddrs[0])
	vaultDenom2 := lq.GetLiquidStakingTokenDenom(suite.valAddrs[1])

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.Ctx = suite.Ctx.WithBlockTime(previousAccrualTime)

	initialVault1RewardFactor := d("0.04")
	initialVault2RewardFactor := d("0.04")

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "ukava",
					RewardFactor:   initialVault1RewardFactor,
				},
			},
		},
		{
			CollateralType: vaultDenom2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "ukava",
					RewardFactor:   initialVault2RewardFactor,
				},
			},
		},
	}

	suite.keeper.storeGlobalEarnIndexes(suite.Ctx, globalIndexes)

	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, vaultDenom1, suite.Ctx.BlockTime())
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, vaultDenom2, suite.Ctx.BlockTime())

	val := suite.GetAbciValidator(suite.valAddrs[0])

	// Mint tokens, distribute to validators, claim staking rewards
	// 1 hour later
	_, resBeginBlock := suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{{
					Validator:       val,
					SignedLastBlock: true,
				}},
			},
		},
	)

	// check time and factors
	suite.StoredEarnTimeEquals(vaultDenom1, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(vaultDenom2, suite.Ctx.BlockTime())

	validatorRewards, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	suite.Require().Contains(validatorRewards, suite.valAddrs[0].String(), "there should be claim events for validator 1")
	suite.Require().Contains(validatorRewards, suite.valAddrs[1].String(), "there should be claim events for validator 2")

	// Total staking rewards / total source shares (**deposited in earn** not total minted)
	// types.RewardIndexes.Quo() uses Dec.Quo() which uses bankers rounding.
	// So we need to use Dec.Quo() to also round vs Dec.QuoInt() which truncates
	expectedIndexes1 := validatorRewards[suite.valAddrs[0].String()].
		AmountOf("ukava").
		ToDec().
		Quo(userDepositAmount0.ToDec())

	expectedIndexes2 := validatorRewards[suite.valAddrs[1].String()].
		AmountOf("ukava").
		ToDec().
		Quo(userDepositAmount1.ToDec())

	// Only contains staking rewards
	suite.StoredEarnIndexesEqual(vaultDenom1, types.RewardIndexes{
		{
			CollateralType: "ukava",
			RewardFactor:   initialVault1RewardFactor.Add(expectedIndexes1),
		},
	})

	suite.StoredEarnIndexesEqual(vaultDenom2, types.RewardIndexes{
		{
			CollateralType: "ukava",
			RewardFactor:   initialVault2RewardFactor.Add(expectedIndexes2),
		},
	})
}
