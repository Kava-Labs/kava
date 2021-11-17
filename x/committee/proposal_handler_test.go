package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

var testTime time.Time = time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)

func NewCommitteeGenState(cdc codec.Codec, gs *types.GenesisState) app.GenesisState {
	return app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(gs)}
}

type ProposalHandlerTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context

	addresses   []sdk.AccAddress
	testGenesis *types.GenesisState
}

func (suite *ProposalHandlerTestSuite) SetupTest() {
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
	suite.testGenesis = types.NewGenesisState(
		2,
		[]types.Committee{
			types.MustNewMemberCommittee(
				1,
				"This committee is for testing.",
				suite.addresses[:3],
				[]types.Permission{&types.GodPermission{}},
				testutil.D("0.667"),
				time.Hour*24*7,
				types.TALLY_OPTION_FIRST_PAST_THE_POST,
			),
			types.MustNewMemberCommittee(
				2,
				"member committee",
				suite.addresses[2:],
				nil,
				testutil.D("0.667"),
				time.Hour*24*7,
				types.TALLY_OPTION_FIRST_PAST_THE_POST,
			),
		},
		types.Proposals{
			types.MustNewProposal(
				govtypes.NewTextProposal("A Title", "A description of this proposal."), 1, 1, testTime.Add(7*24*time.Hour),
			),
		},
		[]types.Vote{
			{ProposalID: 1, Voter: suite.addresses[0], VoteType: types.VOTE_TYPE_YES},
		},
	)
}

func (suite *ProposalHandlerTestSuite) TestProposalHandler_ChangeCommittee() {
	testCases := []struct {
		name       string
		proposal   types.CommitteeChangeProposal
		expectPass bool
	}{
		{
			name: "add new",
			proposal: types.MustNewCommitteeChangeProposal(
				"A Title",
				"A proposal description.",
				types.MustNewMemberCommittee(
					34,
					"member committee",
					suite.addresses[:1],
					[]types.Permission{},
					testutil.D("1"),
					time.Hour*24,
					types.TALLY_OPTION_DEADLINE,
				),
			),
			expectPass: true,
		},
		{
			name: "update",
			proposal: types.MustNewCommitteeChangeProposal(
				"A Title",
				"A proposal description.",
				types.MustNewMemberCommittee(
					suite.testGenesis.GetCommittees()[0].GetID(),
					"member committee",
					suite.addresses, // add new members
					suite.testGenesis.GetCommittees()[0].GetPermissions(),
					suite.testGenesis.GetCommittees()[0].GetVoteThreshold(),
					suite.testGenesis.GetCommittees()[0].GetProposalDuration(),
					types.TALLY_OPTION_FIRST_PAST_THE_POST,
				),
			),
			expectPass: true,
		},
		{
			name: "invalid title",
			proposal: types.MustNewCommitteeChangeProposal(
				"A Title That Is Much Too Long And Really Quite Unreasonable Given That It Is Trying To Fulfill The Roll Of An Acceptable Governance Proposal Title That Should Succinctly Communicate The Goal And Contents Of The Proposed Proposal To All Parties Involved",
				"A proposal description.",
				suite.testGenesis.GetCommittees()[0],
			),
			expectPass: false,
		},
		{
			name: "invalid committee",
			proposal: types.MustNewCommitteeChangeProposal(
				"A Title",
				"A proposal description.",
				types.MustNewMemberCommittee(
					suite.testGenesis.GetCommittees()[0].GetID(),
					"member committee",
					append(suite.addresses, suite.addresses[0]), // duplicate address
					suite.testGenesis.GetCommittees()[0].GetPermissions(),
					suite.testGenesis.GetCommittees()[0].GetVoteThreshold(),
					suite.testGenesis.GetCommittees()[0].GetProposalDuration(),
					types.TALLY_OPTION_DEADLINE,
				),
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
				NewCommitteeGenState(suite.app.AppCodec(), suite.testGenesis),
			)
			suite.ctx = suite.app.NewContext(true, tmproto.Header{Height: 1, Time: testTime})
			handler := committee.NewProposalHandler(suite.keeper)

			oldProposals := suite.keeper.GetProposalsByCommittee(suite.ctx, tc.proposal.GetNewCommittee().GetID())

			// Run
			err := handler(suite.ctx, &tc.proposal)

			// Check
			if tc.expectPass {
				suite.NoError(err)
				// check committee is accurate
				actualCom, found := suite.keeper.GetCommittee(suite.ctx, tc.proposal.GetNewCommittee().GetID())
				suite.True(found)
				testutil.AssertProtoMessageJSON(suite.T(), suite.app.AppCodec(), tc.proposal.GetNewCommittee(), actualCom)

				// check proposals and votes for this committee have been removed
				suite.Empty(suite.keeper.GetProposalsByCommittee(suite.ctx, tc.proposal.GetNewCommittee().GetID()))
				for _, p := range oldProposals {
					suite.Empty(suite.keeper.GetVotesByProposal(suite.ctx, p.ID))
				}
			} else {
				suite.Error(err)
				testutil.AssertProtoMessageJSON(suite.T(), suite.app.AppCodec(), suite.testGenesis, committee.ExportGenesis(suite.ctx, suite.keeper))
			}
		})
	}
}

func (suite *ProposalHandlerTestSuite) TestProposalHandler_DeleteCommittee() {
	testCases := []struct {
		name       string
		proposal   types.CommitteeDeleteProposal
		expectPass bool
	}{
		{
			name: "normal",
			proposal: types.NewCommitteeDeleteProposal(
				"A Title",
				"A proposal description.",
				suite.testGenesis.GetCommittees()[0].GetID(),
			),
			expectPass: true,
		},
		{
			name: "invalid title",
			proposal: types.NewCommitteeDeleteProposal(
				"A Title That Is Much Too Long And Really Quite Unreasonable Given That It Is Trying To Fulfill The Roll Of An Acceptable Governance Proposal Title That Should Succinctly Communicate The Goal And Contents Of The Proposed Proposal To All Parties Involved",
				"A proposal description.",
				suite.testGenesis.GetCommittees()[1].GetID(),
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
				NewCommitteeGenState(suite.app.AppCodec(), suite.testGenesis),
			)
			suite.ctx = suite.app.NewContext(true, tmproto.Header{Height: 1, Time: testTime})
			handler := committee.NewProposalHandler(suite.keeper)

			oldProposals := suite.keeper.GetProposalsByCommittee(suite.ctx, tc.proposal.CommitteeID)

			// Run
			err := handler(suite.ctx, &tc.proposal)

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
				testutil.AssertProtoMessageJSON(suite.T(), suite.app.AppCodec(), suite.testGenesis, committee.ExportGenesis(suite.ctx, suite.keeper))
			}
		})
	}
}

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}
