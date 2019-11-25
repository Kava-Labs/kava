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

/*
API Design:

Currently CDPs do not have IDs so standard REST uri conventions (ie GET /cdps/{cdp-id}) don't work too well.

Get one or more cdps
	GET /cdps?collateralDenom={denom}&owner={address}&underCollateralizedAt={price}
Modify a CDP (idempotent). Create is not separated out because conceptually all CDPs already exist (just with zero collateral and debt). // TODO is making this idempotent actually useful?
	PUT /cdps
Get the module params, including authorized collateral denoms.
	GET /params
*/

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/cdps", getCdpsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/cdps/params", getParamsHandlerFn(cliCtx)).Methods("GET")
}

func getCdpsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get parameters from the URL
		ownerBech32 := r.URL.Query().Get(types.RestOwner)
		collateralDenom := r.URL.Query().Get(types.RestCollateralDenom)
		priceString := r.URL.Query().Get(types.RestUnderCollateralizedAt)

		// Construct querier params
		querierParams := types.QueryCdpsParams{}

		if len(ownerBech32) != 0 {
			owner, err := sdk.AccAddressFromBech32(ownerBech32)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			querierParams.Owner = owner
		}

		if len(collateralDenom) != 0 {
			// TODO validate denom
			querierParams.CollateralDenom = collateralDenom
		}

		if len(priceString) != 0 {
			price, err := sdk.NewDecFromStr(priceString)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			querierParams.UnderCollateralizedAt = price
		}

		querierParamsBz, err := cliCtx.Codec.MarshalJSON(querierParams)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Get the CDPs
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetCdps), querierParamsBz)
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
