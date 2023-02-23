package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

func TestGenesisState_Validate(t *testing.T) {
	testTime := time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)
	addresses := []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest4"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest5"))),
	}

	testGenesis := types.NewGenesisState(
		2,
		[]types.Committee{
			types.MustNewMemberCommittee(
				1,
				"This members committee is for testing.",
				addresses[:3],
				nil,
				testutil.D("0.667"),
				time.Hour*24*7,
				types.TALLY_OPTION_FIRST_PAST_THE_POST,
			),
			types.MustNewMemberCommittee(
				2,
				"This members committee is also for testing.",
				addresses[:3],
				nil,
				testutil.D("0.8"),
				time.Hour*24*21,
				types.TALLY_OPTION_FIRST_PAST_THE_POST,
			),
			types.MustNewTokenCommittee(
				3,
				"This token committee is for testing.",
				addresses[:3],
				nil,
				testutil.D("0.8"),
				time.Hour*24*21,
				types.TALLY_OPTION_DEADLINE,
				sdk.MustNewDecFromStr("0.4"),
				"hard",
			),
		},
		types.Proposals{
			types.MustNewProposal(
				govv1beta1.NewTextProposal("A Title", "A description of this proposal."), 1, 1, testTime.Add(7*24*time.Hour)),
		},
		[]types.Vote{
			{ProposalID: 1, Voter: addresses[0], VoteType: types.VOTE_TYPE_YES},
			{ProposalID: 1, Voter: addresses[1], VoteType: types.VOTE_TYPE_YES},
		},
	)

	testCases := []struct {
		name       string
		genState   *types.GenesisState
		expectPass bool
	}{
		{
			name:       "default",
			genState:   types.DefaultGenesisState(),
			expectPass: true,
		},
		{
			name:       "normal",
			genState:   testGenesis,
			expectPass: true,
		},
		{
			name: "duplicate committee IDs",
			genState: types.NewGenesisState(
				testGenesis.NextProposalID,
				append(testGenesis.GetCommittees(), testGenesis.GetCommittees()[0]),
				testGenesis.Proposals,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "invalid committee",
			genState: types.NewGenesisState(
				testGenesis.NextProposalID,
				append(testGenesis.GetCommittees(), &types.MemberCommittee{BaseCommittee: &types.BaseCommittee{}}),
				testGenesis.Proposals,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "duplicate proposal IDs",
			genState: types.NewGenesisState(
				testGenesis.NextProposalID,
				testGenesis.GetCommittees(),
				append(testGenesis.Proposals, testGenesis.Proposals[0]),
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "invalid NextProposalID",
			genState: types.NewGenesisState(
				0,
				testGenesis.GetCommittees(),
				testGenesis.Proposals,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "proposal without committee",
			genState: types.NewGenesisState(
				testGenesis.NextProposalID+1,
				testGenesis.GetCommittees(),
				append(
					testGenesis.Proposals,
					types.MustNewProposal(
						govv1beta1.NewTextProposal("A Title", "A description of this proposal."),
						testGenesis.NextProposalID,
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
			genState: types.NewGenesisState(
				testGenesis.NextProposalID,
				testGenesis.GetCommittees(),
				append(testGenesis.Proposals, types.Proposal{}),
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "vote without proposal",
			genState: types.NewGenesisState(
				testGenesis.NextProposalID,
				testGenesis.GetCommittees(),
				nil,
				testGenesis.Votes,
			),
			expectPass: false,
		},
		{
			name: "invalid vote",
			genState: types.NewGenesisState(
				testGenesis.NextProposalID,
				testGenesis.GetCommittees(),
				testGenesis.Proposals,
				append(testGenesis.Votes, types.Vote{}),
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
