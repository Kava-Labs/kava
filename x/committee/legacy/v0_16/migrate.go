package v0_16

import (
	"fmt"
	"reflect"
	"sort"

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

	v016bep3types "github.com/kava-labs/kava/x/bep3/types"
	v016cdptypes "github.com/kava-labs/kava/x/cdp/types"
	v015committee "github.com/kava-labs/kava/x/committee/legacy/v0_15"
	v016committee "github.com/kava-labs/kava/x/committee/types"
	v016hardtypes "github.com/kava-labs/kava/x/hard/types"
	v015kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_15"
	v016kavadist "github.com/kava-labs/kava/x/kavadist/types"
	v016pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// migrateWhitelist returns an string slice of json keys that should be whitelisted on the whitelist interface
func migrateWhitelist(whitelist interface{}, ignoredTag string) []string {
	allowed := []string{}
	v := reflect.ValueOf(whitelist)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		tag := typeOfS.Field(i).Tag.Get("json")
		if tag != ignoredTag && tag != "" {
			val, ok := v.Field(i).Interface().(bool)
			if ok && val {
				allowed = append(allowed, tag)
			}
		}
	}
	sort.Strings(allowed)
	return allowed
}

// isSubparamAllowed returns true if the subspace and key is allowed in the v15 permissions
func isSubparamAllowed(permission v015committee.SubParamChangePermission, subspace string, key string) bool {
	for _, allowed := range permission.AllowedParams {
		if allowed.Key == key && allowed.Subspace == subspace {
			return true
		}
	}
	return false
}

type subspaceKeyPair struct {
	key      []byte
	subspace string
}

// migrateSubParamPermissions converts v15 SubParamChangePermissions to v16 ParamsChangePermission
func migrateSubParamPermissions(permission v015committee.SubParamChangePermission, isStabilityCommittee bool) *v016committee.ParamsChangePermission {
	changes := v016committee.AllowedParamsChanges{}

	// migrate allowed params
	pairsToAvoid := []subspaceKeyPair{
		{key: v016cdptypes.KeyCollateralParams, subspace: v016cdptypes.ModuleName},
		{key: v016cdptypes.KeyDebtParam, subspace: v016cdptypes.ModuleName},
		{key: v016bep3types.KeyAssetParams, subspace: v016bep3types.ModuleName},
		{key: v016pricefeedtypes.KeyMarkets, subspace: v016pricefeedtypes.ModuleName},
		{key: v016hardtypes.KeyMoneyMarkets, subspace: v016hardtypes.ModuleName},
	}
	for _, allowed := range permission.AllowedParams {
		shouldAvoid := false
		for _, pair := range pairsToAvoid {
			if string(pair.key) == allowed.Key && pair.subspace == allowed.Subspace {
				shouldAvoid = true
				break
			}
		}
		if !shouldAvoid {
			changes = append(changes, v016committee.AllowedParamsChange{
				Subspace: allowed.Subspace,
				Key:      allowed.Key,
			})
		}
	}

	// migrate collateral params
	if isSubparamAllowed(permission, v016cdptypes.ModuleName, string(v016cdptypes.KeyCollateralParams)) {
		change := v016committee.AllowedParamsChange{
			Key:      string(v016cdptypes.KeyCollateralParams),
			Subspace: string(v016cdptypes.ModuleName),
		}
		requirements := []v016committee.SubparamRequirement{}
		for _, param := range permission.AllowedCollateralParams {
			requirement := v016committee.SubparamRequirement{
				Key:                        "type",
				Val:                        param.Type,
				AllowedSubparamAttrChanges: []string{},
			}
			allowed := migrateWhitelist(param, "type")
			requirement.AllowedSubparamAttrChanges = allowed
			requirements = append(requirements, requirement)
		}

		// add new requirement for stability committee
		if isStabilityCommittee {
			requirement := v016committee.SubparamRequirement{
				Key: "type",
				Val: "swp-a",
				AllowedSubparamAttrChanges: []string{
					"auction_size", "check_collateralization_index_count", "debt_limit",
					"keeper_reward_percentage", "stability_fee",
				},
			}
			requirements = append(requirements, requirement)
		}

		change.MultiSubparamsRequirements = requirements
		changes = append(changes, change)
	}

	// migrate debt params
	if isSubparamAllowed(permission, string(v016cdptypes.ModuleName), string(v016cdptypes.KeyDebtParam)) {
		change := v016committee.AllowedParamsChange{
			Subspace:                   v016cdptypes.ModuleName,
			Key:                        string(v016cdptypes.KeyDebtParam),
			SingleSubparamAllowedAttrs: migrateWhitelist(permission.AllowedDebtParam, ""),
		}
		changes = append(changes, change)
	}

	// migrate asset params
	if isSubparamAllowed(permission, string(v016bep3types.ModuleName), string(v016bep3types.KeyAssetParams)) {
		change := v016committee.AllowedParamsChange{
			Key:      string(v016bep3types.KeyAssetParams),
			Subspace: string(v016bep3types.ModuleName),
		}
		requirements := []v016committee.SubparamRequirement{}
		for _, param := range permission.AllowedAssetParams {
			requirement := v016committee.SubparamRequirement{
				Key:                        "denom",
				Val:                        param.Denom,
				AllowedSubparamAttrChanges: []string{},
			}
			allowed := migrateWhitelist(param, "denom")
			requirement.AllowedSubparamAttrChanges = allowed
			requirements = append(requirements, requirement)
		}
		change.MultiSubparamsRequirements = requirements
		changes = append(changes, change)
	}

	// migrate markets
	if isSubparamAllowed(permission, string(v016pricefeedtypes.ModuleName), string(v016pricefeedtypes.KeyMarkets)) {
		change := v016committee.AllowedParamsChange{
			Key:      string(v016pricefeedtypes.KeyMarkets),
			Subspace: string(v016pricefeedtypes.ModuleName),
		}
		requirements := []v016committee.SubparamRequirement{}
		for _, param := range permission.AllowedMarkets {
			requirement := v016committee.SubparamRequirement{
				Key:                        "market_id",
				Val:                        param.MarketID,
				AllowedSubparamAttrChanges: []string{},
			}
			allowed := migrateWhitelist(param, "market_id")
			requirement.AllowedSubparamAttrChanges = allowed
			requirements = append(requirements, requirement)
		}
		change.MultiSubparamsRequirements = requirements
		changes = append(changes, change)
	}

	// migrate money markets
	if isSubparamAllowed(permission, string(v016hardtypes.ModuleName), string(v016hardtypes.KeyMoneyMarkets)) {
		change := v016committee.AllowedParamsChange{
			Key:      string(v016hardtypes.KeyMoneyMarkets),
			Subspace: string(v016hardtypes.ModuleName),
		}
		requirements := []v016committee.SubparamRequirement{}
		for _, param := range permission.AllowedMoneyMarkets {
			requirement := v016committee.SubparamRequirement{
				Key:                        "denom",
				Val:                        param.Denom,
				AllowedSubparamAttrChanges: []string{},
			}
			allowed := migrateWhitelist(param, "denom")
			requirement.AllowedSubparamAttrChanges = allowed
			requirements = append(requirements, requirement)
		}

		// add new requirement for stability committee
		if isStabilityCommittee {
			requirement := v016committee.SubparamRequirement{
				Key: "denom",
				Val: "swp",
				AllowedSubparamAttrChanges: []string{
					"borrow_limit", "interest_rate_model",
					"keeper_reward_percentage", "reserve_factor",
				},
			}
			requirements = append(requirements, requirement)
		}

		change.MultiSubparamsRequirements = requirements
		changes = append(changes, change)
	}

	return &v016committee.ParamsChangePermission{
		AllowedParamsChanges: changes,
	}
}

func migratePermission(v015permission v015committee.Permission, isStabilityCommittee bool) *codectypes.Any {
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
			protoProposal = migrateSubParamPermissions(v015permission, isStabilityCommittee)
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
				isStabilityCommittee := committee.GetID() == 1
				permissions[i] = migratePermission(permission, isStabilityCommittee)
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
				permissions[i] = migratePermission(permission, false)
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

	// Make some updates to the stability committee
	if committee.GetID() == 1 {
		// Add requirement to collatora params

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
