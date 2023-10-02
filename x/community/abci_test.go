package community_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community"
	"github.com/kava-labs/kava/x/community/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestABCIStakingRewardsArePaidOutOnDisableInflationBlock(t *testing.T) {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates()
	keeper := tApp.GetCommunityKeeper()
	accountKeeper := tApp.GetAccountKeeper()
	bankKeeper := tApp.GetBankKeeper()

	// a block that runs after addition of the disable inflation code on chain
	// but before the disable inflation time
	initialBlockTime := time.Now()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: initialBlockTime})

	poolAcc := accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	feeCollectorAcc := accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	disableTime := initialBlockTime.Add(9 * time.Second)

	// set state
	params, _ := keeper.GetParams(ctx)
	params.UpgradeTimeDisableInflation = disableTime
	params.UpgradeTimeSetStakingRewardsPerSecond = sdkmath.LegacyNewDec(1000000) // 1 KAVA
	params.StakingRewardsPerSecond = sdkmath.LegacyZeroDec()
	keeper.SetParams(ctx, params)

	// fund community pool account
	tApp.FundAccount(ctx, poolAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(10000000)))) // 10 KAVA
	initialFeeCollectorBalance := bankKeeper.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava").Amount

	// run one block
	community.BeginBlocker(ctx, keeper)

	// assert that staking rewards in parameters are still set to zero
	params, found := keeper.GetParams(ctx)
	require.True(t, found)
	require.Equal(t, sdkmath.LegacyZeroDec(), params.StakingRewardsPerSecond)

	// assert no rewards are given yet
	rewards := bankKeeper.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava").Amount.Sub(initialFeeCollectorBalance)
	require.Equal(t, sdkmath.ZeroInt(), rewards)

	// new block when disable inflation runs, 10 seconds from initial block for easy math
	blockTime := disableTime.Add(1 * time.Second)
	ctx = tApp.NewContext(true, tmproto.Header{Height: ctx.BlockHeight() + 1, Time: blockTime})

	// run the next block
	community.BeginBlocker(ctx, keeper)

	// assert that staking rewards have been set and disable inflation time is zero
	params, found = keeper.GetParams(ctx)
	require.True(t, found)
	require.True(t, params.UpgradeTimeDisableInflation.IsZero())
	require.Equal(t, sdkmath.LegacyNewDec(1000000), params.StakingRewardsPerSecond)

	// assert that 10 KAVA has been distributed in rewards
	rewards = bankKeeper.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava").Amount.Sub(initialFeeCollectorBalance)
	require.Equal(t, sdkmath.NewInt(10000000).String(), rewards.String())
}
