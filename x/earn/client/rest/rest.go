package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/earn/types"
)

type (
	// CommunityPoolDepositProposalReq defines a community pool deposit proposal request body.
	CommunityPoolDepositProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title       string         `json:"title" yaml:"title"`
		Description string         `json:"description" yaml:"description"`
		Amount      sdk.Coin       `json:"amount" yaml:"amount"`
		Deposit     sdk.Coins      `json:"deposit" yaml:"deposit"`
		Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
	}
	// CommunityPoolWithdrawProposalReq defines a community pool deposit proposal request body.
	CommunityPoolWithdrawProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title       string         `json:"title" yaml:"title"`
		Description string         `json:"description" yaml:"description"`
		Amount      sdk.Coin       `json:"amount" yaml:"amount"`
		Deposit     sdk.Coins      `json:"deposit" yaml:"deposit"`
		Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
	}
)

// DepositProposalRESTHandler returns a ProposalRESTHandler that exposes the community pool deposit REST handler with a given sub-route.
func DepositProposalRESTHandler(cliCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ProposalTypeCommunityPoolDeposit,
		Handler:  postDepositProposalHandlerFn(cliCtx),
	}
}

func postDepositProposalHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CommunityPoolDepositProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		content := types.NewCommunityPoolDepositProposal(req.Title, req.Description, req.Amount)
		msg, err := govtypes.NewMsgSubmitProposal(content, req.Deposit, req.Proposer)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// WithdrawProposalRESTHandler returns a ProposalRESTHandler that exposes the community pool deposit REST handler with a given sub-route.
func WithdrawProposalRESTHandler(cliCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ProposalTypeCommunityPoolWithdraw,
		Handler:  postWithdrawProposalHandlerFn(cliCtx),
	}
}

func postWithdrawProposalHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CommunityPoolWithdrawProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		content := types.NewCommunityPoolWithdrawProposal(req.Title, req.Description, req.Amount)
		msg, err := govtypes.NewMsgSubmitProposal(content, req.Deposit, req.Proposer)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
