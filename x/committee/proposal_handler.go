package committee

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Error {
		switch c := content.(type) {
		case CommitteeChangeProposal:
			return handleCommitteeChangeProposal(ctx, k, c)
		case CommitteeDeleteProposal:
			return handleCommitteeDeleteProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized %s proposal content type: %T", ModuleName, c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

func handleCommitteeChangeProposal(ctx sdk.Context, k Keeper, committeeProposal CommitteeChangeProposal) sdk.Error {
	if err := committeeProposal.ValidateBasic(); err != nil {
		return ErrInvalidCommittee(DefaultCodespace, err.Error())
	}

	// Remove all committee's ongoing proposals
	k.IterateProposals(ctx, func(p Proposal) bool {
		if p.CommitteeID != committeeProposal.NewCommittee.ID {
			return false
		}
		k.DeleteProposalAndVotes(ctx, p.ID)
		return false
	})

	// update/create the committee
	k.SetCommittee(ctx, committeeProposal.NewCommittee)
	return nil
}

func handleCommitteeDeleteProposal(ctx sdk.Context, k Keeper, committeeProposal CommitteeDeleteProposal) sdk.Error {
	if err := committeeProposal.ValidateBasic(); err != nil {
		return ErrInvalidPubProposal(DefaultCodespace, err.Error())
	}

	// Remove all committee's ongoing proposals
	k.IterateProposals(ctx, func(p Proposal) bool {
		if p.CommitteeID != committeeProposal.CommitteeID {
			return false
		}
		k.DeleteProposalAndVotes(ctx, p.ID)
		return false
	})

	k.DeleteCommittee(ctx, committeeProposal.CommitteeID)
	return nil
}
