package rest_v0_3

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	v18de63auth "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
	valvesting "github.com/kava-labs/kava/x/validator-vesting"
	v0_3valvesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_3"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	s := r.PathPrefix("/v0_3").Subrouter()

	s.HandleFunc("/node_info", rpc.NodeInfoRequestHandlerFn(cliCtx)).Methods("GET")
	s.HandleFunc(
		"/auth/accounts/{address}", QueryAccountRequestHandlerFn(cliCtx),
	).Methods("GET")
	s.HandleFunc("/txs/{hash}", authrest.QueryTxRequestHandlerFn(cliCtx)).Methods("GET")
	// r.HandleFunc("/txs", QueryTxsRequestHandlerFn(cliCtx)).Methods("GET") // assume they don't need GET here
	s.HandleFunc("/txs", authrest.BroadcastTxRequest(cliCtx)).Methods("POST")

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

		account, height, err := accGetter.GetAccountWithHeight(cliCtx, addr)

		// convert v0.8 account type into old v0.3 account type so that it json marshals into the v0.3 format
		oldAccount := rollbackAccountType(account)
		// use old codec with old account interface registered
		cliCtx = cliCtx.WithCodec(makeCodecV03())

		if err != nil {
			if err := accGetter.EnsureExists(cliCtx, addr); err != nil {
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

func makeCodecV03() *codec.Codec {
	v0_3Codec := codec.New()
	codec.RegisterCrypto(v0_3Codec)
	v18de63auth.RegisterCodec(v0_3Codec)
	v18de63auth.RegisterCodecVesting(v0_3Codec)
	v18de63supply.RegisterCodec(v0_3Codec)
	v0_3valvesting.RegisterCodec(v0_3Codec)
	return v0_3Codec
}
func rollbackAccountType(newAccount authtypes.AccountI) v18de63auth.Account {
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

	case authtypes.ModuleAccount:
		ba := v18de63auth.BaseAccount(*(acc.BaseAccount))
		ma := authtypes.ModuleAccount{
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
