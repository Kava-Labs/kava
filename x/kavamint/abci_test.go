package kavamint_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/suite"

	communitytypes "github.com/kava-labs/kava/x/community/types"
	"github.com/kava-labs/kava/x/kavamint"
	"github.com/kava-labs/kava/x/kavamint/keeper"
	"github.com/kava-labs/kava/x/kavamint/testutil"
	"github.com/kava-labs/kava/x/kavamint/types"
)

type abciTestSuite struct {
	testutil.KavamintTestSuite
}

func (suite *abciTestSuite) SetupTest() {
	suite.KavamintTestSuite.SetupTest()
}

func (suite abciTestSuite) CheckModuleBalance(ctx sdk.Context, moduleName string, expectedAmount sdk.Int) {
	denom := suite.StakingKeeper.BondDenom(ctx)
	amount := suite.App.GetModuleAccountBalance(ctx, moduleName, denom)
	suite.Require().Equal(expectedAmount, amount)
}

func (suite *abciTestSuite) CheckFeeCollectorBalance(ctx sdk.Context, expectedAmount sdk.Int) {
	suite.CheckModuleBalance(ctx, authtypes.FeeCollectorName, expectedAmount)
}

func (suite *abciTestSuite) CheckKavamintBalance(ctx sdk.Context, expectedAmount sdk.Int) {
	suite.CheckModuleBalance(ctx, types.ModuleName, expectedAmount)
}

func (suite *abciTestSuite) CheckCommunityPoolBalance(ctx sdk.Context, expectedAmount sdk.Int) {
	suite.CheckModuleBalance(ctx, communitytypes.ModuleAccountName, expectedAmount)
}

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) Test_BeginBlocker_MintsExpectedTokens() {
	testCases := []struct {
		name                    string
		blockTime               uint64
		communityPoolInflation  sdk.Dec
		stakingRewardsApy       sdk.Dec
		bondedRatio             sdk.Dec
		setup                   func()
		expCommunityPoolBalance sdk.Int
		expFeeCollectorBalance  sdk.Int
	}{
		{
			name:                   "sanity check: a year of seconds mints total yearly inflation",
			blockTime:              keeper.SecondsPerYear,
			communityPoolInflation: sdk.NewDecWithPrec(20, 2),
			stakingRewardsApy:      sdk.NewDecWithPrec(20, 2),
			bondedRatio:            sdk.NewDecWithPrec(50, 2),
			// 20% inflation on 1e10 tokens -> 2e9 minted
			expCommunityPoolBalance: sdk.NewInt(2e9),
			// 20% APY, 50% bonded (5e9) -> 1e9 minted
			expFeeCollectorBalance: sdk.NewInt(1e9),
		},
		{
			name:                    "mints staking rewards, handles 0 community pool inflation",
			blockTime:               6,
			communityPoolInflation:  sdk.ZeroDec(),
			stakingRewardsApy:       sdk.NewDecWithPrec(20, 2),
			bondedRatio:             sdk.NewDecWithPrec(40, 2),
			expCommunityPoolBalance: sdk.ZeroInt(),
			// 20% APY for 6 seconds
			// bond ratio is 40%, so total supply * ratio = 1e10 * .4
			// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.20%2C31536000%5D%5C%2841%29%2C6%5D-1%5C%2841%29*1e10*.4
			// => 138.75 => truncated to 138 tokens.
			expFeeCollectorBalance: sdk.NewInt(138),
		},
		{
			name:                   "mints community pool inflation, handles 0 staking rewards",
			blockTime:              6,
			communityPoolInflation: sdk.NewDecWithPrec(80, 2),
			stakingRewardsApy:      sdk.ZeroDec(),
			bondedRatio:            sdk.NewDecWithPrec(40, 2),
			// 80% APY for 6 seconds
			// total supply = 1e10
			// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.80%2C31536000%5D%5C%2841%29%2C6%5D-1%5C%2841%29*1e10
			// => 1118.32 => truncated to 1118 tokens.
			expCommunityPoolBalance: sdk.NewInt(1118),
			expFeeCollectorBalance:  sdk.ZeroInt(),
		},
		{
			name:                    "mints no tokens if all inflation is zero",
			blockTime:               6,
			communityPoolInflation:  sdk.ZeroDec(),
			stakingRewardsApy:       sdk.ZeroDec(),
			bondedRatio:             sdk.NewDecWithPrec(40, 2),
			expCommunityPoolBalance: sdk.ZeroInt(),
			expFeeCollectorBalance:  sdk.ZeroInt(),
		},
		{
			name:                   "mints community pool inflation and staking rewards",
			blockTime:              6,
			communityPoolInflation: sdk.NewDecWithPrec(50, 2),
			stakingRewardsApy:      sdk.NewDecWithPrec(20, 2),
			bondedRatio:            sdk.NewDecWithPrec(35, 2),
			// 50% APY for 6 seconds
			// total supply = 1e10
			// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.5%2C31536000%5D%5C%2841%29%2C6%5D-1%5C%2841%29*1e10
			// => 771.43 => truncated to 771 tokens.
			expCommunityPoolBalance: sdk.NewInt(771),
			// 20% APY for 6 seconds
			// total bonded = 1e10 * 35%
			// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.20%2C31536000%5D%5C%2841%29%2C6%5D-1%5C%2841%29*1e10*.35
			// => 121.41 => truncated to 121 tokens.
			expFeeCollectorBalance: sdk.NewInt(121),
		},
		{
			name:                   "handles extra long block time",
			blockTime:              60, // like if we're upgrading the network & it takes an hour to get back up
			communityPoolInflation: sdk.NewDecWithPrec(50, 2),
			stakingRewardsApy:      sdk.NewDecWithPrec(20, 2),
			bondedRatio:            sdk.NewDecWithPrec(35, 2),
			// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.5%2C31536000%5D%5C%2841%29%2C60%5D-1%5C%2841%29*1e10
			expCommunityPoolBalance: sdk.NewInt(7714),
			// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.20%2C31536000%5D%5C%2841%29%2C60%5D-1%5C%2841%29*1e10*.35
			expFeeCollectorBalance: sdk.NewInt(1214),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// set store and params
			suite.Keeper.SetPreviousBlockTime(suite.Ctx, suite.Ctx.BlockTime())
			suite.Keeper.SetParams(
				suite.Ctx,
				types.NewParams(tc.communityPoolInflation, tc.stakingRewardsApy),
			)

			// set bonded token ratio
			suite.SetBondedTokenRatio(tc.bondedRatio)
			staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

			// run begin blocker
			kavamint.BeginBlocker(suite.Ctx, suite.Keeper)

			// expect everything empty to start
			suite.CheckFeeCollectorBalance(suite.Ctx, sdk.ZeroInt())
			suite.CheckKavamintBalance(suite.Ctx, sdk.ZeroInt())
			suite.CheckCommunityPoolBalance(suite.Ctx, sdk.ZeroInt())

			// expect initial block time set
			startBlockTime := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
			suite.Require().False(startBlockTime.IsZero())
			suite.Require().Equal(suite.Ctx.BlockTime(), startBlockTime)

			// run begin blocker again to mint inflation
			ctx2 := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Second * time.Duration(tc.blockTime)))
			kavamint.BeginBlocker(ctx2, suite.Keeper)

			// check expected balances
			suite.CheckCommunityPoolBalance(ctx2, tc.expCommunityPoolBalance)
			suite.CheckFeeCollectorBalance(ctx2, tc.expFeeCollectorBalance)

			// x/kavamint balance should always be 0 because 100% should be transferred out every block
			suite.CheckKavamintBalance(ctx2, sdk.ZeroInt())

			// expect time to be updated
			endBlockTime := suite.Keeper.GetPreviousBlockTime(ctx2)
			suite.Require().False(endBlockTime.IsZero())
			suite.Require().Equal(ctx2.BlockTime(), endBlockTime)
		})
	}
}

func (suite *abciTestSuite) Test_BeginBlocker_DefaultsToBlockTime() {
	suite.SetupTest()

	// unset previous block time
	suite.Keeper.SetPreviousBlockTime(suite.Ctx, time.Time{})

	// run begin blocker
	kavamint.BeginBlocker(suite.Ctx, suite.Keeper)

	// ensure block time gets set
	blockTime := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
	suite.False(blockTime.IsZero())
	suite.Equal(suite.Ctx.BlockTime(), blockTime)
}
