package keeper_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context

	addresses []sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, abci.Header{})
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *KeeperTestSuite) TestSubmitProposal() {
	normalCom := types.Committee{
		ID:          12,
		Members:     suite.addresses[:2],
		Permissions: []types.Permission{types.GodPermission{}},
	}
	noPermissionsCom := types.Committee{
		ID:          12,
		Members:     suite.addresses[:2],
		Permissions: []types.Permission{},
	}

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
			// Create local testApp because suite doesn't run the SetupTest function for subtests, which would mean the app state is not be reset between subtests.
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
				_, found := keeper.GetProposal(ctx, id)
				suite.True(found)
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

	testcases := []struct {
		name       string
		proposalID uint64
		voter      sdk.AccAddress
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
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests, which would mean the app state is not be reset between subtests.
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{})
			tApp.InitializeFromGenesisStates()

			// setup the committee and proposal
			keeper.SetCommittee(ctx, normalCom)
			_, err := keeper.SubmitProposal(ctx, normalCom.Members[0], normalCom.ID, gov.NewTextProposal("A Title", "A description of this proposal."))
			suite.NoError(err)

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

func (suite *KeeperTestSuite) TestCloseOutProposal() {
	// setup test
	suite.app.InitializeFromGenesisStates()
	// TODO replace below with genesis state
	normalCom := types.Committee{
		ID:          12,
		Members:     suite.addresses[:2],
		Permissions: []types.Permission{types.GodPermission{}},
	}
	suite.keeper.SetCommittee(suite.ctx, normalCom)
	pprop := gov.NewTextProposal("A Title", "A description of this proposal.")
	id, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop)
	suite.NoError(err)
	err = suite.keeper.AddVote(suite.ctx, id, normalCom.Members[0])
	suite.NoError(err)
	err = suite.keeper.AddVote(suite.ctx, id, normalCom.Members[1])
	suite.NoError(err)

	// run test
	err = suite.keeper.CloseOutProposal(suite.ctx, id)

	// check
	suite.NoError(err)
	_, found := suite.keeper.GetProposal(suite.ctx, id)
	suite.False(found)
	suite.keeper.IterateVotes(suite.ctx, id, func(v types.Vote) bool {
		suite.Fail("found vote when none should exist")
		return false
	})

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

func (suite *KeeperTestSuite) TestGetSetDeleteCommittee() {
	// setup test
	com := types.Committee{
		ID: 12,
		// TODO other fields
	}

	// write and read from store
	suite.keeper.SetCommittee(suite.ctx, com)
	readCommittee, found := suite.keeper.GetCommittee(suite.ctx, com.ID)

	// check before and after match
	suite.True(found)
	suite.Equal(com, readCommittee)

	// delete from store
	suite.keeper.DeleteCommittee(suite.ctx, com.ID)

	// check does not exist
	_, found = suite.keeper.GetCommittee(suite.ctx, com.ID)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestGetSetProposal() {
	// test setup
	prop := types.Proposal{
		ID: 12,
		// TODO other fields
	}

	// write and read from store
	suite.keeper.SetProposal(suite.ctx, prop)
	readProposal, found := suite.keeper.GetProposal(suite.ctx, prop.ID)

	// check before and after match
	suite.True(found)
	suite.Equal(prop, readProposal)

	// delete from store
	suite.keeper.DeleteProposal(suite.ctx, prop.ID)

	// check does not exist
	_, found = suite.keeper.GetProposal(suite.ctx, prop.ID)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestGetSetVote() {
	// test setup
	vote := types.Vote{
		ProposalID: 12,
		Voter:      suite.addresses[0],
		// TODO other fields
	}

	// write and read from store
	suite.keeper.SetVote(suite.ctx, vote)
	readVote, found := suite.keeper.GetVote(suite.ctx, vote.ProposalID, vote.Voter)

	// check before and after match
	suite.True(found)
	suite.Equal(vote, readVote)

	// delete from store
	suite.keeper.DeleteVote(suite.ctx, vote.ProposalID, vote.Voter)

	// check does not exist
	_, found = suite.keeper.GetVote(suite.ctx, vote.ProposalID, vote.Voter)
	suite.False(found)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
