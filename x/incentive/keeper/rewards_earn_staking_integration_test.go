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

	keeper TestKeeper
	addrs  []sdk.AccAddress
}

func TestEarnStakingRewardsIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(EarnStakingRewardsIntegrationTestSuite))
}

func (suite *EarnStakingRewardsIntegrationTestSuite) SetupTest() {
	suite.IntegrationTester.SetupTest()

	suite.keeper = TestKeeper{
		Keeper: suite.App.GetIncentiveKeeper(),
	}

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *EarnStakingRewardsIntegrationTestSuite) SetDenoms() {
	mk := suite.App.GetMintKeeper()

	// Use ukava for mint denom
	mParams := mk.GetParams(suite.Ctx)
	mParams.MintDenom = "ukava"

	mk.SetParams(suite.Ctx, mParams)
}

func (suite *EarnStakingRewardsIntegrationTestSuite) TestStakingRewardsDistributed() {
	sk := suite.App.GetStakingKeeper()

	userAddr1, userAddr2, validatorAddr1, validatorAddr2 := suite.addrs[0],
		suite.addrs[1],
		suite.addrs[2],
		suite.addrs[3]

	valAddr1 := sdk.ValAddress(validatorAddr1)
	valAddr2 := sdk.ValAddress(validatorAddr2)

	// Setup app with test state
	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleAccount(userAddr1, cs(c("ukava", 1e12))).
		WithSimpleAccount(userAddr2, cs(c("ukava", 1e12))).
		WithSimpleAccount(validatorAddr1, cs(c("ukava", 1e12))).
		WithSimpleAccount(validatorAddr2, cs(c("ukava", 1e12)))

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
		WithMinter(sdk.OneDec(), sdk.ZeroDec())

	suite.StartChainWithBuilders(
		authBuilder,
		incentiveBuilder,
		savingsBuilder,
		earnBuilder,
		stakingBuilder,
		mintBuilder,
	)

	suite.SetDenoms()

	selfDelegationAmount := c("ukava", 1e9)

	// derivative 1: 8 total staked, 7 to earn, 1 not in earn
	// derivative 2: 2 total staked, 1 to earn, 1 not in earn
	userDepositAmount1 := c("ukava", 7e9)
	userDepositAmount2 := c("ukava", 1e9)

	userMintAmount1 := c("ukava", 1e9)
	userMintAmount2 := c("ukava", 1e9)

	// Create two validators
	err := suite.DeliverMsgCreateValidator(valAddr1, selfDelegationAmount)
	suite.Require().NoError(err)

	err = suite.DeliverMsgCreateValidator(valAddr2, selfDelegationAmount)
	suite.Require().NoError(err)

	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	// Delegate to validator, mint derivative and deposit to earn
	err = suite.DeliverRouterMsgDelegateMintDeposit(userAddr1, valAddr1, userDepositAmount1)
	suite.Require().NoError(err)

	err = suite.DeliverRouterMsgDelegateMintDeposit(userAddr1, valAddr2, userDepositAmount2)
	suite.Require().NoError(err)

	// Additional delegate + mint derivative that is **not** deposited to earn
	err = suite.DeliverMsgDelegateMint(userAddr1, valAddr1, userMintAmount1)
	suite.Require().NoError(err)

	err = suite.DeliverMsgDelegateMint(userAddr1, valAddr2, userMintAmount2)
	suite.Require().NoError(err)

	// Get derivative denoms
	lq := suite.App.GetLiquidKeeper()
	vaultDenom1 := lq.GetLiquidStakingTokenDenom(valAddr1)
	vaultDenom2 := lq.GetLiquidStakingTokenDenom(valAddr2)

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

	validator1, found := sk.GetValidator(suite.Ctx, valAddr1)
	suite.Require().True(found)

	pk, err := validator1.ConsPubKey()
	suite.Require().NoError(err)

	val := abci.Validator{
		Address: pk.Address(),
		Power:   100,
	}

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

	suite.Require().Contains(validatorRewards, valAddr1.String(), "there should be claim events for validator 1")
	suite.Require().Contains(validatorRewards, valAddr2.String(), "there should be claim events for validator 2")

	// Total staking rewards / total source shares (**deposited in earn** not total minted)
	// types.RewardIndexes.Quo() uses Dec.Quo() which uses bankers rounding.
	// So we need to use Dec.Quo() to also round vs Dec.QuoInt() which truncates
	expectedIndexes1 := validatorRewards[valAddr1.String()].
		AmountOf("ukava").
		ToDec().
		Quo(userDepositAmount1.Amount.ToDec())

	expectedIndexes2 := validatorRewards[valAddr2.String()].
		AmountOf("ukava").
		ToDec().
		Quo(userDepositAmount2.Amount.ToDec())

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
