package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper
	app        app.TestApp
	ctx        sdk.Context

	addresses []sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.bankKeeper = suite.app.GetBankKeeper()
	suite.ctx = suite.app.NewContext(true, tmproto.Header{})
	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	suite.addresses = accAddresses
}

func (suite *KeeperTestSuite) TestGetSetDeleteCommittee() {
	// setup test
	com := mustNewTestMemberCommittee(suite.addresses)

	// write and read from store
	suite.keeper.SetCommittee(suite.ctx, com)
	readCommittee, found := suite.keeper.GetCommittee(suite.ctx, com.ID)

	// check before and after match
	suite.Require().True(found)
	expectedJson, err := suite.app.AppCodec().MarshalJSON(com)
	suite.Require().NoError(err)
	actualJson, err := suite.app.AppCodec().MarshalJSON(readCommittee)
	suite.Require().NoError(err)
	suite.Equal(expectedJson, actualJson)
	suite.Require().Equal(com.GetPermissions(), readCommittee.GetPermissions())

	// delete from store
	suite.keeper.DeleteCommittee(suite.ctx, com.ID)

	// check does not exist
	_, found = suite.keeper.GetCommittee(suite.ctx, com.ID)
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestGetSetDeleteProposal() {
	// test setup
	prop, err := types.NewProposal(
		govtypes.NewTextProposal("A Title", "A description of this proposal."),
		12,
		0,
		time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.Require().NoError(err)

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

func (suite *KeeperTestSuite) TestGetSetDeleteVote() {
	// test setup
	vote := types.Vote{
		ProposalID: 12,
		Voter:      suite.addresses[0],
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

func (suite *KeeperTestSuite) TestGetCommittees() {
	committeesCount := 10
	for i := 0; i < committeesCount; i++ {
		com := mustNewTestMemberCommittee(suite.addresses)
		com.ID = uint64(i)
		suite.keeper.SetCommittee(suite.ctx, com)
	}
	committees := suite.keeper.GetCommittees(suite.ctx)
	suite.Require().Len(committees, committeesCount)
}

func (suite *KeeperTestSuite) TestGetAndSetProposal() {
	proposal := mustNewTestProposal()

	// Get no proposal
	actualProposal, found := suite.keeper.GetProposal(suite.ctx, proposal.ID)
	suite.Require().False(found)
	suite.Require().Equal(types.Proposal{}, actualProposal)

	// Set and get new proposal
	suite.keeper.SetProposal(suite.ctx, proposal)
	actualProposal, found = suite.keeper.GetProposal(suite.ctx, proposal.ID)
	suite.Require().True(found)
	suite.Require().Equal(proposal, actualProposal)
}

func (suite *KeeperTestSuite) TestGetProposalsByCommittee() {
	committee := mustNewTestMemberCommittee(suite.addresses)
	proposalsCount := 4
	for i := 0; i < proposalsCount; i++ {
		proposal := mustNewTestProposal()
		proposal.ID = uint64(i)
		proposal.CommitteeID = committee.ID
		suite.keeper.SetProposal(suite.ctx, proposal)
	}
	proposal := mustNewTestProposal()
	proposal.ID = uint64(proposalsCount)
	proposal.CommitteeID = committee.ID + 1
	suite.keeper.SetProposal(suite.ctx, proposal)

	// No proposals
	actualProposals := suite.keeper.GetProposalsByCommittee(suite.ctx, committee.ID+2)
	suite.Require().Len(actualProposals, 0)

	// Proposals for existing committees
	actualProposals = suite.keeper.GetProposalsByCommittee(suite.ctx, committee.ID)
	suite.Require().Len(actualProposals, proposalsCount)
	actualProposals = suite.keeper.GetProposalsByCommittee(suite.ctx, committee.ID+1)
	suite.Require().Len(actualProposals, 1)

	// Make sure proposals have expected data
	suite.Require().Equal(proposal, actualProposals[0])
}

func (suite *KeeperTestSuite) TestGetVotesByProposal() {
	proposal := mustNewTestProposal()
	suite.keeper.SetProposal(suite.ctx, proposal)
	votes := []types.Vote{
		types.NewVote(proposal.ID, suite.addresses[0], types.VOTE_TYPE_NO),
		types.NewVote(proposal.ID, suite.addresses[1], types.VOTE_TYPE_ABSTAIN),
		types.NewVote(proposal.ID, suite.addresses[1], types.VOTE_TYPE_YES),
	}
	expectedVotes := []types.Vote{votes[0], votes[2]}
	for _, vote := range votes {
		suite.keeper.SetVote(suite.ctx, vote)
	}
	actualVotes := suite.keeper.GetVotesByProposal(suite.ctx, proposal.ID)
	suite.Require().Len(actualVotes, len(expectedVotes))
	suite.Require().ElementsMatch(expectedVotes, actualVotes)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
