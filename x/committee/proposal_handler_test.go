package committee_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/types"
)

var testTime time.Time = time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)

func NewCommitteeGenState(cdc *codec.Codec, gs committee.GenesisState) app.GenesisState {
	return app.GenesisState{committee.ModuleName: cdc.MustMarshalJSON(gs)}
}

type ProposalHandlerTestSuite struct {
	suite.Suite

	keeper committee.Keeper
	app    app.TestApp
	ctx    sdk.Context

	addresses   []sdk.AccAddress
	testGenesis committee.GenesisState
}

func (suite *ProposalHandlerTestSuite) SetupTest() {
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
	suite.testGenesis = committee.NewGenesisState(
		2,
		[]committee.Committee{
			{
				ID:               1,
				Description:      "This committee is for testing.",
				Members:          suite.addresses[:3],
				Permissions:      []types.Permission{types.GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
			},
			{
				ID:               2,
				Members:          suite.addresses[2:],
				Permissions:      nil,
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
			},
		},
		[]committee.Proposal{
			{ID: 1, CommitteeID: 1, PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."), Deadline: testTime.Add(7 * 24 * time.Hour)},
		},
		[]committee.Vote{
			{ProposalID: 1, Voter: suite.addresses[0]},
		},
	)
}

func (suite *ProposalHandlerTestSuite) TestProposalHandler_ChangeCommittee() {
	testCases := []struct {
		name       string
		proposal   committee.CommitteeChangeProposal
		expectPass bool
	}{
		{
			name: "add new",
			proposal: committee.NewCommitteeChangeProposal(
				"A Title",
				"A proposal description.",
				committee.Committee{
					ID:               34,
					Members:          suite.addresses[:1],
					VoteThreshold:    d("1"),
					ProposalDuration: time.Hour * 24,
				},
			),
			expectPass: true,
		},
		{
			name: "update",
			proposal: committee.NewCommitteeChangeProposal(
				"A Title",
				"A proposal description.",
				committee.Committee{
					ID:               suite.testGenesis.Committees[0].ID,
					Members:          suite.addresses, // add new members
					Permissions:      suite.testGenesis.Committees[0].Permissions,
					VoteThreshold:    suite.testGenesis.Committees[0].VoteThreshold,
					ProposalDuration: suite.testGenesis.Committees[0].ProposalDuration,
				},
			),
			expectPass: true,
		},
		{
			name: "invalid title",
			proposal: committee.NewCommitteeChangeProposal(
				"A Title That Is Much Too Long And Really Quite Unreasonable Given That It Is Trying To Fulfill The Roll Of An Acceptable Governance Proposal Title That Should Succinctly Communicate The Goal And Contents Of The Proposed Proposal To All Parties Involved",
				"A proposal description.",
				suite.testGenesis.Committees[0],
			),
			expectPass: false,
		},
		{
			name: "invalid committee",
			proposal: committee.NewCommitteeChangeProposal(
				"A Title",
				"A proposal description.",
				committee.Committee{
					ID:               suite.testGenesis.Committees[0].ID,
					Members:          append(suite.addresses, suite.addresses[0]), // duplicate address
					Permissions:      suite.testGenesis.Committees[0].Permissions,
					VoteThreshold:    suite.testGenesis.Committees[0].VoteThreshold,
					ProposalDuration: suite.testGenesis.Committees[0].ProposalDuration,
				},
			),
			expectPass: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup
			suite.app = app.NewTestApp()
			suite.keeper = suite.app.GetCommitteeKeeper()
			suite.app = suite.app.InitializeFromGenesisStates(
				NewCommitteeGenState(suite.app.Codec(), suite.testGenesis),
			)
			suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: testTime})
			handler := committee.NewProposalHandler(suite.keeper)

			oldProposals := suite.keeper.GetProposalsByCommittee(suite.ctx, tc.proposal.NewCommittee.ID)

			// Run
			err := handler(suite.ctx, tc.proposal)

			// Check
			if tc.expectPass {
				suite.NoError(err)
				// check committee is accurate
				actualCom, found := suite.keeper.GetCommittee(suite.ctx, tc.proposal.NewCommittee.ID)
				suite.True(found)
				suite.Equal(tc.proposal.NewCommittee, actualCom)

				// check proposals and votes for this committee have been removed
				suite.Empty(suite.keeper.GetProposalsByCommittee(suite.ctx, tc.proposal.NewCommittee.ID))
				for _, p := range oldProposals {
					suite.Empty(suite.keeper.GetVotesByProposal(suite.ctx, p.ID))
				}
			} else {
				suite.Error(err)
				suite.Equal(suite.testGenesis, committee.ExportGenesis(suite.ctx, suite.keeper))
			}
		})
	}
}

func (suite *ProposalHandlerTestSuite) TestProposalHandler_DeleteCommittee() {
	testCases := []struct {
		name       string
		proposal   committee.CommitteeDeleteProposal
		expectPass bool
	}{
		{
			name: "normal",
			proposal: committee.NewCommitteeDeleteProposal(
				"A Title",
				"A proposal description.",
				suite.testGenesis.Committees[0].ID,
			),
			expectPass: true,
		},
		{
			name: "invalid title",
			proposal: committee.NewCommitteeDeleteProposal(
				"A Title That Is Much Too Long And Really Quite Unreasonable Given That It Is Trying To Fulfill The Roll Of An Acceptable Governance Proposal Title That Should Succinctly Communicate The Goal And Contents Of The Proposed Proposal To All Parties Involved",
				"A proposal description.",
				suite.testGenesis.Committees[1].ID,
			),
			expectPass: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup
			suite.app = app.NewTestApp()
			suite.keeper = suite.app.GetCommitteeKeeper()
			suite.app = suite.app.InitializeFromGenesisStates(
				NewCommitteeGenState(suite.app.Codec(), suite.testGenesis),
			)
			suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: testTime})
			handler := committee.NewProposalHandler(suite.keeper)

			oldProposals := suite.keeper.GetProposalsByCommittee(suite.ctx, tc.proposal.CommitteeID)

			// Run
			err := handler(suite.ctx, tc.proposal)

			// Check
			if tc.expectPass {
				suite.NoError(err)
				// check committee has been removed
				_, found := suite.keeper.GetCommittee(suite.ctx, tc.proposal.CommitteeID)
				suite.False(found)

				// check proposals and votes for this committee have been removed
				suite.Empty(suite.keeper.GetProposalsByCommittee(suite.ctx, tc.proposal.CommitteeID))
				for _, p := range oldProposals {
					suite.Empty(suite.keeper.GetVotesByProposal(suite.ctx, p.ID))
				}
			} else {
				suite.Error(err)
				suite.Equal(suite.testGenesis, committee.ExportGenesis(suite.ctx, suite.keeper))
			}
		})
	}
}

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}
