package v0_16

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	v040distr "github.com/cosmos/cosmos-sdk/x/distribution/types"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v040params "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"
	v040upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v015committee "github.com/kava-labs/kava/x/committee/legacy/v0_15"
	v016committee "github.com/kava-labs/kava/x/committee/types"
	v015kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_15"
	v016kavadist "github.com/kava-labs/kava/x/kavadist/types"
)

func migratePermission(v015permission v015committee.Permission) *codectypes.Any {
	var protoProposal proto.Message

	switch v015permission := v015permission.(type) {
	case v015committee.GodPermission:
		{
			protoProposal = &v016committee.GodPermission{}
		}
	case v015committee.TextPermission:
		{
			protoProposal = &v016committee.TextPermission{}
		}
	case v015committee.SoftwareUpgradePermission:
		{
			protoProposal = &v016committee.SoftwareUpgradePermission{}
		}
	case v015committee.SimpleParamChangePermission:
		{
			changes := make(v016committee.AllowedParamsChanges, len(v015permission.AllowedParams))
			for i, param := range v015permission.AllowedParams {
				changes[i] = v016committee.AllowedParamsChange{
					Subspace: param.Subspace,
					Key:      param.Key,
				}
			}
			protoProposal = &v016committee.ParamsChangePermission{
				AllowedParamsChanges: changes,
			}
		}
	case v015committee.SubParamChangePermission:
		{
			// TODO: Not implemented
			// for now just convert these params change permission without sub param restrictions
			changes := make(v016committee.AllowedParamsChanges, len(v015permission.AllowedParams))
			for i, param := range v015permission.AllowedParams {
				changes[i] = v016committee.AllowedParamsChange{
					Subspace: param.Subspace,
					Key:      param.Key,
				}
			}
			protoProposal = &v016committee.ParamsChangePermission{
				AllowedParamsChanges: changes,
			}
		}
	default:
		panic(fmt.Errorf("'%s' is not a valid permission", v015permission))
	}

	// Convert the content into Any.
	contentAny, err := codectypes.NewAnyWithValue(protoProposal)
	if err != nil {
		panic(err)
	}

	return contentAny
}

func migrateTallyOption(oldTallyOption v015committee.TallyOption) v016committee.TallyOption {
	switch oldTallyOption {
	case v015committee.NullTallyOption:
		return v016committee.TALLY_OPTION_UNSPECIFIED
	case v015committee.FirstPastThePost:
		return v016committee.TALLY_OPTION_FIRST_PAST_THE_POST
	case v015committee.Deadline:
		return v016committee.TALLY_OPTION_DEADLINE
	default:
		panic(fmt.Errorf("'%s' is not a valid tally option", oldTallyOption))
	}
}

func migrateCommittee(committee v015committee.Committee) *codectypes.Any {
	var protoProposal proto.Message
	switch committee := committee.(type) {
	case v015committee.MemberCommittee:
		{
			permissions := make([]*codectypes.Any, len(committee.Permissions))
			for i, permission := range committee.Permissions {
				permissions[i] = migratePermission(permission)
			}

			protoProposal = &v016committee.MemberCommittee{
				BaseCommittee: &v016committee.BaseCommittee{
					ID:               committee.ID,
					Description:      committee.Description,
					Members:          committee.Members,
					Permissions:      permissions,
					VoteThreshold:    committee.VoteThreshold,
					ProposalDuration: committee.ProposalDuration,
					TallyOption:      migrateTallyOption(committee.TallyOption),
				},
			}
		}
	case v015committee.TokenCommittee:
		{
			permissions := make([]*codectypes.Any, len(committee.Permissions))
			for i, permission := range committee.Permissions {
				permissions[i] = migratePermission(permission)
			}

			protoProposal = &v016committee.TokenCommittee{
				BaseCommittee: &v016committee.BaseCommittee{
					ID:               committee.ID,
					Description:      committee.Description,
					Members:          committee.Members,
					Permissions:      permissions,
					VoteThreshold:    committee.VoteThreshold,
					ProposalDuration: committee.ProposalDuration,
					TallyOption:      migrateTallyOption(committee.TallyOption),
				},
				Quorum:     committee.Quorum,
				TallyDenom: committee.TallyDenom,
			}
		}
	default:
		panic(fmt.Errorf("'%s' is not a valid committee", committee))
	}

	// Convert the content into Any.
	contentAny, err := codectypes.NewAnyWithValue(protoProposal)
	if err != nil {
		panic(err)
	}

	return contentAny
}

func migrateCommittees(v015committees v015committee.Committees) []*codectypes.Any {
	committees := make([]*codectypes.Any, len(v015committees))
	for i, committee := range v015committees {
		committees[i] = migrateCommittee(committee)
	}
	return committees
}

func migrateContent(oldContent v036gov.Content) *codectypes.Any {
	var protoProposal proto.Message

	switch oldContent := oldContent.(type) {
	case v036gov.TextProposal:
		{
			protoProposal = &v040gov.TextProposal{
				Title:       oldContent.Title,
				Description: oldContent.Description,
			}
			// Convert the content into Any.
			contentAny, err := codectypes.NewAnyWithValue(protoProposal)
			if err != nil {
				panic(err)
			}

			return contentAny
		}
	case v036distr.CommunityPoolSpendProposal:
		{
			protoProposal = &v040distr.CommunityPoolSpendProposal{
				Title:       oldContent.Title,
				Description: oldContent.Description,
				Recipient:   oldContent.Recipient.String(),
				Amount:      oldContent.Amount,
			}
		}
	case v038upgrade.CancelSoftwareUpgradeProposal:
		{
			protoProposal = &v040upgrade.CancelSoftwareUpgradeProposal{
				Description: oldContent.Description,
				Title:       oldContent.Title,
			}
		}
	case v038upgrade.SoftwareUpgradeProposal:
		{
			protoProposal = &v040upgrade.SoftwareUpgradeProposal{
				Description: oldContent.Description,
				Title:       oldContent.Title,
				Plan: v040upgrade.Plan{
					Name:   oldContent.Plan.Name,
					Height: oldContent.Plan.Height,
					Info:   oldContent.Plan.Info,
				},
			}
		}
	case v036params.ParameterChangeProposal:
		{
			newChanges := make([]v040params.ParamChange, len(oldContent.Changes))
			for i, oldChange := range oldContent.Changes {
				newChanges[i] = v040params.ParamChange{
					Subspace: oldChange.Subspace,
					Key:      oldChange.Key,
					Value:    oldChange.Value,
				}
			}

			protoProposal = &v040params.ParameterChangeProposal{
				Description: oldContent.Description,
				Title:       oldContent.Title,
				Changes:     newChanges,
			}
		}
	case v015kavadist.CommunityPoolMultiSpendProposal:
		{
			newRecipients := make([]v016kavadist.MultiSpendRecipient, len(oldContent.RecipientList))
			for i, recipient := range oldContent.RecipientList {
				newRecipients[i] = v016kavadist.MultiSpendRecipient{
					Address: recipient.Address.String(),
					Amount:  recipient.Amount,
				}
			}

			protoProposal = &v016kavadist.CommunityPoolMultiSpendProposal{
				Description:   oldContent.Description,
				Title:         oldContent.Title,
				RecipientList: newRecipients,
			}
		}
	default:
		panic(fmt.Errorf("%T is not a valid proposal content type", oldContent))
	}

	// Convert the content into Any.
	contentAny, err := codectypes.NewAnyWithValue(protoProposal)
	if err != nil {
		panic(err)
	}

	return contentAny
}

func migrateProposals(v015proposals []v015committee.Proposal) v016committee.Proposals {
	proposals := make(v016committee.Proposals, len(v015proposals))
	for i, v15proposal := range v015proposals {
		proposals[i] = v016committee.Proposal{
			ID:          v15proposal.ID,
			Content:     migrateContent(v15proposal.PubProposal),
			CommitteeID: v15proposal.CommitteeID,
			Deadline:    v15proposal.Deadline,
		}
	}
	return proposals
}

func migrateVoteType(oldVoteType v015committee.VoteType) v016committee.VoteType {
	switch oldVoteType {
	case v015committee.Yes:
		return v016committee.VOTE_TYPE_YES
	case v015committee.No:
		return v016committee.VOTE_TYPE_NO
	case v015committee.Abstain:
		return v016committee.VOTE_TYPE_ABSTAIN
	case v015committee.NullVoteType:
		return v016committee.VOTE_TYPE_UNSPECIFIED
	default:
		panic(fmt.Errorf("'%s' is not a valid vote type", oldVoteType))
	}
}

func migrateVotes(v15votes []v015committee.Vote) []v016committee.Vote {
	votes := make([]v016committee.Vote, len(v15votes))
	for i, v15vote := range v15votes {
		votes[i] = v016committee.Vote{
			ProposalID: v15vote.ProposalID,
			Voter:      v15vote.Voter,
			VoteType:   migrateVoteType(v15vote.VoteType),
		}
	}
	return votes
}

func Migrate(oldState v015committee.GenesisState) *v016committee.GenesisState {
	newState := v016committee.GenesisState{
		NextProposalID: oldState.NextProposalID,
		Committees:     migrateCommittees(oldState.Committees),
		Proposals:      migrateProposals(oldState.Proposals),
		Votes:          migrateVotes(oldState.Votes),
	}
	return &newState
}
