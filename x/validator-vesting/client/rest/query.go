package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/kava-labs/kava/client/rest"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

func registerQueryRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/circulatingsupply", types.QueryPath), getCirculatingSupplyHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/totalsupply", types.QueryPath), getTotalSupplyHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/circulatingsupplyhard", types.QueryPath), getCirculatingSupplyHARDHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/circulatingsupplyusdx", types.QueryPath), getCirculatingSupplyUSDXHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/circulatingsupplyswp", types.QueryPath), getCirculatingSupplySWPHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/totalsupplyhard", types.QueryPath), getTotalSupplyHARDHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/totalsupplyusdx", types.QueryPath), getTotalSupplyUSDXHandlerFn(cliCtx)).Methods("GET")
}

func getTotalSupplyHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(cliCtx)
		res, err := queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resBytes, err := json.Marshal(res.Amount.Int64())
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(resBytes)
	}
}

func getCirculatingSupplyHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(cliCtx)
		res, err := queryClient.CirculatingSupply(context.Background(), &types.QueryCirculatingSupplyRequest{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resBytes, err := json.Marshal(res.Amount.Int64())
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(resBytes)
	}
}

func getCirculatingSupplyHARDHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(cliCtx)
		res, err := queryClient.CirculatingSupplyHARD(context.Background(), &types.QueryCirculatingSupplyHARDRequest{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resBytes, err := json.Marshal(res.Amount.Int64())
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(resBytes)
	}
}

func getCirculatingSupplyUSDXHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(cliCtx)
		res, err := queryClient.CirculatingSupplyUSDX(context.Background(), &types.QueryCirculatingSupplyUSDXRequest{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resBytes, err := json.Marshal(res.Amount.Int64())
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(resBytes)
	}
}

func getCirculatingSupplySWPHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(cliCtx)
		res, err := queryClient.CirculatingSupplySWP(context.Background(), &types.QueryCirculatingSupplySWPRequest{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resBytes, err := json.Marshal(res.Amount.Int64())
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(resBytes)
	}
}

func getTotalSupplyHARDHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(cliCtx)
		res, err := queryClient.TotalSupplyHARD(context.Background(), &types.QueryTotalSupplyHARDRequest{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resBytes, err := json.Marshal(res.Amount.Int64())
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(resBytes)
	}
}

func getTotalSupplyUSDXHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(cliCtx)
		res, err := queryClient.TotalSupplyUSDX(context.Background(), &types.QueryTotalSupplyUSDXRequest{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resBytes, err := json.Marshal(res.Amount.Int64())
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(resBytes)
	}
}
