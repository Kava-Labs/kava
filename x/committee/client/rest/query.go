package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/committee/client/common"
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
}

// ------------------------------------------
//				Committees
// ------------------------------------------

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
		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryCommitteeParams(committeeID))
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

// ------------------------------------------
//				Proposals
// ------------------------------------------

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
		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryCommitteeParams(committeeID))
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

		proposal, height, err := common.QueryProposalByID(cliCtx, cliCtx.Codec, types.ModuleName, proposalID)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := cliCtx.Codec.MarshalJSON(proposal)
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
		res, err := common.QueryProposer(cliCtx, proposalID)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Write response
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ------------------------------------------
//				Votes
// ------------------------------------------

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
			err := errors.New(fmt.Sprintf("%s required but not specified", RestProposalID))
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

		// Write response
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

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
			err := errors.New(fmt.Sprintf("%s required but not specified", RestProposalID))
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
