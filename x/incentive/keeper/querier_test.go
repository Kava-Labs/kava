package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
)

type QuerierTestSuite struct {
	unitTester
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}

func (suite *QuerierTestSuite) TestGetStakingAPR() {
	communityTax := sdk.MustNewDecFromStr("0.90")
	inflation := sdk.MustNewDecFromStr("0.75")

	bondedTokens := int64(120_000_000_000000)
	liquidStakedTokens := int64(60_000_000_000000)
	totalSupply := int64(289_138_414_286684)

	// inflation values below are used to regression test the switch from x/mint to x/kavamint
	// rather than define the total inflation w/ a community tax, we now directly define
	// inflation for staking rewards & inflation for the community pool.
	// derive these values from the above values in order to verify no change to output
	bondedRatio := sdk.NewDec(bondedTokens).Quo(sdk.NewDec(totalSupply))
	communityInflation := inflation.
		Mul(communityTax).
		Quo(bondedRatio)
	stakingRewardsApy := inflation.
		Mul(sdk.OneDec().Sub(communityTax)).
		Quo(bondedRatio)

	usdcDenom := "erc20/multichain/usdc"
	usdcSupply := int64(2_500_000_000000)

	earnKeeper := newFakeEarnKeeper().
		addVault("bkava-asdf", earntypes.NewVaultShare("bkava-asdf", sdk.NewDec(liquidStakedTokens))).
		addVault(usdcDenom, earntypes.NewVaultShare(usdcDenom, sdk.NewDec(usdcSupply)))

	suite.keeper = suite.NewTestKeeper(&fakeParamSubspace{}).
		WithKavamintKeeper(
			newFakeKavamintKeeper().
				setCommunityInflation(communityInflation).
				setStakingApy(stakingRewardsApy),
		).
		WithStakingKeeper(
			newFakeStakingKeeper().addBondedTokens(bondedTokens),
		).
		WithBankKeeper(
			newFakeBankKeeper().setSupply(sdk.NewCoin(types.BondDenom, sdk.NewInt(totalSupply))),
		).
		WithEarnKeeper(earnKeeper).
		WithLiquidKeeper(
			newFakeLiquidKeeper().addDerivative(suite.ctx, "bkava-asdf", sdk.NewInt(liquidStakedTokens)),
		).
		WithPricefeedKeeper(
			newFakePricefeedKeeper().
				setPrice(pricefeedtypes.NewCurrentPrice("kava:usd:30", sdk.MustNewDecFromStr("1.5"))).
				setPrice(pricefeedtypes.NewCurrentPrice("usdc:usd:30", sdk.OneDec())),
		).
		Build()

	// ~18% APR
	expectedStakingAPY := inflation.
		Mul(sdk.OneDec().Sub(communityTax)).
		Quo(sdk.NewDec(bondedTokens).Quo(sdk.NewDec(totalSupply)))

	// Staking APR = (Inflation Rate * (1 - Community Tax)) / (Bonded Tokens / Circulating Supply)
	aprWithoutIncentives, err := keeper.GetStakingAPR(suite.ctx, suite.keeper, types.Params{})
	suite.Require().NoError(err)
	suite.Require().Equal(
		expectedStakingAPY,
		aprWithoutIncentives,
	)

	suite.T().Logf("Staking APR without incentives: %s", aprWithoutIncentives)

	params := types.Params{
		EarnRewardPeriods: types.MultiRewardPeriods{
			{
				Active:         true,
				CollateralType: "bkava",
				Start:          suite.ctx.BlockTime().Add(-time.Hour),
				End:            suite.ctx.BlockTime().Add(time.Hour),
				RewardsPerSecond: sdk.NewCoins(
					sdk.NewCoin("ukava", sdk.NewInt(190258)),
				),
			},
			{
				Active:         true,
				CollateralType: "erc20/multichain/usdc",
				Start:          suite.ctx.BlockTime().Add(-time.Hour),
				End:            suite.ctx.BlockTime().Add(time.Hour),
				RewardsPerSecond: sdk.NewCoins(
					sdk.NewCoin("ukava", sdk.NewInt(5284)),
				),
			},
		},
	}

	suite.Run("GetStakingAPR", func() {
		aprWithIncentives, err := keeper.GetStakingAPR(suite.ctx, suite.keeper, params)
		suite.Require().NoError(err)
		// Approx 10% increase in APR from incentives
		suite.Require().Equal(sdk.MustNewDecFromStr("0.280711113729177500"), aprWithIncentives)

		suite.Require().Truef(
			aprWithIncentives.GT(aprWithoutIncentives),
			"APR with incentives (%s) should be greater than APR without incentives (%s)",
		)
	})

	suite.Run("GetAPYFromMultiRewardPeriod", func() {
		vaultTotalValue, err := earnKeeper.GetVaultTotalValue(suite.ctx, usdcDenom)
		suite.Require().NoError(err)
		suite.Require().True(vaultTotalValue.Amount.IsPositive())

		apy, err := keeper.GetAPYFromMultiRewardPeriod(
			suite.ctx,
			suite.keeper,
			usdcDenom,
			params.EarnRewardPeriods[1],
			vaultTotalValue.Amount,
		)
		suite.Require().NoError(err)
		suite.Require().Equal(
			sdk.MustNewDecFromStr("0.099981734400000000"),
			apy,
			"usdc apy should be approx 10%",
		)
	})
}
