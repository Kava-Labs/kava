package keeper_test

import (
	"testing"
	"time"

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
			newFakeBankKeeper().setSupply(sdk.NewCoin(types.BondDenom, sdk.NewInt(totalSupply))),
		).
		WithEarnKeeper(
			newFakeEarnKeeper().
				addVault("bkava-asdf", earntypes.NewVaultShare("bkava-asdf", sdk.NewDec(liquidStakedTokens))),
		).
		WithLiquidKeeper(
			newFakeLiquidKeeper().addDerivative(suite.ctx, "bkava-asdf", sdk.NewInt(liquidStakedTokens)),
		).
		WithPricefeedKeeper(
			newFakePricefeedKeeper().
				setPrice(
					pricefeedtypes.NewCurrentPrice(
						"ukava:usd:30",
						sdk.MustNewDecFromStr("1.5"),
					)),
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
		},
	}

	aprWithIncentives, err := keeper.GetStakingAPR(suite.ctx, suite.keeper, params)
	suite.Require().NoError(err)
	// Approx 10% increase in APR from incentives
	suite.Require().Equal(sdk.MustNewDecFromStr("0.280711113729177500"), aprWithIncentives)

	suite.Require().Truef(
		aprWithIncentives.GT(aprWithoutIncentives),
		"APR with incentives (%s) should be greater than APR without incentives (%s)",
	)
}
