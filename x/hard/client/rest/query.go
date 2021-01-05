package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/hard/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/parameters", types.ModuleName), queryParamsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/deposits", types.ModuleName), queryDepositsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/accounts", types.ModuleName), queryModAccountsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/borrows", types.ModuleName), queryBorrowsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/borrow", types.ModuleName), queryBorrowHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/borrowed", types.ModuleName), queryBorrowedHandlerFn(cliCtx)).Methods("GET")
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

func queryDepositsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		var depositDenom string
		var depositOwner sdk.AccAddress

		if x := r.URL.Query().Get(RestDenom); len(x) != 0 {
			depositDenom = strings.TrimSpace(x)
		}

		if x := r.URL.Query().Get(RestOwner); len(x) != 0 {
			depositOwnerStr := strings.ToLower(strings.TrimSpace(x))
			depositOwner, err = sdk.AccAddressFromBech32(depositOwnerStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("cannot parse address from deposit owner %s", depositOwnerStr))
				return
			}
		}

		params := types.NewQueryDepositParams(page, limit, depositDenom, depositOwner)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetDeposits)
		res, height, err := cliCtx.QueryWithData(route, bz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryBorrowsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		var borrowDenom string
		var borrowOwner sdk.AccAddress

		if x := r.URL.Query().Get(RestBorrowDenom); len(x) != 0 {
			borrowDenom = strings.TrimSpace(x)
		}

		if x := r.URL.Query().Get(RestOwner); len(x) != 0 {
			borrowOwnerStr := strings.ToLower(strings.TrimSpace(x))
			borrowOwner, err = sdk.AccAddressFromBech32(borrowOwnerStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("cannot parse address from borrow owner %s", borrowOwnerStr))
				return
			}
		}

		params := types.NewQueryBorrowsParams(page, limit, borrowOwner, borrowDenom)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetBorrows)
		res, height, err := cliCtx.QueryWithData(route, bz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryBorrowHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, _, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		var borrowOwner sdk.AccAddress

		if x := r.URL.Query().Get(RestOwner); len(x) != 0 {
			borrowOwnerStr := strings.ToLower(strings.TrimSpace(x))
			borrowOwner, err = sdk.AccAddressFromBech32(borrowOwnerStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("cannot parse address from borrow owner %s", borrowOwnerStr))
				return
			}
		}

		params := types.NewQueryBorrowParams(borrowOwner)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetBorrow)
		res, height, err := cliCtx.QueryWithData(route, bz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryBorrowedHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, _, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		var borrowDenom string

		if x := r.URL.Query().Get(RestBorrowDenom); len(x) != 0 {
			borrowDenom = strings.TrimSpace(x)
		}

		params := types.NewQueryBorrowedParams(borrowDenom)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetBorrowed)
		res, height, err := cliCtx.QueryWithData(route, bz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryModAccountsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		var name string

		if x := r.URL.Query().Get(RestName); len(x) != 0 {
			name = strings.TrimSpace(x)
		}

		params := types.NewQueryAccountParams(page, limit, name)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetModuleAccounts)
		res, height, err := cliCtx.QueryWithData(route, bz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
