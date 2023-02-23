package committee

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

func NewProposalHandler(k keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.CommitteeChangeProposal:
			return handleCommitteeChangeProposal(ctx, k, c)
		case *types.CommitteeDeleteProposal:
			return handleCommitteeDeleteProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

func handleCommitteeChangeProposal(ctx sdk.Context, k keeper.Keeper, committeeProposal *types.CommitteeChangeProposal) error {
	if err := committeeProposal.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidPubProposal, err.Error())
	}

	// Remove all committee's ongoing proposals
	proposals := k.GetProposalsByCommittee(ctx, committeeProposal.GetNewCommittee().GetID())
	for _, p := range proposals {
		k.CloseProposal(ctx, p, types.Failed)
	}

	// update/create the committee
	k.SetCommittee(ctx, committeeProposal.GetNewCommittee())
	return nil
}

func handleCommitteeDeleteProposal(ctx sdk.Context, k keeper.Keeper, committeeProposal *types.CommitteeDeleteProposal) error {
	if err := committeeProposal.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidPubProposal, err.Error())
	}

	// Remove all committee's ongoing proposals
	proposals := k.GetProposalsByCommittee(ctx, committeeProposal.CommitteeID)
	for _, p := range proposals {
		k.CloseProposal(ctx, p, types.Failed)
	}

	k.DeleteCommittee(ctx, committeeProposal.CommitteeID)
	return nil
}
