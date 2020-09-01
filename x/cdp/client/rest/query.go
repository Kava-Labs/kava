package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	r.HandleFunc("/cdp/savingsratedist", getSavingsRateDistributedHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/cdp/{%s}/{%s}", types.RestOwner, types.RestCollateralType), queryCdpHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/collateralType/{%s}", types.RestCollateralType), queryCdpsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/ratio/{%s}/{%s}", types.RestCollateralType, types.RestRatio), queryCdpsByRatioHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/cdps/cdp/deposits/{%s}/{%s}", types.RestOwner, types.RestCollateralType), queryCdpDepositsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/v2cdps"), queryV2CdpsHandlerFn(cliCtx)).Methods("GET")
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

func getSavingsRateDistributedHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetSavingsRateDistributed), nil)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryV2CdpsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		var cdpCollateralDenom string
		var cdpOwner sdk.AccAddress
		var cdpID uint64

		if x := r.URL.Query().Get(RestCollateralDenom); len(x) != 0 {
			cdpCollateralDenom = strings.TrimSpace(x)
			err := sdk.ValidateDenom(cdpCollateralDenom)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if x := r.URL.Query().Get(RestOwner); len(x) != 0 {
			cdpOwnerStr := strings.ToLower(strings.TrimSpace(x))
			cdpOwner, err = sdk.AccAddressFromBech32(cdpOwnerStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("cannot parse address from cdp owner %s", cdpOwnerStr))
			}
		}

		if x := r.URL.Query().Get(RestID); len(x) != 0 {
			cdpID, err = strconv.ParseUint(strings.TrimSpace(x), 10, 64)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		params := types.NewQueryV2CdpsParams(page, limit, cdpCollateralDenom, cdpOwner, cdpID)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetV2Cdps)
		res, height, err := cliCtx.QueryWithData(route, bz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
