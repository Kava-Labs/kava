package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/incentive/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/rewards", types.ModuleName), queryRewardsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/parameters", types.ModuleName), queryParamsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/reward-factors", types.ModuleName), queryRewardFactorsHandlerFn(cliCtx)).Methods("GET")
}

func queryRewardsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		var owner sdk.AccAddress
		if x := r.URL.Query().Get(types.RestClaimOwner); len(x) != 0 {
			ownerStr := strings.ToLower(strings.TrimSpace(x))
			owner, err = sdk.AccAddressFromBech32(ownerStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("cannot parse address from claim owner %s", ownerStr))
				return
			}
		}

		var rewardType string
		if x := r.URL.Query().Get(types.RestClaimType); len(x) != 0 {
			rewardType = strings.ToLower(strings.TrimSpace(x))
		}

		var unsynced bool
		if x := r.URL.Query().Get(types.RestUnsynced); len(x) != 0 {
			unsyncedStr := strings.ToLower(strings.TrimSpace(x))
			unsynced, err = strconv.ParseBool(unsyncedStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("cannot parse bool from unsynced flag %s", unsyncedStr))
				return
			}
		}

		params := types.NewQueryRewardsParams(page, limit, owner, unsynced)
		switch strings.ToLower(rewardType) {
		case "hard":
			executeHardRewardsQuery(w, cliCtx, params)
		case "usdx_minting":
			executeUSDXMintingRewardsQuery(w, cliCtx, params)
		case "delegator":
			executeDelegatorRewardsQuery(w, cliCtx, params)
		default:
			executeAllRewardQueries(w, cliCtx, params)
		}
	}
}

func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/parameters", types.QuerierRoute)

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryRewardFactorsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		var denom string

		if x := r.URL.Query().Get(RestDenom); len(x) != 0 {
			denom = strings.TrimSpace(x)
		}

		params := types.NewQueryRewardFactorsParams(denom)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetRewardFactors)
		res, height, err := cliCtx.QueryWithData(route, bz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func executeHardRewardsQuery(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryRewardsParams) {
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetHardRewards), bz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	cliCtx = cliCtx.WithHeight(height)
	rest.PostProcessResponse(w, cliCtx, res)
}

func executeUSDXMintingRewardsQuery(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryRewardsParams) {
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetUSDXMintingRewards), bz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	cliCtx = cliCtx.WithHeight(height)
	rest.PostProcessResponse(w, cliCtx, res)
}

func executeDelegatorRewardsQuery(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryRewardsParams) {
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetDelegatorRewards), bz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	cliCtx = cliCtx.WithHeight(height)
	rest.PostProcessResponse(w, cliCtx, res)
}

func executeAllRewardQueries(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryRewardsParams) {

	paramsBz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}
	hardRes, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetHardRewards), paramsBz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var hardClaims types.HardLiquidityProviderClaims
	cliCtx.Codec.MustUnmarshalJSON(hardRes, &hardClaims)

	usdxMintingRes, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetUSDXMintingRewards), paramsBz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var usdxMintingClaims types.USDXMintingClaims
	cliCtx.Codec.MustUnmarshalJSON(usdxMintingRes, &usdxMintingClaims)

	delegatorRes, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetDelegatorRewards), paramsBz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var delegatorClaims types.DelegatorClaims
	cliCtx.Codec.MustUnmarshalJSON(delegatorRes, &delegatorClaims)

	cliCtx = cliCtx.WithHeight(height)

	type rewardResult struct {
		HardClaims        types.HardLiquidityProviderClaims `json:"hard_claims" yaml:"hard_claims"`
		UsdxMintingClaims types.USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"`
		DelegatorClaims   types.DelegatorClaims             `json:"delegator_claims" yaml:"delegator_claims"`
	}

	res := rewardResult{
		HardClaims:        hardClaims,
		UsdxMintingClaims: usdxMintingClaims,
		DelegatorClaims:   delegatorClaims,
	}

	resBz, err := cliCtx.Codec.MarshalJSON(res)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal result: %s", err))
		return
	}

	rest.PostProcessResponse(w, cliCtx, resBz)
}
