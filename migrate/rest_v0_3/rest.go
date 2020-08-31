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
	v18de63sdk "github.com/kava-labs/kava/migrate/v0_8/sdk/types"
	v032tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32"
	v032tendermintrpc "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32/rpccore"
	valvesting "github.com/kava-labs/kava/x/validator-vesting"
	v0_3valvesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_3"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	s := r.PathPrefix("/v0_3").Subrouter()

	s.HandleFunc("/node_info", rpc.NodeInfoRequestHandlerFn(cliCtx)).Methods("GET")

	s.HandleFunc("/auth/accounts/{address}", QueryAccountRequestHandlerFn(cliCtx)).Methods("GET")

	s.HandleFunc("/txs/{hash}", QueryTxRequestHandlerFn(cliCtx)).Methods("GET")
	// r.HandleFunc("/txs", QueryTxsRequestHandlerFn(cliCtx)).Methods("GET") // TODO does trust wallet query txs?
	s.HandleFunc("/txs", authrest.BroadcastTxRequest(cliCtx)).Methods("POST")

	s.HandleFunc("/blocks/latest", LatestBlockRequestHandlerFn(cliCtx)).Methods("GET")

	// These endpoints are unchanged between cosmos v18de63 and v0.38.4, but can't import private methods so copy and pasting handler methods.
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

	res, err := node.Block(height)
	if err != nil {
		return nil, err
	}

	// Convert block to old type
	header := v032tendermint.Header{
		Version: v032tendermint.Consensus{
			Block: v032tendermint.Protocol(res.Block.Header.Version.Block),
			App:   v032tendermint.Protocol(res.Block.Header.Version.App),
		},
		ChainID:  res.Block.Header.ChainID,
		Height:   res.Block.Header.Height,
		Time:     res.Block.Header.Time,
		NumTxs:   0, // trust wallet doesn't use this field
		TotalTxs: 0, // trust wallet doesn't use this field

		LastBlockID: res.Block.Header.LastBlockID,

		LastCommitHash:     res.Block.Header.LastCommitHash,
		DataHash:           res.Block.Header.DataHash,
		ValidatorsHash:     res.Block.Header.ValidatorsHash,
		NextValidatorsHash: res.Block.Header.NextValidatorsHash,
		ConsensusHash:      res.Block.Header.ConsensusHash,
		AppHash:            res.Block.Header.AppHash,
		LastResultsHash:    res.Block.Header.LastResultsHash,
		EvidenceHash:       res.Block.Header.EvidenceHash,
		ProposerAddress:    res.Block.Header.ProposerAddress,
	}
	block := v032tendermint.Block{
		Header:     header,
		Data:       res.Block.Data,
		Evidence:   res.Block.Evidence,
		LastCommit: nil, // trust wallet doesn't need to access commit info
	}
	blockMeta := v032tendermint.BlockMeta{
		BlockID: res.BlockID,
		Header:  header,
	}
	oldResponse := v032tendermintrpc.ResultBlock{
		Block:     &block,
		BlockMeta: &blockMeta,
	}

	return codec.Cdc.MarshalJSON(oldResponse)
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
		if err != nil {
			if strings.Contains(err.Error(), hashHexStr) {
				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// convert v0.8 TxResponse to a v0.3 Tx Response
		oldOutput := rollbackTxResponseType(output)
		if oldOutput.Empty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("no transaction found with hash %s", hashHexStr))
		}

		rest.PostProcessResponseBare(w, cliCtx, oldOutput)
	}
}

func rollbackTxResponseType(response sdk.TxResponse) v18de63sdk.TxResponse {
	events := sdk.StringEvents{}
	for _, msgLog := range response.Logs {
		events = append(events, msgLog.Events...)
	}
	return v18de63sdk.TxResponse{
		Height:    response.Height,
		TxHash:    response.TxHash,
		Codespace: response.Codespace,
		Code:      response.Code,
		Data:      response.Data,
		RawLog:    response.RawLog,
		Logs:      nil, // trust wallet doesn't use logs, so leaving them out
		Info:      response.Info,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
		Tx:        response.Tx,
		Timestamp: response.Timestamp,
		Events:    events,
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
