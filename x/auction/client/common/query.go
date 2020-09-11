package common

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/auction/types"
)

const (
	defaultPage  = 1
	defaultLimit = 30
)

// QueryAuctionByID returns an auction from state if present or falls back to searching old blocks
func QueryAuctionByID(cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string, auctionID uint64) (types.Auction, int64, error) {
	bz, err := cdc.MarshalJSON(types.NewQueryAuctionParams(auctionID))
	if err != nil {
		return nil, 0, err
	}

	res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAuction), bz)

	if err == nil {
		var auction types.Auction
		cdc.MustUnmarshalJSON(res, &auction)

		return auction, height, nil
	}

	// NOTE: !errors.Is(err, types.ErrUnknownProposal) does not work here
	if err != nil && !strings.Contains(err.Error(), "auction not found") {
		return nil, 0, err
	}

	res, height, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryNextAuctionID), nil)
	if err != nil {
		return nil, 0, err
	}

	var nextAuctionID uint64
	cdc.MustUnmarshalJSON(res, &nextAuctionID)

	if auctionID >= nextAuctionID {
		return nil, 0, sdkerrors.Wrapf(types.ErrAuctionNotFound, "%d", auctionID)
	}

	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, "place_bid"),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeAuctionBid, types.AttributeKeyAuctionID, []byte(fmt.Sprintf("%d", auctionID))),
	}

	searchResult, err := utils.QueryTxsByEvents(cliCtx, events, defaultPage, defaultLimit)
	if err != nil {
		return nil, 0, err
	}

	maxHeight := int64(0)
	found := false

	for _, info := range searchResult.Txs {
		for _, msg := range info.Tx.GetMsgs() {
			if msg.Type() == "place_bid" {
				found = true
				if info.Height > maxHeight {
					maxHeight = info.Height
				}
			}
		}
	}

	if !found {
		return nil, 0, sdkerrors.Wrapf(types.ErrAuctionNotFound, "%d", auctionID)
	}

	queryCLIContext := cliCtx.WithHeight(maxHeight)
	res, height, err = queryCLIContext.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAuction), bz)
	if err != nil {
		return nil, 0, err
	}

	// Decode and print results
	var auction types.Auction
	cdc.MustUnmarshalJSON(res, &auction)
	return auction, height, nil
}
