package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tendermint/tendermint/crypto"
)

func d(s string) sdk.Dec { return sdk.MustNewDecFromStr(s) }
func TestGenesisState_Validate(t *testing.T) {
	testTime := time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)
	addresses := []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest4"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest5"))),
	}
	testGenesis := GenesisState{
		NextProposalID: 2,
		Committees: []Committee{
			{
				ID:               1,
				Description:      "This committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
			},
			{
				ID:               2,
				Description:      "This committee is also for testing.",
				Members:          addresses[2:],
				Permissions:      nil,
				VoteThreshold:    d("0.8"),
				ProposalDuration: time.Hour * 24 * 21,
			},
		},
		Proposals: []Proposal{
			{ID: 1, CommitteeID: 1, PubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."), Deadline: testTime.Add(7 * 24 * time.Hour)},
		},
		Votes: []Vote{
			{ProposalID: 1, Voter: addresses[0]},
			{ProposalID: 1, Voter: addresses[1]},
		},
	}

	testCases := []struct {
		name       string
		genState   GenesisState
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
			genState: GenesisState{
				NextProposalID: testGenesis.NextProposalID,
				Committees:     append(testGenesis.Committees, testGenesis.Committees[0]),
				Proposals:      testGenesis.Proposals,
				Votes:          testGenesis.Votes,
			},
			expectPass: false,
		},
		{
			name: "invalid committee",
			genState: GenesisState{
				NextProposalID: testGenesis.NextProposalID,
				Committees:     append(testGenesis.Committees, Committee{}),
				Proposals:      testGenesis.Proposals,
				Votes:          testGenesis.Votes,
			},
			expectPass: false,
		},
		{
			name: "duplicate proposal IDs",
			genState: GenesisState{
				NextProposalID: testGenesis.NextProposalID,
				Committees:     testGenesis.Committees,
				Proposals:      append(testGenesis.Proposals, testGenesis.Proposals[0]),
				Votes:          testGenesis.Votes,
			},
			expectPass: false,
		},
		{
			name: "invalid NextProposalID",
			genState: GenesisState{
				NextProposalID: 0,
				Committees:     testGenesis.Committees,
				Proposals:      testGenesis.Proposals,
				Votes:          testGenesis.Votes,
			},
			expectPass: false,
		},
		{
			name: "proposal without committee",
			genState: GenesisState{
				NextProposalID: testGenesis.NextProposalID + 1,
				Committees:     testGenesis.Committees,
				Proposals: append(
					testGenesis.Proposals,
					Proposal{
						ID:          testGenesis.NextProposalID,
						PubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."),
						CommitteeID: 247, // doesn't exist
					}),
				Votes: testGenesis.Votes,
			},
			expectPass: false,
		},
		{
			name: "invalid proposal",
			genState: GenesisState{
				NextProposalID: testGenesis.NextProposalID,
				Committees:     testGenesis.Committees,
				Proposals:      append(testGenesis.Proposals, Proposal{}),
				Votes:          testGenesis.Votes,
			},
			expectPass: false,
		},
		{
			name: "vote without proposal",
			genState: GenesisState{
				NextProposalID: testGenesis.NextProposalID,
				Committees:     testGenesis.Committees,
				Proposals:      nil,
				Votes:          testGenesis.Votes,
			},
			expectPass: false,
		},
		{
			name: "invalid vote",
			genState: GenesisState{
				NextProposalID: testGenesis.NextProposalID,
				Committees:     testGenesis.Committees,
				Proposals:      testGenesis.Proposals,
				Votes:          append(testGenesis.Votes, Vote{}),
			},
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
