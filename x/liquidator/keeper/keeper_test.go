package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquidator/types"
)

func TestKeeper_SeizeAndStartCollateralAuction(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)

	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates(
		app.NewAuthGenState(addrs, []sdk.Coins{cs(c("btc", 100))}),
		NewPFGenState("btc", sdk.MustNewDecFromStr("8000.00")),
		NewCDPGenState(),
		NewLiquidatorGenState(),
	)
	ctx := tApp.NewContext(false, abci.Header{})

	require.NoError(t, tApp.GetCDPKeeper().ModifyCDP(ctx, addrs[0], "btc", i(3), i(16000)))

	_, err := tApp.GetPriceFeedKeeper().SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("7999.99"), time.Unix(9999999999, 0))
	require.NoError(t, err)
	require.NoError(t, tApp.GetPriceFeedKeeper().SetCurrentPrices(ctx, "btc"))

	// Run test function
	auctionID, err := tApp.GetLiquidatorKeeper().SeizeAndStartCollateralAuction(ctx, addrs[0], "btc")

	// Check CDP
	require.NoError(t, err)
	cdp, found := tApp.GetCDPKeeper().GetCDP(ctx, addrs[0], "btc")
	require.True(t, found)
	require.Equal(t, cdp.CollateralAmount, i(2)) // original amount - params.CollateralAuctionSize
	require.Equal(t, cdp.Debt, i(10667))         // original debt scaled by amount of collateral removed
	// Check auction exists
	_, found = tApp.GetAuctionKeeper().GetAuction(ctx, auctionID)
	require.True(t, found)
}

func TestKeeper_StartDebtAuction(t *testing.T) {
	// Setup
	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates(
		NewLiquidatorGenState(),
	)
	keeper := tApp.GetLiquidatorKeeper()
	ctx := tApp.NewContext(false, abci.Header{})

	initSDebt := types.SeizedDebt{i(2000), i(0)}
	keeper.SetSeizedDebt(ctx, initSDebt)

	// Execute
	auctionID, err := keeper.StartDebtAuction(ctx)

	// Check
	require.NoError(t, err)
	require.Equal(t,
		types.SeizedDebt{
			initSDebt.Total,
			initSDebt.SentToAuction.Add(keeper.GetParams(ctx).DebtAuctionSize),
		},
		keeper.GetSeizedDebt(ctx),
	)
	_, found := tApp.GetAuctionKeeper().GetAuction(ctx, auctionID)
	require.True(t, found)
}

func TestKeeper_partialSeizeCDP(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)

	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates(
		app.NewAuthGenState(addrs, []sdk.Coins{cs(c("btc", 100))}),
		NewPFGenState("btc", sdk.MustNewDecFromStr("8000.00")),
		NewCDPGenState(),
		NewLiquidatorGenState(),
	)
	ctx := tApp.NewContext(false, abci.Header{})

	tApp.GetCDPKeeper().ModifyCDP(ctx, addrs[0], "btc", i(3), i(16000))

	tApp.GetPriceFeedKeeper().SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("7999.99"), tmtime.Now().Add(time.Hour*1))
	tApp.GetPriceFeedKeeper().SetCurrentPrices(ctx, "btc")

	// Run test function
	err := tApp.GetLiquidatorKeeper().PartialSeizeCDP(ctx, addrs[0], "btc", i(2), i(10000))

	// Check
	require.NoError(t, err)
	cdp, found := tApp.GetCDPKeeper().GetCDP(ctx, addrs[0], "btc")
	require.True(t, found)
	require.Equal(t, i(1), cdp.CollateralAmount)
	require.Equal(t, i(6000), cdp.Debt)
}

func TestKeeper_GetSetSeizedDebt(t *testing.T) {
	// Setup
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{})
	debt := types.SeizedDebt{i(234247645), i(2343)}

	// Run test function
	tApp.GetLiquidatorKeeper().SetSeizedDebt(ctx, debt)
	readDebt := tApp.GetLiquidatorKeeper().GetSeizedDebt(ctx)

	// Check
	require.Equal(t, debt, readDebt)
}
