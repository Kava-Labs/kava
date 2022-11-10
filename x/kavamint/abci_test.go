package kavamint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	communitytypes "github.com/kava-labs/kava/x/community/types"
	"github.com/kava-labs/kava/x/kavamint"
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

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) TestBeginBlockerMintsExpectedTokens() {
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
			name:                    "mints staking rewards, handles 0 community pool inflation",
			blockTime:               6,
			communityPoolInflation:  sdk.ZeroDec(),
			stakingRewardsApy:       sdk.NewDecWithPrec(20, 2),
			bondedRatio:             sdk.OneDec(),
			expCommunityPoolBalance: sdk.ZeroInt(),
			// 20% APY for 6 seconds
			// bond ratio is 100%, so total supply = bonded supply = 1e10
			// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.20%2C31536000%5D%5C%2841%29%2C6%5D-1%5C%2841%29*1e10
			// => 346.88 => truncated to 346 tokens.
			expFeeCollectorBalance: sdk.NewInt(346),
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
			startBlockTime, startTimeFound := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
			suite.Require().True(startTimeFound)
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
			endBlockTime, endTimeFound := suite.Keeper.GetPreviousBlockTime(ctx2)
			suite.Require().True(endTimeFound)
			suite.Require().Equal(ctx2.BlockTime(), endBlockTime)
		})
	}
}
