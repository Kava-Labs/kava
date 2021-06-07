package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/committee/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/committees/{%s}/proposals", types.ModuleName, RestCommitteeID), postProposalHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/proposals/{%s}/votes", types.ModuleName, RestProposalID), postVoteHandlerFn(cliCtx)).Methods("POST")
}

// PostProposalReq defines the properties of a proposal request's body.
type PostProposalReq struct {
	BaseReq     rest.BaseReq      `json:"base_req" yaml:"base_req"`
	PubProposal types.PubProposal `json:"pub_proposal" yaml:"pub_proposal"`
	Proposer    sdk.AccAddress    `json:"proposer" yaml:"proposer"`
}

func postProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse and validate url params
		vars := mux.Vars(r)
		if len(vars[RestCommitteeID]) == 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%s required but not specified", RestCommitteeID))
			return
		}
		committeeID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestCommitteeID])
		if !ok {
			return
		}

		// Parse and validate http request body
		var req PostProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		if err := req.PubProposal.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Create and return a StdTx
		msg := types.NewMsgSubmitProposal(req.PubProposal, req.Proposer, committeeID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// PostVoteReq defines the properties of a vote request's body.
type PostVoteReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Voter   sdk.AccAddress `json:"voter" yaml:"voter"`
	Vote    types.VoteType `json:"vote" yaml:"vote"`
}

func postVoteHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse and validate url params
		vars := mux.Vars(r)
		if len(vars[RestProposalID]) == 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%s required but not specified", RestProposalID))
			return
		}
		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, vars[RestProposalID])
		if !ok {
			return
		}

		if len(vars[RestVote]) == 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%s required but not specified", RestVote))
			return
		}

		rawVote := strings.ToLower(strings.TrimSpace(vars[RestVote]))
		if len(rawVote) == 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid %s: %s", RestVote, rawVote))
			return
		}

		var vote types.VoteType
		switch rawVote {
		case "yes", "y":
			vote = types.Yes
		case "no", "n":
			vote = types.No
		default:
			rest.WriteErrorResponse(w, http.StatusBadRequest, "must specify a valid vote type (\"yes\", \"y\"/\"no\" \"n\")")
			return
		}

		// Parse and validate http request body
		var req PostVoteReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// Create and return a StdTx
		msg := types.NewMsgVote(req.Voter, proposalID, vote)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// This is a rest handler for for the gov module, that handles committee change/delete proposals.
type PostGovProposalReq struct {
	BaseReq  rest.BaseReq     `json:"base_req" yaml:"base_req"`
	Content  govtypes.Content `json:"content" yaml:"content"`
	Proposer sdk.AccAddress   `json:"proposer" yaml:"proposer"`
	Deposit  sdk.Coins        `json:"deposit" yaml:"deposit"`
}

func ProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "committee",
		Handler:  postGovProposalHandlerFn(cliCtx),
	}
}

func postGovProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse and validate http request body
		var req PostGovProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		if err := req.Content.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Create and return a StdTx
		msg := govtypes.NewMsgSubmitProposal(req.Content, req.Deposit, req.Proposer)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
