package rest

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/auction/types"
)

type placeBidReq struct {
	BaseReq   rest.BaseReq `json:"base_req"`
	AuctionID string       `json:"auction_id"`
	Bidder    string       `json:"bidder"`
	Bid       string       `json:"bid"`
	Lot       string       `json:"lot"`
}

const (
	restAuctionID = "auction_id"
	restBidder    = "bidder"
	restBid       = "bid"
	restLot       = "lot"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		fmt.Sprintf("/auction/bid/{%s}/{%s}/{%s}", restAuctionID, restBidder, restBid), bidHandlerFn(cliCtx)).Methods("PUT")
}

func bidHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req placeBidReq
		vars := mux.Vars(r)
		strAuctionID := vars[restAuctionID]
		bechBidder := vars[restBidder]
		strBid := vars[restBid]

		auctionID, err := strconv.ParseUint(strAuctionID, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		bidder, err := sdk.AccAddressFromBech32(bechBidder)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		bid, err := sdk.ParseCoin(strBid)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgPlaceBid(auctionID, bidder, bid)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if !bytes.Equal(fromAddr, bidder) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must bid from own address")
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})

	}
}
