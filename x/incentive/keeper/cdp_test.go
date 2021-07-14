package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
)

func TestRiskyCDPsAccumulateRewards(t *testing.T) {
	genesisTime := time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
	_, addrs := app.GeneratePrivKeyAddressPairs(5)

	initialCollateral := c("bnb", 1_000_000_000)
	user := addrs[0]
	authBuilder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(user, cs(initialCollateral))

	collateralType := "bnb-a"
	rewardsPerSecond := c(types.USDXMintingRewardDenom, 1_000_000)

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(genesisTime).
		WithSimpleUSDXRewardPeriod(collateralType, rewardsPerSecond)

	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates(
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMultiFromTime(genesisTime),
		NewCDPGenStateMulti(),
		incentBuilder.BuildMarshalled(),
	)
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: genesisTime})

	// Setup cdp state containing one CDP
	cdpKeeper := tApp.GetCDPKeeper()
	err := cdpKeeper.AddCdp(ctx, user, initialCollateral, c("usdx", 100_000_000), collateralType)
	require.NoError(t, err)

	// Skip ahead two blocks to accumulate both interest and usdx reward for the cdp
	// Two blocks are required because the cdp begin blocker runs before incentive begin blocker.
	// In the first begin block the cdp is synced, which triggers its claim to sync. But no global rewards have accumulated yet so the sync does nothing.
	// Global rewards accumulate immediately after during the incentive begin blocker.
	// Rewards are added to the cdp's claim in the next block when the cdp is synced.
	_ = tApp.EndBlocker(ctx, abci.RequestEndBlock{})
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(10 * time.Minute))
	_ = tApp.BeginBlocker(ctx, abci.RequestBeginBlock{}) // height and time in header are ignored by module begin blockers

	_ = tApp.EndBlocker(ctx, abci.RequestEndBlock{})
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(10 * time.Minute))
	_ = tApp.BeginBlocker(ctx, abci.RequestBeginBlock{})

	// check cdp rewards
	cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, user, collateralType)
	require.True(t, found)
	// This additional sync adds the rewards accumulated at the end of the last begin block.
	// They weren't added during the begin blocker as the incentive BB runs after the CDP BB.
	incentiveKeeper := tApp.GetIncentiveKeeper()
	incentiveKeeper.SynchronizeUSDXMintingReward(ctx, cdp)
	claim, found := incentiveKeeper.GetUSDXMintingClaim(ctx, user)
	require.True(t, found)

	// rewards are roughly rewardsPerSecond * secondsElapsed (10mins) * num blocks (2)
	require.Equal(t, c(types.USDXMintingRewardDenom, 1_200_000_557), claim.Reward)
}
