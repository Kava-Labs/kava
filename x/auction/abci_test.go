package auction

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestKeeper_EndBlocker(t *testing.T) {
	// setup keeper and auction
	mapp, keeper, addresses, _ := setUpMockApp()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)

	seller := addresses[0]
	keeper.StartForwardAuction(ctx, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 0))

	// run the endblocker, simulating a block height after auction expiry
	expiryBlock := ctx.BlockHeight() + int64(DefaultMaxAuctionDuration)
	EndBlocker(ctx.WithBlockHeight(expiryBlock), keeper)

	// check auction has been closed
	_, found := keeper.GetAuction(ctx, 0)
	require.False(t, found)
}
