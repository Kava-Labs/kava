package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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

	usdcDenom := "erc20/multichain/usdc"
	usdcSupply := int64(2_500_000_000000)

	earnKeeper := newFakeEarnKeeper().
		addVault("bkava-asdf", earntypes.NewVaultShare("bkava-asdf", sdk.NewDec(liquidStakedTokens))).
		addVault(usdcDenom, earntypes.NewVaultShare(usdcDenom, sdk.NewDec(usdcSupply)))

	suite.keeper = suite.NewTestKeeper(&fakeParamSubspace{}).
		WithDistrKeeper(
			newFakeDistrKeeper().setCommunityTax(communityTax),
		).
		WithMintKeeper(
			newFakeMintKeeper().
				setMinter(minttypes.NewMinter(inflation, sdk.OneDec())),
		).
		WithStakingKeeper(
			newFakeStakingKeeper().addBondedTokens(bondedTokens),
		).
		WithBankKeeper(
			newFakeBankKeeper().setSupply(sdk.NewCoin(types.BondDenom, sdkmath.NewInt(totalSupply))),
		).
		WithEarnKeeper(earnKeeper).
		WithLiquidKeeper(
			newFakeLiquidKeeper().addDerivative(suite.ctx, "bkava-asdf", sdkmath.NewInt(liquidStakedTokens)),
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
					sdk.NewCoin("ukava", sdkmath.NewInt(190258)),
				),
			},
			{
				Active:         true,
				CollateralType: "erc20/multichain/usdc",
				Start:          suite.ctx.BlockTime().Add(-time.Hour),
				End:            suite.ctx.BlockTime().Add(time.Hour),
				RewardsPerSecond: sdk.NewCoins(
					sdk.NewCoin("ukava", sdkmath.NewInt(5284)),
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
