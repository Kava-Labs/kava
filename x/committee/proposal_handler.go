package committee

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case CommitteeChangeProposal:
			return handleCommitteeChangeProposal(ctx, k, c)
		case CommitteeDeleteProposal:
			return handleCommitteeDeleteProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", ModuleName, c)
		}
	}
}

func handleCommitteeChangeProposal(ctx sdk.Context, k Keeper, committeeProposal CommitteeChangeProposal) error {
	if err := committeeProposal.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(ErrInvalidPubProposal, err.Error())
	}

	// Remove all committee's ongoing proposals
	proposals := k.GetProposalsByCommittee(ctx, committeeProposal.NewCommittee.ID)
	for _, p := range proposals {
		k.DeleteProposalAndVotes(ctx, p.ID)
	}

	// update/create the committee
	k.SetCommittee(ctx, committeeProposal.NewCommittee)
	return nil
}

func handleCommitteeDeleteProposal(ctx sdk.Context, k Keeper, committeeProposal CommitteeDeleteProposal) error {
	if err := committeeProposal.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(ErrInvalidPubProposal, err.Error())
	}

	// Remove all committee's ongoing proposals
	proposals := k.GetProposalsByCommittee(ctx, committeeProposal.CommitteeID)
	for _, p := range proposals {
		k.DeleteProposalAndVotes(ctx, p.ID)
	}

	k.DeleteCommittee(ctx, committeeProposal.CommitteeID)
	return nil
}
