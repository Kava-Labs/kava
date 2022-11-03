package kavamint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

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

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) TestBeginBlockerMintsStakingRewards() {
	ctx, kavamintKeeper := suite.Ctx, suite.KavamintTestSuite.Keeper

	bondDenom := kavamintKeeper.BondDenom(ctx)
	bondedRatio := sdk.OneDec()
	blockTime := uint64(6) // 6 seconds
	stakingApy := sdk.NewDecWithPrec(20, 2)

	kavamintKeeper.SetParams(ctx, types.NewParams(sdk.ZeroDec(), stakingApy))
	kavamintKeeper.SetPreviousBlockTime(ctx, ctx.BlockTime())

	// determine factor based on 20% APY, compounded per second, for 6 seconds
	rate, err := keeper.CalculateInflationRate(stakingApy, blockTime)
	suite.Require().NoError(err)

	// set bonded token ratio
	totalSupply := suite.SetBondedTokenRatio(bondedRatio)
	staking.EndBlocker(ctx, suite.StakingKeeper)

	kavamint.BeginBlocker(ctx, kavamintKeeper)

	// expect nothing added to fee pool.
	suite.CheckFeeCollectorBalance(ctx, sdk.ZeroInt())
	// expect nothing in kavamint module
	suite.CheckKavamintBalance(ctx, sdk.ZeroInt())

	// expect block time set
	startBlockTime, startTimeFound := kavamintKeeper.GetPreviousBlockTime(ctx)
	suite.Require().True(startTimeFound)
	suite.Require().Equal(ctx.BlockTime(), startBlockTime)

	// begin blocker again
	ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(time.Second * time.Duration(blockTime)))
	kavamint.BeginBlocker(ctx2, kavamintKeeper)

	// expect amount added to fee pool.
	expectedAmount := rate.MulInt(totalSupply.AmountOf(bondDenom)).TruncateInt()
	suite.CheckFeeCollectorBalance(ctx2, expectedAmount)
	// now, ensure that amount is what we expect
	// 20% APY for 6 seconds
	// bond ratio is 100%, so total supply = bonded supply = 1e10
	// https://www.wolframalpha.com/input?i2d=true&i=%5C%2840%29Power%5B%5C%2840%29Surd%5B1.20%2C31536000%5D%5C%2841%29%2C6%5D-1%5C%2841%29*1e10
	// => 346.88 => truncated to 346 tokens.
	suite.CheckFeeCollectorBalance(ctx2, sdk.NewInt(346))

	// kavamint balance should still be 0 because 100% was transferred out
	suite.CheckKavamintBalance(ctx2, sdk.ZeroInt())

	// expect time to be updated
	endBlockTime, endTimeFound := kavamintKeeper.GetPreviousBlockTime(ctx)
	suite.Require().True(endTimeFound)
	suite.Require().Equal(ctx2.BlockTime(), endBlockTime)
}
