package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/tendermint/tendermint/crypto"
)

func MustNewMemberCommittee(id uint64, description string, members []string, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption) *MemberCommittee {
	committee, err := NewMemberCommittee(id, description, members, permissions, threshold, duration, tallyOption)
	if err != nil {
		panic(err)
	}
	return committee
}

func MustNewTokenCommitteeNewTokenCommittee(id uint64, description string, members []string, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption, quorum sdk.Dec, tallyDenom string) *TokenCommittee {
	committee, err := NewTokenCommittee(id, description, members, permissions, threshold, duration, tallyOption, quorum, tallyDenom)
	if err != nil {
		panic(err)
	}
	return committee
}

func MustNewProposal(pubProposal PubProposal, id uint64, committeeId uint64, deadline time.Time) Proposal {
	proposal, err := NewProposal(pubProposal, id, committeeId, deadline)
	if err != nil {
		panic(err)
	}
	return proposal
}

func TestGenesisState_Validate(t *testing.T) {
	testTime := time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)
	addresses := []string{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest4"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest5"))).String(),
	}

	testGenesis := NewGenesisState(
		2,
		[]Committee{
			MustNewMemberCommittee(
				1,
				"This members committee is for testing.",
				addresses[:3],
				[]Permission{&GodPermission{}},
				d("0.667"),
				time.Hour*24*7,
				TALLY_OPTION_FIRST_PAST_THE_POST,
			),
			MustNewMemberCommittee(
				2,
				"This members committee is also for testing.",
				addresses[:3],
				nil,
				d("0.8"),
				time.Hour*24*21,
				TALLY_OPTION_FIRST_PAST_THE_POST,
			),
			MustNewTokenCommitteeNewTokenCommittee(
				3,
				"This token committee is for testing.",
				addresses[:3],
				nil,
				d("0.8"),
				time.Hour*24*21,
				TALLY_OPTION_DEADLINE,
				sdk.MustNewDecFromStr("0.4"),
				"hard",
			),
		},
		[]Proposal{MustNewProposal(
			govtypes.NewTextProposal("A Title", "A description of this proposal."), 1, 1, testTime.Add(7*24*time.Hour)),
		},
		[]Vote{
			{ProposalId: 1, Voter: addresses[0], VoteType: VOTE_TYPE_YES},
			{ProposalId: 1, Voter: addresses[1], VoteType: VOTE_TYPE_YES},
		},
	)

	testCases := []struct {
		name       string
		genState   *GenesisState
		expectPass bool
	}{
		{
			name:       "default",
			genState:   DefaultGenesisState(),
			expectPass: true,
		},
		{
			name:       "normal",
			genState:   testGenesis,
			expectPass: true,
		},
		{
			name: "duplicate committee IDs",
			genState: NewGenesisState(
				testGenesis.NextProposalId,
				append(testGenesis.GetCommittees(), testGenesis.GetCommittees()[0]),
				testGenesis.Proposals,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "invalid committee",
			genState: NewGenesisState(
				testGenesis.NextProposalId,
				append(testGenesis.GetCommittees(), &MemberCommittee{BaseCommittee: &BaseCommittee{}}),
				testGenesis.Proposals,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "duplicate proposal IDs",
			genState: NewGenesisState(
				testGenesis.NextProposalId,
				testGenesis.GetCommittees(),
				append(testGenesis.Proposals, testGenesis.Proposals[0]),
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "invalid NextProposalId",
			genState: NewGenesisState(
				0,
				testGenesis.GetCommittees(),
				testGenesis.Proposals,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "proposal without committee",
			genState: NewGenesisState(
				testGenesis.NextProposalId+1,
				testGenesis.GetCommittees(),
				append(
					testGenesis.Proposals,
					MustNewProposal(
						govtypes.NewTextProposal("A Title", "A description of this proposal."),
						testGenesis.NextProposalId,
						47, // doesn't exist
						testTime.Add(7*24*time.Hour),
					),
				),
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "invalid proposal",
			genState: NewGenesisState(
				testGenesis.NextProposalId,
				testGenesis.GetCommittees(),
				append(testGenesis.Proposals, Proposal{}),
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "vote without proposal",
			genState: NewGenesisState(
				testGenesis.NextProposalId,
				testGenesis.GetCommittees(),
				nil,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "invalid vote",
			genState: NewGenesisState(
				testGenesis.NextProposalId,
				testGenesis.GetCommittees(),
				testGenesis.Proposals,
				append(testGenesis.Votes, Vote{}),
			),
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.genState.Validate()

			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
