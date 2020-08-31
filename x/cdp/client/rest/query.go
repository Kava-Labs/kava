package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/kava-labs/kava/x/cdp/types"
)

// define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/cdp/accounts", getAccountsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/cdp/parameters", getParamsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/cdp/{%s}/{%s}", types.RestOwner, types.RestCollateralType), queryCdpHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/collateralType/{%s}", types.RestCollateralType), queryCdpsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/ratio/{%s}/{%s}", types.RestCollateralType, types.RestRatio), queryCdpsByRatioHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/cdp/deposits/{%s}/{%s}", types.RestOwner, types.RestCollateralType), queryCdpDepositsHandlerFn(cliCtx)).Methods("GET")
}

func queryCdpHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		ownerBech32 := vars[types.RestOwner]
		collateralType := vars[types.RestCollateralType]

		owner, err := sdk.AccAddressFromBech32(ownerBech32)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.NewQueryCdpParams(owner, collateralType)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
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
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		collateralType := vars[types.RestCollateralType]

		params := types.NewQueryCdpsParams(collateralType)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
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
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		collateralType := vars[types.RestCollateralType]
		ratioStr := vars[types.RestRatio]

		ratioDec, sdkError := sdk.NewDecFromStr(ratioStr)
		if sdkError != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkError.Error())
			return
		}

		params := types.NewQueryCdpsByRatioParams(collateralType, ratioDec)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetCdpsByCollateralization), bz)
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
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		ownerBech32 := vars[types.RestOwner]
		collateralType := vars[types.RestCollateralType]

		owner, err := sdk.AccAddressFromBech32(ownerBech32)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.NewQueryCdpDeposits(owner, collateralType)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
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
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetParams), nil)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getAccountsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetAccounts), nil)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
