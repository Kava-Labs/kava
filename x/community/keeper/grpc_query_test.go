package keeper_test

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
)

type grpcQueryTestSuite struct {
	testutil.Suite

	queryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.Suite.SetupTest()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestGrpcQueryBalance() {
	var expCoins sdk.Coins

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			name:  "handles response with no balance",
			setup: func() { expCoins = sdk.Coins{} },
		},
		{
			name: "handles response with balance",
			setup: func() {
				expCoins = sdk.NewCoins(
					sdk.NewCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewCoin("usdx", sdkmath.NewInt(1000)),
				)
				suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, expCoins)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()
			res, err := suite.queryClient.Balance(context.Background(), &types.QueryBalanceRequest{})
			suite.Require().NoError(err)
			suite.Require().True(expCoins.IsEqual(res.Coins))
		})
	}
}

func (suite *grpcQueryTestSuite) TestGrpcQueryTotalBalance() {
	var expCoins sdk.DecCoins

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			name:  "handles response with no balance",
			setup: func() { expCoins = sdk.DecCoins{} },
		},
		{
			name: "handles response with balance",
			setup: func() {
				expCoins = sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdx", sdkmath.NewInt(1000)),
				)

				coins, _ := expCoins.TruncateDecimal()

				suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, coins)
			},
		},
		{
			name: "handles response with both x/community + x/distribution balance",
			setup: func() {
				decCoins1 := sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdx", sdkmath.NewInt(1000)),
				)

				coins, _ := decCoins1.TruncateDecimal()

				err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				decCoins2 := sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdc", sdkmath.NewInt(1000)),
				)

				// Add to x/distribution community pool (just state, not actual coins)
				dk := suite.App.GetDistrKeeper()
				feePool := dk.GetFeePool(suite.Ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(decCoins2...)
				dk.SetFeePool(suite.Ctx, feePool)

				expCoins = decCoins1.Add(decCoins2...)
			},
		},
		{
			name: "handles response with only x/distribution balance",
			setup: func() {
				expCoins = sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdc", sdkmath.NewInt(1000)),
				)

				// Add to x/distribution community pool (just state, not actual coins)
				dk := suite.App.GetDistrKeeper()
				feePool := dk.GetFeePool(suite.Ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(expCoins...)
				dk.SetFeePool(suite.Ctx, feePool)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()
			res, err := suite.queryClient.TotalBalance(context.Background(), &types.QueryTotalBalanceRequest{})
			suite.Require().NoError(err)
			suite.Require().True(expCoins.IsEqual(res.Pool))
		})
	}
}

// backported from v0.25.x. Does not actually use `rewardsPerSec` because concept does not exist.
// NOTE: this test makes use of the fact that there is always an initial 1e6 bonded tokens
// To adjust the bonded ratio, it adjusts the total supply by minting tokens.
func (suite *grpcQueryTestSuite) TestGrpcQueryAnnualizedRewards() {
	testCases := []struct {
		name          string
		bondedRatio   sdk.Dec
		inflation     sdk.Dec
		rewardsPerSec sdkmath.LegacyDec
		communityTax  sdk.Dec
		expectedRate  sdkmath.LegacyDec
	}{
		{
			name:          "sanity check: no inflation, no rewards => 0%",
			bondedRatio:   sdk.MustNewDecFromStr("0.3456"),
			inflation:     sdk.ZeroDec(),
			rewardsPerSec: sdkmath.LegacyZeroDec(),
			expectedRate:  sdkmath.LegacyZeroDec(),
		},
		{
			name:          "inflation sanity check: 100% inflation, 100% bonded => 100%",
			bondedRatio:   sdk.OneDec(),
			inflation:     sdk.OneDec(),
			rewardsPerSec: sdkmath.LegacyZeroDec(),
			expectedRate:  sdkmath.LegacyOneDec(),
		},
		{
			name:          "inflation sanity check: 100% community tax => 0%",
			bondedRatio:   sdk.OneDec(),
			inflation:     sdk.OneDec(),
			communityTax:  sdk.OneDec(),
			rewardsPerSec: sdkmath.LegacyZeroDec(),
			expectedRate:  sdkmath.LegacyZeroDec(),
		},
		{
			name:          "inflation enabled: realistic example",
			bondedRatio:   sdk.MustNewDecFromStr("0.148"),
			inflation:     sdk.MustNewDecFromStr("0.595"),
			communityTax:  sdk.MustNewDecFromStr("0.9495"),
			rewardsPerSec: sdkmath.LegacyZeroDec(),
			// expect ~20.23%
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.203023625910000000"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// set inflation
			mk := suite.App.GetMintKeeper()
			minter := mk.GetMinter(suite.Ctx)
			minter.Inflation = tc.inflation
			mk.SetMinter(suite.Ctx, minter)

			// set community tax
			communityTax := sdk.ZeroDec()
			if !tc.communityTax.IsNil() {
				communityTax = tc.communityTax
			}
			dk := suite.App.GetDistrKeeper()
			distParams := dk.GetParams(suite.Ctx)
			distParams.CommunityTax = communityTax
			dk.SetParams(suite.Ctx, distParams)

			// set bonded tokens
			suite.adjustBondedRatio(tc.bondedRatio)

			// query for annualized rewards
			res, err := suite.queryClient.AnnualizedRewards(suite.Ctx, &types.QueryAnnualizedRewardsRequest{})
			// verify results match expected
			suite.Require().NoError(err)
			suite.Equal(tc.expectedRate, res.StakingRewards)
		})
	}
}

// adjustBondRatio changes the ratio of bonded coins
// it leverages the fact that there is a constant number of bonded tokens
// and adjusts the total supply to make change the bonded ratio.
// returns the new total supply of the bond denom
func (suite *grpcQueryTestSuite) adjustBondedRatio(desiredRatio sdk.Dec) sdkmath.Int {
	// from the InitGenesis validator
	bondedTokens := sdkmath.NewInt(1e6)
	bondDenom := suite.App.GetStakingKeeper().BondDenom(suite.Ctx)

	// first, burn all non-delegated coins (bonded ratio = 100%)
	suite.App.DeleteGenesisValidatorCoins(suite.T(), suite.Ctx)

	if desiredRatio.Equal(sdk.OneDec()) {
		return bondedTokens
	}

	// mint new tokens to adjust the bond ratio
	newTotalSupply := sdk.NewDecFromInt(bondedTokens).Quo(desiredRatio).TruncateInt()
	coinsToMint := newTotalSupply.Sub(bondedTokens)
	err := suite.App.FundAccount(suite.Ctx, app.RandomAddress(), sdk.NewCoins(sdk.NewCoin(bondDenom, coinsToMint)))
	suite.Require().NoError(err)

	return newTotalSupply
}
