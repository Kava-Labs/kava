package common

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/kava-labs/kava/x/committee/types"
)

// Note: QueryProposer is copied in from the gov module

const (
	defaultPage  = 1
	defaultLimit = 30 // should be consistent with tendermint/tendermint/rpc/core/pipe.go:19
)

// Proposer contains metadata of a governance proposal used for querying a proposer.
type Proposer struct {
	ProposalID uint64 `json:"proposal_id" yaml:"proposal_id"`
	Proposer   string `json:"proposer" yaml:"proposer"`
}

// NewProposer returns a new Proposer given id and proposer
func NewProposer(proposalID uint64, proposer string) Proposer {
	return Proposer{proposalID, proposer}
}

func (p Proposer) String() string {
	return fmt.Sprintf("Proposal with ID %d was proposed by %s", p.ProposalID, p.Proposer)
}

// QueryProposer will query for a proposer of a governance proposal by ID.
func QueryProposer(cliCtx client.Context, proposalID uint64) (Proposer, error) {
	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalSubmit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	searchResult, err := authtx.QueryTxsByEvents(cliCtx, events, defaultPage, defaultLimit, "")
	if err != nil {
		return Proposer{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single proposal under the given conditions
			if subMsg, ok := msg.(*types.MsgSubmitProposal); ok {
				return NewProposer(proposalID, subMsg.Proposer), nil
			}
		}
	}

	return Proposer{}, fmt.Errorf("failed to find the proposer for proposalID %d", proposalID)
}
