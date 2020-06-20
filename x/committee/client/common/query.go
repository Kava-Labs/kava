package common

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

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
func QueryProposer(cliCtx context.CLIContext, proposalID uint64) (Proposer, error) {
	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalSubmit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	searchResult, err := utils.QueryTxsByEvents(cliCtx, events, defaultPage, defaultLimit)
	if err != nil {
		return Proposer{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.Tx.GetMsgs() {
			// there should only be a single proposal under the given conditions
			if msg.Type() == types.TypeMsgSubmitProposal {
				subMsg := msg.(types.MsgSubmitProposal)
				return NewProposer(proposalID, subMsg.Proposer.String()), nil
			}
		}
	}

	return Proposer{}, fmt.Errorf("failed to find the proposer for proposalID %d", proposalID)
}

// QueryProposalByID returns a proposal from state if present or fallbacks to searching old blocks
func QueryProposalByID(cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string, proposalID uint64) (*types.Proposal, int64, error) {
	bz, err := cdc.MarshalJSON(types.NewQueryProposalParams(proposalID))
	if err != nil {
		return nil, 0, err
	}

	res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryProposal), bz)

	if err == nil {
		var proposal *types.Proposal
		cdc.MustUnmarshalJSON(res, &proposal)

		return proposal, height, nil
	}

	// NOTE: !errors.Is(err, types.ErrUnknownProposal) does not work here
	if err != nil && !strings.Contains(err.Error(), "proposal not found") {
		return nil, 0, err
	}

	res, height, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryNextProposalID), nil)
	if err != nil {
		return nil, 0, err
	}

	var nextProposalID uint64
	cdc.MustUnmarshalJSON(res, &nextProposalID)

	if proposalID >= nextProposalID {
		return nil, 0, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}

	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalSubmit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
	}

	searchResult, err := utils.QueryTxsByEvents(cliCtx, events, defaultPage, defaultLimit)
	if err != nil {
		return nil, 0, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.Tx.GetMsgs() {
			if msg.Type() == types.TypeMsgSubmitProposal {
				subMsg := msg.(types.MsgSubmitProposal)

				// return found Proposal
				// NOTE: no deadline?
				return &types.Proposal{
					ID:          proposalID,
					CommitteeID: subMsg.CommitteeID,
					PubProposal: subMsg.PubProposal,
				}, height, nil
			}
		}
	}

	return nil, 0, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
}
