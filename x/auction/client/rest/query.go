package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/auction/client/common"
	"github.com/kava-labs/kava/x/auction/types"
)

const restAuctionID = "auction-id"

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/auctions", types.ModuleName), queryAuctionsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/auctions/{%s}", types.ModuleName, restAuctionID), queryAuctionHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/parameters", types.ModuleName), getParamsHandlerFn(cliCtx)).Methods("GET")
}

func queryAuctionHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		if len(vars[restAuctionID]) == 0 {
			err := fmt.Errorf("%s required but not specified", restAuctionID)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		auctionID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[restAuctionID])
		if !ok {
			return
		}
		auction, height, err := common.QueryAuctionByID(cliCtx, cliCtx.Codec, types.ModuleName, auctionID)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		// Decode and return results
		cliCtx = cliCtx.WithHeight(height)
		auctionWithPhase := types.NewAuctionWithPhase(auction)
		rest.PostProcessResponse(w, cliCtx, cliCtx.Codec.MustMarshalJSON(auctionWithPhase))
	}
}

func queryAuctionsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		var auctionType string
		var auctionDenom string
		var auctionPhase string

		if x := r.URL.Query().Get(RestType); len(x) != 0 {
			auctionType = strings.ToLower(strings.TrimSpace(x))
			if auctionType != types.CollateralAuctionType &&
				auctionType != types.SurplusAuctionType &&
				auctionType != types.DebtAuctionType {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid auction type %s", x))
				return
			}
		}

		if x := r.URL.Query().Get(RestDenom); len(x) != 0 {
			auctionDenom = strings.TrimSpace(x)
			err := sdk.ValidateDenom(auctionDenom)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if x := r.URL.Query().Get(RestPhase); len(x) != 0 {
			auctionPhase = strings.ToLower(strings.TrimSpace(x))
			if auctionType != types.CollateralAuctionType && len(auctionType) > 0 {
				rest.WriteErrorResponse(w, http.StatusBadRequest, "cannot apply phase flag to non-collateral auction type")
				return
			}
			if auctionPhase != types.ForwardAuctionPhase && auctionPhase != types.ReverseAuctionPhase {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid auction phase %s", x))
				return
			}
		}

		params := types.NewQueryAllAuctionParams(page, limit, auctionType, auctionDenom, auctionPhase)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetAuctions)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)

		// Unmarshal to Auction and remarshal as AuctionWithPhase
		var auctions types.Auctions
		err = cliCtx.Codec.UnmarshalJSON(res, &auctions)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		auctionsWithPhase := []types.AuctionWithPhase{} // using empty slice so json returns [] instead of null when there's no auctions
		for _, a := range auctions {
			auctionsWithPhase = append(auctionsWithPhase, types.NewAuctionWithPhase(a))
		}
		rest.PostProcessResponse(w, cliCtx, cliCtx.Codec.MustMarshalJSON(auctionsWithPhase))
	}
}

func getParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		// Get the params
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetParams), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Decode and return results
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
