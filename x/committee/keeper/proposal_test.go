package keeper_test

import (
	"reflect"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/types"
)

func (suite *KeeperTestSuite) TestSubmitProposal() {
	normalCom := types.Committee{
		ID:                  12,
		Description:         "This committee is for testing.",
		Members:             suite.addresses[:2],
		Permissions:         []types.Permission{types.GodPermission{}},
		VoteThreshold:       d("0.667"),
		MaxProposalDuration: time.Hour * 24 * 7,
	}
	noPermissionsCom := normalCom
	noPermissionsCom.Permissions = []types.Permission{}

	testcases := []struct {
		name        string
		committee   types.Committee
		pubProposal types.PubProposal
		proposer    sdk.AccAddress
		committeeID uint64
		expectPass  bool
	}{
		{
			name:        "normal",
			committee:   normalCom,
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    normalCom.Members[0],
			committeeID: normalCom.ID,
			expectPass:  true,
		},
		{
			name:        "invalid proposal",
			committee:   normalCom,
			pubProposal: nil,
			proposer:    normalCom.Members[0],
			committeeID: normalCom.ID,
			expectPass:  false,
		},
		{
			name: "missing committee",
			// no committee
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    suite.addresses[0],
			committeeID: 0,
			expectPass:  false,
		},
		{
			name:        "not a member",
			committee:   normalCom,
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    suite.addresses[4],
			committeeID: normalCom.ID,
			expectPass:  false,
		},
		{
			name:        "not enough permissions",
			committee:   noPermissionsCom,
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    noPermissionsCom.Members[0],
			committeeID: noPermissionsCom.ID,
			expectPass:  false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests,
			// which would mean the app state is not be reset between subtests.
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{})
			tApp.InitializeFromGenesisStates()
			// setup committee (if required)
			if !(reflect.DeepEqual(tc.committee, types.Committee{})) {
				keeper.SetCommittee(ctx, tc.committee)
			}

			id, err := keeper.SubmitProposal(ctx, tc.proposer, tc.committeeID, tc.pubProposal)

			if tc.expectPass {
				suite.NoError(err)
				pr, found := keeper.GetProposal(ctx, id)
				suite.True(found)
				suite.Equal(tc.committeeID, pr.CommitteeID)
				suite.Equal(ctx.BlockTime().Add(tc.committee.MaxProposalDuration), pr.Deadline)
			} else {
				suite.NotNil(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestAddVote() {
	normalCom := types.Committee{
		ID:          12,
		Members:     suite.addresses[:2],
		Permissions: []types.Permission{types.GodPermission{}},
	}
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	testcases := []struct {
		name       string
		proposalID uint64
		voter      sdk.AccAddress
		voteTime   time.Time
		expectPass bool
	}{
		{
			name:       "normal",
			proposalID: types.DefaultNextProposalID,
			voter:      normalCom.Members[0],
			expectPass: true,
		},
		{
			name:       "nonexistent proposal",
			proposalID: 9999999,
			voter:      normalCom.Members[0],
			expectPass: false,
		},
		{
			name:       "voter not committee member",
			proposalID: types.DefaultNextProposalID,
			voter:      suite.addresses[4],
			expectPass: false,
		},
		{
			name:       "proposal expired",
			proposalID: types.DefaultNextProposalID,
			voter:      normalCom.Members[0],
			voteTime:   firstBlockTime.Add(normalCom.MaxProposalDuration),
			expectPass: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests, which would mean the app state is not be reset between subtests.
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})
			tApp.InitializeFromGenesisStates()

			// setup the committee and proposal
			keeper.SetCommittee(ctx, normalCom)
			_, err := keeper.SubmitProposal(ctx, normalCom.Members[0], normalCom.ID, gov.NewTextProposal("A Title", "A description of this proposal."))
			suite.NoError(err)

			ctx = ctx.WithBlockTime(tc.voteTime)
			err = keeper.AddVote(ctx, tc.proposalID, tc.voter)

			if tc.expectPass {
				suite.NoError(err)
				_, found := keeper.GetVote(ctx, tc.proposalID, tc.voter)
				suite.True(found)
			} else {
				suite.NotNil(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetProposalResult() {
	normalCom := types.Committee{
		ID:                  12,
		Description:         "This committee is for testing.",
		Members:             suite.addresses[:5],
		Permissions:         []types.Permission{types.GodPermission{}},
		VoteThreshold:       d("0.667"),
		MaxProposalDuration: time.Hour * 24 * 7,
	}
	var defaultID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	testcases := []struct {
		name           string
		committee      types.Committee
		votes          []types.Vote
		proposalPasses bool
		expectPass     bool
	}{
		{
			name:      "enough votes",
			committee: normalCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: suite.addresses[0]},
				{ProposalID: defaultID, Voter: suite.addresses[1]},
				{ProposalID: defaultID, Voter: suite.addresses[2]},
				{ProposalID: defaultID, Voter: suite.addresses[3]},
			},
			proposalPasses: true,
			expectPass:     true,
		},
		{
			name:      "not enough votes",
			committee: normalCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: suite.addresses[0]},
			},
			proposalPasses: false,
			expectPass:     true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests, which would mean the app state is not be reset between subtests.
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})

			tApp.InitializeFromGenesisStates(
				committeeGenState(
					tApp.Codec(),
					[]types.Committee{tc.committee},
					[]types.Proposal{{
						PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
						ID:          defaultID,
						CommitteeID: tc.committee.ID,
						Deadline:    firstBlockTime.Add(time.Hour * 24 * 7),
					}},
					tc.votes,
				),
			)

			proposalPasses, err := keeper.GetProposalResult(ctx, defaultID)

			if tc.expectPass {
				suite.NoError(err)
				suite.Equal(tc.proposalPasses, proposalPasses)
			} else {
				suite.NotNil(err)
			}
		})
	}
}

func committeeGenState(cdc *codec.Codec, committees []types.Committee, proposals []types.Proposal, votes []types.Vote) app.GenesisState {
	gs := types.NewGenesisState(
		uint64(len(proposals)+1),
		committees,
		proposals,
		votes,
	)
	return app.GenesisState{committee.ModuleName: cdc.MustMarshalJSON(gs)}
}

type UnregisteredProposal struct {
	gov.TextProposal
}

func (UnregisteredProposal) ProposalRoute() string { return "unregistered" }
func (UnregisteredProposal) ProposalType() string  { return "unregistered" }

var _ types.PubProposal = UnregisteredProposal{}

func (suite *KeeperTestSuite) TestValidatePubProposal() {

	testcases := []struct {
		name        string
		pubProposal types.PubProposal
		expectPass  bool
	}{
		{
			name:        "valid",
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			expectPass:  true,
		},
		{
			name:        "invalid (missing title)",
			pubProposal: gov.TextProposal{Description: "A description of this proposal."},
			expectPass:  false,
		},
		{
			name:        "invalid (unregistered)",
			pubProposal: UnregisteredProposal{gov.TextProposal{Title: "A Title", Description: "A description of this proposal."}},
			expectPass:  false,
		},
		{
			name:        "invalid (nil)",
			pubProposal: nil,
			expectPass:  false,
		},
		// TODO test case when the handler fails
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			err := suite.keeper.ValidatePubProposal(suite.ctx, tc.pubProposal)
			if tc.expectPass {
				suite.NoError(err)
			} else {
				suite.NotNil(err)
			}
		})
	}
}
