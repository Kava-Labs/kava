package auction_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction"
)

func TestKeeper_EndBlocker(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	seller := addrs[0]

	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates(
		app.NewAuthGenState(addrs, []sdk.Coins{cs(c("token1", 100), c("token2", 100))}),
	)

	ctx := tApp.NewContext(true, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	auctionID, err := keeper.StartForwardAuction(ctx, seller, c("token1", 20), c("token2", 0))
	require.NoError(t, err)

	// Run the endblocker, simulating a block height just before auction expiry
	preExpiryHeight := ctx.BlockHeight() + int64(auction.DefaultMaxAuctionDuration) - 1
	auction.EndBlocker(ctx.WithBlockHeight(preExpiryHeight), keeper)

	// Check auction has not been closed yet
	_, found := keeper.GetAuction(ctx, auctionID)
	require.True(t, found)

	// Run the endblocker, simulating a block height just after auction expiry
	expiryHeight := preExpiryHeight + 1
	auction.EndBlocker(ctx.WithBlockHeight(expiryHeight), keeper)

	// Check auction has been closed
	_, found = keeper.GetAuction(ctx, auctionID)
	require.False(t, found)
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
