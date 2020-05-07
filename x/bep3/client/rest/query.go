package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/kava-labs/kava/x/bep3/types"
)

const restSwapID = "swap-id"
const restDenom = "denom"

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/swap/{%s}", types.ModuleName, restSwapID), queryAtomicSwapHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/swaps", types.ModuleName), queryAtomicSwapsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/supply/{%s}", types.ModuleName, restDenom), queryAssetSupplyHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/supplies", types.ModuleName), queryAssetSuppliesHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/parameters", types.ModuleName), queryParamsHandlerFn(cliCtx)).Methods("GET")

}

func queryAtomicSwapHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		if len(vars[restSwapID]) == 0 {
			err := fmt.Errorf("%s required but not specified", restSwapID)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		swapID, err := hex.DecodeString(vars[restSwapID])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		bz, err := cliCtx.Codec.MarshalJSON(types.QueryAtomicSwapByID{SwapID: swapID})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("/custom/%s/%s", types.ModuleName, types.QueryGetAtomicSwap), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Decode and return results
		cliCtx = cliCtx.WithHeight(height)

		var swap types.AtomicSwap
		err = cliCtx.Codec.UnmarshalJSON(res, &swap)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, cliCtx.Codec.MustMarshalJSON(swap))
	}
}

// HTTP request handler to query list of atomic swaps filtered by optional params
func queryAtomicSwapsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		var (
			involveAddr   sdk.AccAddress
			expiration    uint64
			swapStatus    types.SwapStatus
			swapDirection types.SwapDirection
		)

		if x := r.URL.Query().Get(RestInvolve); len(x) != 0 {
			involveAddr, err = sdk.AccAddressFromBech32(x)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if x := r.URL.Query().Get(RestExpiration); len(x) != 0 {
			expiration, err = strconv.ParseUint(x, 10, 64)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if x := r.URL.Query().Get(RestStatus); len(x) != 0 {
			swapStatus = types.NewSwapStatusFromString(x)
			if !swapStatus.IsValid() {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid swap status %s", swapStatus))
				return
			}
		}

		if x := r.URL.Query().Get(RestDirection); len(x) != 0 {
			swapDirection = types.NewSwapDirectionFromString(x)
			if !swapDirection.IsValid() {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid swap direction %s", swapDirection))
				return
			}
		}

		params := types.NewQueryAtomicSwaps(page, limit, involveAddr, expiration, swapStatus, swapDirection)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryGetAtomicSwaps)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryAssetSupplyHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		denom := []byte(vars[restDenom])
		params := types.NewQueryAssetSupply(denom)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("/custom/%s/%s", types.ModuleName, types.QueryGetAssetSupply), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Decode and return results
		cliCtx = cliCtx.WithHeight(height)

		var assetSupply types.AssetSupply
		err = cliCtx.Codec.UnmarshalJSON(res, &assetSupply)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, cliCtx.Codec.MustMarshalJSON(assetSupply))
	}
}

func queryAssetSuppliesHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetAssetSupplies)

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Decode and return results
		cliCtx = cliCtx.WithHeight(height)

		var supplies types.AssetSupplies
		err = cliCtx.Codec.UnmarshalJSON(res, &supplies)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		// using empty slice so json returns [] instead of null when there's no swaps
		sliceSupplies := types.AssetSupplies{}
		for _, s := range supplies {
			sliceSupplies = append(sliceSupplies, s)
		}
		rest.PostProcessResponse(w, cliCtx, cliCtx.Codec.MustMarshalJSON(sliceSupplies))
	}
}

func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryGetParams)

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
