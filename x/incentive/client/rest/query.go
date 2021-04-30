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

		var denom string
		if x := r.URL.Query().Get(types.RestClaimDenom); len(x) != 0 {
			denom = strings.ToLower(strings.TrimSpace(x))
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

		if unsynced {
			switch strings.ToLower(rewardType) {
			case "hard":
				params := types.NewQueryHardRewardsUnsyncedParams(page, limit, owner)
				executeHardRewardsUnsyncedQuery(w, cliCtx, params)
			case "usdx_minting":
				params := types.NewQueryUSDXMintingRewardsUnsyncedParams(page, limit, owner)
				executeUSDXMintingRewardsUnsyncedQuery(w, cliCtx, params)
			default:
				hardParams := types.NewQueryHardRewardsUnsyncedParams(page, limit, owner)
				usdxMintingParams := types.NewQueryUSDXMintingRewardsUnsyncedParams(page, limit, owner)
				executeBothUnsyncedRewardQueries(w, cliCtx, hardParams, usdxMintingParams)
			}
		} else {
			switch strings.ToLower(rewardType) {
			case "hard":
				params := types.NewQueryHardRewardsParams(page, limit, owner, denom)
				executeHardRewardsQuery(w, cliCtx, params)
			case "usdx_minting":
				params := types.NewQueryUSDXMintingRewardsParams(page, limit, owner)
				executeUSDXMintingRewardsQuery(w, cliCtx, params)
			default:
				hardParams := types.NewQueryHardRewardsParams(page, limit, owner, denom)
				usdxMintingParams := types.NewQueryUSDXMintingRewardsParams(page, limit, owner)
				executeBothRewardQueries(w, cliCtx, hardParams, usdxMintingParams)
			}
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

func executeHardRewardsQuery(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryHardRewardsParams) {
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

func executeHardRewardsUnsyncedQuery(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryHardRewardsUnsyncedParams) {
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetHardRewardsUnsynced), bz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	cliCtx = cliCtx.WithHeight(height)
	rest.PostProcessResponse(w, cliCtx, res)
}

func executeUSDXMintingRewardsQuery(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryUSDXMintingRewardsParams) {
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

func executeUSDXMintingRewardsUnsyncedQuery(w http.ResponseWriter, cliCtx context.CLIContext, params types.QueryUSDXMintingRewardsUnsyncedParams) {
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetUSDXMintingRewardsUnsynced), bz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	cliCtx = cliCtx.WithHeight(height)
	rest.PostProcessResponse(w, cliCtx, res)
}

func executeBothRewardQueries(w http.ResponseWriter, cliCtx context.CLIContext,
	hardParams types.QueryHardRewardsParams, usdxMintingParams types.QueryUSDXMintingRewardsParams) {
	hardBz, err := cliCtx.Codec.MarshalJSON(hardParams)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	hardRes, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetHardRewards), hardBz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var hardClaims types.HardLiquidityProviderClaims
	cliCtx.Codec.MustUnmarshalJSON(hardRes, &hardClaims)

	usdxMintingBz, err := cliCtx.Codec.MarshalJSON(usdxMintingParams)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	usdxMintingRes, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetUSDXMintingRewards), usdxMintingBz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var usdxMintingClaims types.USDXMintingClaims
	cliCtx.Codec.MustUnmarshalJSON(usdxMintingRes, &usdxMintingClaims)

	cliCtx = cliCtx.WithHeight(height)

	type rewardResult struct {
		HardClaims        types.HardLiquidityProviderClaims `json:"hard_claims" yaml:"hard_claims"`
		UsdxMintingClaims types.USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"`
	}

	res := rewardResult{
		HardClaims:        hardClaims,
		UsdxMintingClaims: usdxMintingClaims,
	}

	resBz, err := cliCtx.Codec.MarshalJSON(res)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal result: %s", err))
		return
	}

	rest.PostProcessResponse(w, cliCtx, resBz)
}

func executeBothUnsyncedRewardQueries(w http.ResponseWriter, cliCtx context.CLIContext,
	hardParams types.QueryHardRewardsUnsyncedParams, usdxMintingParams types.QueryUSDXMintingRewardsUnsyncedParams) {
	hardBz, err := cliCtx.Codec.MarshalJSON(hardParams)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	hardRes, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetHardRewardsUnsynced), hardBz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var hardClaims types.HardLiquidityProviderClaims
	cliCtx.Codec.MustUnmarshalJSON(hardRes, &hardClaims)

	usdxMintingBz, err := cliCtx.Codec.MarshalJSON(usdxMintingParams)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
		return
	}

	usdxMintingRes, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/incentive/%s", types.QueryGetUSDXMintingRewardsUnsynced), usdxMintingBz)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var usdxMintingClaims types.USDXMintingClaims
	cliCtx.Codec.MustUnmarshalJSON(usdxMintingRes, &usdxMintingClaims)

	cliCtx = cliCtx.WithHeight(height)

	type rewardResult struct {
		HardClaims        types.HardLiquidityProviderClaims `json:"hard_claims" yaml:"hard_claims"`
		UsdxMintingClaims types.USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"`
	}

	res := rewardResult{
		HardClaims:        hardClaims,
		UsdxMintingClaims: usdxMintingClaims,
	}

	resBz, err := cliCtx.Codec.MarshalJSON(res)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal result: %s", err))
		return
	}

	rest.PostProcessResponse(w, cliCtx, resBz)
}
