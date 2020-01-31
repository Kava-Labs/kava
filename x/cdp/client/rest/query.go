package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/cdp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/cdp/{%s}/{%s}", types.RestOwner, types.RestCollateralDenom), queryCdpHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/{%s}", types.RestCollateralDenom), queryCdpsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/{%s}/ratio/{%s}", types.RestCollateralDenom, types.RestRatio), queryCdpsByRatioHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/deposits/{%s}/{%s}", types.RestOwner, types.RestCollateralDenom), queryCdpDepositsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/cdp/params", getParamsHandlerFn(cliCtx)).Methods("GET")
}

func queryCdpHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ownerBech32 := vars[types.RestOwner]
		collateralDenom := vars[types.RestCollateralDenom]

		owner, err := sdk.AccAddressFromBech32(ownerBech32)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.NewQueryCdpParams(owner, collateralDenom)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetCdp), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)

	}
}

func queryCdpsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		collateralDenom := vars[types.RestCollateralDenom]

		params := types.NewQueryCdpsParams(collateralDenom)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetCdps), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)

	}
}

func queryCdpsByRatioHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		collateralDenom := vars[types.RestCollateralDenom]
		ratioStr := vars[types.RestRatio]

		ratioDec, sdkError := sdk.NewDecFromStr(ratioStr)
		if sdkError != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkError.Error())
			return
		}

		params := types.NewQueryCdpsByRatioParams(collateralDenom, ratioDec)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetCdps), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)

	}
}

func queryCdpDepositsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ownerBech32 := vars[types.RestOwner]
		collateralDenom := vars[types.RestCollateralDenom]

		owner, err := sdk.AccAddressFromBech32(ownerBech32)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.NewQueryCdpDeposits(owner, collateralDenom)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetCdpDeposits), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)

	}
}

func getParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the params
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetParams), nil)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Return the params
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
