package rest_v0_3

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/supply"

	v18de63auth "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
	v18de63supply "github.com/kava-labs/kava/migrate/v0_8/sdk/supply/v18de63"
	v18de63sdk "github.com/kava-labs/kava/migrate/v0_8/sdk/types/v18de63"
	valvesting "github.com/kava-labs/kava/x/validator-vesting"
	v0_3valvesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_3"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	s := r.PathPrefix("/v0_3").Subrouter()

	// node_info schema has not changed between cosmos v18de63 and v0.38.4
	s.HandleFunc("/node_info", rpc.NodeInfoRequestHandlerFn(cliCtx)).Methods("GET")
	s.HandleFunc("/auth/accounts/{address}", QueryAccountRequestHandlerFn(cliCtx)).Methods("GET")

	s.HandleFunc("/txs/{hash}", QueryTxRequestHandlerFn(cliCtx)).Methods("GET")
	// r.HandleFunc("/txs", QueryTxsRequestHandlerFn(cliCtx)).Methods("GET") // assume they don't need GET here
	s.HandleFunc("/txs", authrest.BroadcastTxRequest(cliCtx)).Methods("POST")

	r.HandleFunc("/blocks/latest", LatestBlockRequestHandlerFn(cliCtx)).Methods("GET")

	// TODO these are unchanged between cosmos v18de63 and v0.38.4, but can't import private methods. Maybe redirect?
	// Get all delegations from a delegator
	s.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations",
		delegatorDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all unbonding delegations from a delegator
	s.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations",
		delegatorUnbondingDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get the total rewards balance from all delegations
	s.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		delegatorRewardsHandlerFn(cliCtx, disttypes.ModuleName),
	).Methods("GET")

}

// REST handler to get the latest block
func LatestBlockRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		output, err := getBlock(cliCtx, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, output)
	}
}

func getBlock(cliCtx context.CLIContext, height *int64) ([]byte, error) {
	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	// header -> BlockchainInfo
	// header, tx -> Block
	// results -> BlockResults
	res, err := node.Block(height)
	if err != nil {
		return nil, err
	}

	// TODO convert block

	// if !cliCtx.TrustNode {
	// 	check, err := cliCtx.Verify(res.Block.Height)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if err := tmliteProxy.ValidateHeader(&res.Block.Header, check); err != nil {
	// 		return nil, err
	// 	}

	// 	if err = tmliteProxy.ValidateBlock(res.Block, check); err != nil {
	// 		return nil, err
	// 	}
	// }

	// if cliCtx.Indent {
	// 	return codec.Cdc.MarshalJSONIndent(res, "", "  ")
	// }

	return codec.Cdc.MarshalJSON(res)
}

// QueryAccountRequestHandlerFn handle auth/accounts queries
// This function is identical to v0.8 except the queried account is cast to the v0.3 account type so it marshals in the old format.
func QueryAccountRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32addr := vars["address"]

		addr, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		accGetter := types.NewAccountRetriever(cliCtx)

		account, height, err := accGetter.GetAccountWithHeight(addr)

		// convert v0.8 account type into old v0.3 account type so that it json marshals into the v0.3 format
		oldAccount := rollbackAccountType(account)
		// use old codec with old account interface registered
		cliCtx = cliCtx.WithCodec(makeCodecV03())

		if err != nil {
			if err := accGetter.EnsureExists(addr); err != nil {
				cliCtx = cliCtx.WithHeight(height)
				rest.PostProcessResponse(w, cliCtx, v18de63auth.BaseAccount{})
				return
			}

			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, oldAccount)
	}
}

// QueryTxRequestHandlerFn implements a REST handler that queries a transaction
// by hash in a committed block.
func QueryTxRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hashHexStr := vars["hash"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		output, err := utils.QueryTx(cliCtx, hashHexStr)
		// convert v0.8 TxResponse to a v0.3 Tx Response
		oldOutput := rollbackTxResponseType(output)
		if err != nil {
			if strings.Contains(err.Error(), hashHexStr) {
				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if oldOutput.Empty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("no transaction found with hash %s", hashHexStr))
		}

		rest.PostProcessResponseBare(w, cliCtx, oldOutput)
	}
}

func rollbackResponseType(response sdk.TxResponse) {
	return v18de63sdk.TxResponse{
		Height:    response.Height,
		TxHash:    response.TxHash,
		Codespace: response.Codespace,
		Code:      response.Code,
		Data:      response.Data,
		RawLog:    response.RawLog,
		Logs:      response.Logs, // TODO need to convert type and add back Success field
		Info:      response.Info,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
		Tx:        response.Tx,
		Timestamp: response.Timestamp,
		Events:    response.Logs[0].Events, // TODO concat events?
	}

}

func makeCodecV03() *codec.Codec {
	v0_3Codec := codec.New()
	codec.RegisterCrypto(v0_3Codec)
	v18de63auth.RegisterCodec(v0_3Codec)
	v18de63auth.RegisterCodecVesting(v0_3Codec)
	v18de63supply.RegisterCodec(v0_3Codec)
	v0_3valvesting.RegisterCodec(v0_3Codec)
	return v0_3Codec
}
func rollbackAccountType(newAccount authexported.Account) v18de63auth.Account {
	switch acc := newAccount.(type) {

	case *auth.BaseAccount:
		return v18de63auth.BaseAccount(*acc)

	case *vestingtypes.PeriodicVestingAccount:
		ba := v18de63auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
		bva := v18de63auth.BaseVestingAccount{
			BaseAccount:      &ba,
			OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
			DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
			DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
			EndTime:          acc.BaseVestingAccount.EndTime,
		}
		var newPeriods v18de63auth.Periods
		for _, p := range acc.VestingPeriods {
			newPeriods = append(newPeriods, v18de63auth.Period(p))
		}
		pva := v18de63auth.PeriodicVestingAccount{
			BaseVestingAccount: &bva,
			StartTime:          acc.StartTime,
			VestingPeriods:     newPeriods,
		}
		return pva

	case *valvesting.ValidatorVestingAccount:
		ba := v18de63auth.BaseAccount(*(acc.PeriodicVestingAccount.BaseVestingAccount.BaseAccount))
		bva := v18de63auth.BaseVestingAccount{
			BaseAccount:      &ba,
			OriginalVesting:  acc.PeriodicVestingAccount.BaseVestingAccount.OriginalVesting,
			DelegatedFree:    acc.PeriodicVestingAccount.BaseVestingAccount.DelegatedFree,
			DelegatedVesting: acc.PeriodicVestingAccount.BaseVestingAccount.DelegatedVesting,
			EndTime:          acc.PeriodicVestingAccount.BaseVestingAccount.EndTime,
		}
		var newPeriods v18de63auth.Periods
		for _, p := range acc.PeriodicVestingAccount.VestingPeriods {
			newPeriods = append(newPeriods, v18de63auth.Period(p))
		}
		pva := v18de63auth.PeriodicVestingAccount{
			BaseVestingAccount: &bva,
			StartTime:          acc.PeriodicVestingAccount.StartTime,
			VestingPeriods:     newPeriods,
		}
		var newVestingProgress []v0_3valvesting.VestingProgress
		for _, p := range acc.VestingPeriodProgress {
			newVestingProgress = append(newVestingProgress, v0_3valvesting.VestingProgress(p))
		}
		vva := v0_3valvesting.ValidatorVestingAccount{
			PeriodicVestingAccount: &pva,
			ValidatorAddress:       acc.ValidatorAddress,
			ReturnAddress:          acc.ReturnAddress,
			SigningThreshold:       acc.SigningThreshold,
			CurrentPeriodProgress:  v0_3valvesting.CurrentPeriodProgress(acc.CurrentPeriodProgress),
			VestingPeriodProgress:  newVestingProgress,
			DebtAfterFailedVesting: acc.DebtAfterFailedVesting,
		}
		return vva

	case supply.ModuleAccount:
		ba := v18de63auth.BaseAccount(*(acc.BaseAccount))
		ma := v18de63supply.ModuleAccount{
			BaseAccount: &ba,
			Name:        acc.Name,
			Permissions: acc.Permissions,
		}
		return ma

	case nil:
		return acc

	default:
		panic(fmt.Errorf("unrecognized account type %+v", acc))
	}
}

// staking handler funcs

// HTTP request handler to query a delegator delegations
func delegatorDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, fmt.Sprintf("custom/%s/%s", stakingtypes.QuerierRoute, stakingtypes.QueryDelegatorDelegations))
}

// HTTP request handler to query a delegator unbonding delegations
func delegatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, "custom/staking/delegatorUnbondingDelegations")
}

func queryDelegator(cliCtx context.CLIContext, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := stakingtypes.NewQueryDelegatorParams(delegatorAddr)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(endpoint, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// distribution handler funcs

// HTTP request handler to query the total rewards balance from all delegations
func delegatorRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		delegatorAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		params := disttypes.NewQueryDelegatorParams(delegatorAddr)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", queryRoute, disttypes.QueryDelegatorTotalRewards)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
func checkDelegatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.AccAddress, bool) {
	addr, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	return addr, true
}
