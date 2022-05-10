package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestGrpcAuctionsFilter(t *testing.T) {
	// setup
	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates()
	auctionsKeeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1})
	_, addrs := app.GeneratePrivKeyAddressPairs(2)

	auctions := []types.Auction{
		types.NewSurplusAuction(
			"sellerMod",
			c("swp", 12345678),
			"usdx",
			time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
		).WithID(0),
		types.NewDebtAuction(
			"buyerMod",
			c("hard", 12345678),
			c("usdx", 12345678),
			time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
			c("debt", 12345678),
		).WithID(1),
		types.NewCollateralAuction(
			"sellerMod",
			c("ukava", 12345678),
			time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
			c("usdx", 12345678),
			types.WeightedAddresses{
				Addresses: addrs,
				Weights:   []sdk.Int{sdk.NewInt(100)},
			},
			c("debt", 12345678),
		).WithID(2),
		types.NewCollateralAuction(
			"sellerMod",
			c("hard", 12345678),
			time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
			c("usdx", 12345678),
			types.WeightedAddresses{
				Addresses: addrs,
				Weights:   []sdk.Int{sdk.NewInt(100)},
			},
			c("debt", 12345678),
		).WithID(3),
	}
	for _, a := range auctions {
		auctionsKeeper.SetAuction(ctx, a)
	}

	qs := keeper.NewQueryServerImpl(auctionsKeeper)

	tests := []struct {
		giveName     string
		giveRequest  types.QueryAuctionsRequest
		wantResponse []types.Auction
	}{
		{
			"empty request",
			types.QueryAuctionsRequest{},
			auctions,
		},
		{
			"denom query swp",
			types.QueryAuctionsRequest{
				Denom: "swp",
			},
			auctions[0:1],
		},
		{
			"denom query usdx all",
			types.QueryAuctionsRequest{
				Denom: "usdx",
			},
			auctions,
		},
		{
			"owner",
			types.QueryAuctionsRequest{
				Owner: addrs[0].String(),
			},
			auctions[2:4],
		},
		{
			"owner and denom",
			types.QueryAuctionsRequest{
				Owner: addrs[0].String(),
				Denom: "hard",
			},
			auctions[3:4],
		},
		{
			"owner, denom, type, phase",
			types.QueryAuctionsRequest{
				Owner: addrs[0].String(),
				Denom: "hard",
				Type:  types.CollateralAuctionType,
				Phase: types.ForwardAuctionPhase,
			},
			auctions[3:4],
		},
	}

	for _, tc := range tests {
		t.Run(tc.giveName, func(t *testing.T) {
			res, err := qs.Auctions(sdk.WrapSDKContext(ctx), &tc.giveRequest)
			require.NoError(t, err)

			var unpackedAuctions []types.Auction

			for _, anyAuction := range res.Auctions {
				var auction types.Auction
				err := tApp.AppCodec().UnpackAny(anyAuction, &auction)
				require.NoError(t, err)

				unpackedAuctions = append(unpackedAuctions, auction)
			}

			require.Equal(t, tc.wantResponse, unpackedAuctions)
		})
	}
}
