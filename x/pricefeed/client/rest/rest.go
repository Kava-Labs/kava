package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"
	"github.com/kava-labs/kava/x/pricefeed/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	restName = "assetCode"
)

type postPriceReq struct {
	BaseReq   rest.BaseReq `json:"base_req"`
	AssetCode string       `json:"asset_code"`
	Price     string       `json:"price"`
	Expiry    string       `json:"expiry"`
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/rawprices", storeName), postPriceHandler(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/rawprices/{%s}", storeName, restName), getRawPricesHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currentprice/{%s}", storeName, restName), getCurrentPriceHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/assets", storeName), getAssetsHandler(cliCtx, storeName)).Methods("GET")
}

func postPriceHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req postPriceReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(baseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		price, err := sdk.NewDecFromStr(req.Price)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		expiryInt, ok := sdk.NewIntFromString(req.Expiry)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "invalid expiry")
			return
		}
		expiry := tmtime.Canonical(time.Unix(expiryInt.Int64(), 0))

		// create the message
		msg := types.NewMsgPostPrice(addr, req.AssetCode, price, expiry)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

func getRawPricesHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restName]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/rawprices/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getCurrentPriceHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restName]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/price/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getAssetsHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/assets/", storeName), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
