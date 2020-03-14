package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/committee/client"
	"github.com/kava-labs/kava/x/committee/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/committees", types.ModuleName), queryCommitteesHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/committees/{%s}", types.ModuleName, RestCommitteeID), queryCommitteeHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/committees/{%s}/proposals", types.ModuleName, RestCommitteeID), queryProposalsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/proposals/{%s}", types.ModuleName, RestProposalID), queryProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/proposals/{%s}/proposer", types.ModuleName, RestProposalID), queryProposerHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/proposals/{%s}/tally", types.ModuleName, RestProposalID), queryTallyOnProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/proposals/{%s}/votes", types.ModuleName, RestProposalID), queryVotesOnProposalHandlerFn(cliCtx)).Methods("GET")
	//r.HandleFunc(fmt.Sprintf("/%s/proposals/{%s}/votes/{%s}", types.ModuleName, RestProposalID, RestVoter), queryVoteHandlerFn(cliCtx)).Methods("GET")
	//r.HandleFunc(fmt.Sprintf("/%s/parameters/{%s}", types.ModuleName, RestParamsType), queryParamsHandlerFn(cliCtx)).Methods("GET")
}

// ---------- Committees ----------

func queryCommitteesHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCommittees), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Write response
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryCommitteeHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		if len(vars[RestCommitteeID]) == 0 {
			err := errors.New("committeeID required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		committeeID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestCommitteeID])
		if !ok {
			return
		}
		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(committeeID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCommittee), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Write response
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ---------- Proposals ----------

func queryProposalsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		if len(vars[RestCommitteeID]) == 0 {
			err := errors.New("committeeID required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		committeeID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestCommitteeID])
		if !ok {
			return
		}
		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(committeeID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryProposals), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Write response
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		if len(vars[RestProposalID]) == 0 {
			err := errors.New("proposalID required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestProposalID])
		if !ok {
			return
		}
		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(proposalID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryProposals), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Write response
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryProposerHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestProposalID])
		if !ok {
			return
		}

		// Query
		res, err := client.QueryProposer(cliCtx, proposalID)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Write response
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ---------- Votes ----------

func queryVotesOnProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		if len(vars[RestProposalID]) == 0 {
			err := errors.New("proposalID required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestProposalID])
		if !ok {
			return
		}
		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(proposalID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryVotes), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// TODO should add this feature back
		// var proposal types.Proposal
		// if err := cliCtx.Codec.UnmarshalJSON(res, &proposal); err != nil {
		// 	rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		// 	return
		// }

		// // For inactive proposals we must query the txs directly to get the votes
		// // as they're no longer in state.
		// propStatus := proposal.Status
		// if !(propStatus == types.StatusVotingPeriod || propStatus == types.StatusDepositPeriod) {
		// 	res, err = gcutils.QueryVotesByTxQuery(cliCtx, params)
		// } else {
		// 	res, _, err = cliCtx.QueryWithData("custom/gov/votes", bz)
		// }

		// if err != nil {
		// 	rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		// 	return
		// }

		// Write response
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// func queryVoteHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		strProposalID := vars[RestProposalID]
// 		bechVoterAddr := vars[RestVoter]

// 		if len(strProposalID) == 0 {
// 			err := errors.New("proposalId required but not specified")
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}

// 		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
// 		if !ok {
// 			return
// 		}

// 		if len(bechVoterAddr) == 0 {
// 			err := errors.New("voter address required but not specified")
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}

// 		voterAddr, err := sdk.AccAddressFromBech32(bechVoterAddr)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}

// 		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
// 		if !ok {
// 			return
// 		}

// 		params := types.NewQueryVoteParams(proposalID, voterAddr)

// 		bz, err := cliCtx.Codec.MarshalJSON(params)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}

// 		res, _, err := cliCtx.QueryWithData("custom/gov/vote", bz)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		var vote types.Vote
// 		if err := cliCtx.Codec.UnmarshalJSON(res, &vote); err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}

// 		// For an empty vote, either the proposal does not exist or is inactive in
// 		// which case the vote would be removed from state and should be queried for
// 		// directly via a txs query.
// 		if vote.Empty() {
// 			bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(proposalID))
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 				return
// 			}

// 			res, _, err = cliCtx.QueryWithData("custom/gov/proposal", bz)
// 			if err != nil || len(res) == 0 {
// 				err := fmt.Errorf("proposalID %d does not exist", proposalID)
// 				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
// 				return
// 			}

// 			res, err = gcutils.QueryVoteByTxQuery(cliCtx, params)
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 		}

// 		rest.PostProcessResponse(w, cliCtx, res)
// 	}
// }

func queryTallyOnProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Prepare params for querier
		vars := mux.Vars(r)
		if len(vars[RestProposalID]) == 0 {
			err := errors.New("proposalID required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestProposalID])
		if !ok {
			return
		}
		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(proposalID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryTally), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Write response
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ---------- Params ----------

// func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		paramType := vars[RestParamsType]

// 		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
// 		if !ok {
// 			return
// 		}

// 		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/gov/%s/%s", types.QueryParams, paramType), nil)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
// 			return
// 		}

// 		cliCtx = cliCtx.WithHeight(height)
// 		rest.PostProcessResponse(w, cliCtx, res)
// 	}
// }
