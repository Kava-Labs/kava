package keeper_test

import (
	"testing"
	"time"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

type keeperTestSuite struct {
	testutil.Suite
}

func (suite *keeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *keeperTestSuite) TestGetSetDeleteCommittee() {
	cdc := suite.App.AppCodec()

	// setup test
	com := mustNewTestMemberCommittee(suite.Addresses)

	// write and read from store
	suite.Keeper.SetCommittee(suite.Ctx, com)
	readCommittee, found := suite.Keeper.GetCommittee(suite.Ctx, com.ID)

	// check before and after match
	suite.Require().True(found)
	expectedJson, err := cdc.MarshalJSON(com)
	suite.Require().NoError(err)
	actualJson, err := cdc.MarshalJSON(readCommittee)
	suite.Require().NoError(err)
	suite.Equal(expectedJson, actualJson)
	suite.Require().Equal(com.GetPermissions(), readCommittee.GetPermissions())

	// delete from store
	suite.Keeper.DeleteCommittee(suite.Ctx, com.ID)

	// check does not exist
	_, found = suite.Keeper.GetCommittee(suite.Ctx, com.ID)
	suite.Require().False(found)
}

func (suite *keeperTestSuite) TestGetSetDeleteProposal() {
	// test setup
	prop, err := types.NewProposal(
		govv1beta1.NewTextProposal("A Title", "A description of this proposal."),
		12,
		0,
		time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.Require().NoError(err)

	// write and read from store
	suite.Keeper.SetProposal(suite.Ctx, prop)
	readProposal, found := suite.Keeper.GetProposal(suite.Ctx, prop.ID)

	// check before and after match
	suite.True(found)
	suite.Equal(prop, readProposal)

	// delete from store
	suite.Keeper.DeleteProposal(suite.Ctx, prop.ID)

	// check does not exist
	_, found = suite.Keeper.GetProposal(suite.Ctx, prop.ID)
	suite.False(found)
}

func (suite *keeperTestSuite) TestGetSetDeleteVote() {
	// test setup
	vote := types.Vote{
		ProposalID: 12,
		Voter:      suite.Addresses[0],
	}

	// write and read from store
	suite.Keeper.SetVote(suite.Ctx, vote)
	readVote, found := suite.Keeper.GetVote(suite.Ctx, vote.ProposalID, vote.Voter)

	// check before and after match
	suite.True(found)
	suite.Equal(vote, readVote)

	// delete from store
	suite.Keeper.DeleteVote(suite.Ctx, vote.ProposalID, vote.Voter)

	// check does not exist
	_, found = suite.Keeper.GetVote(suite.Ctx, vote.ProposalID, vote.Voter)
	suite.False(found)
}

func (suite *keeperTestSuite) TestGetCommittees() {
	committeesCount := 10
	for i := 0; i < committeesCount; i++ {
		com := mustNewTestMemberCommittee(suite.Addresses)
		com.ID = uint64(i)
		suite.Keeper.SetCommittee(suite.Ctx, com)
	}
	committees := suite.Keeper.GetCommittees(suite.Ctx)
	suite.Require().Len(committees, committeesCount)
}

func (suite *keeperTestSuite) TestGetAndSetProposal() {
	proposal := mustNewTestProposal()

	// Get no proposal
	actualProposal, found := suite.Keeper.GetProposal(suite.Ctx, proposal.ID)
	suite.Require().False(found)
	suite.Require().Equal(types.Proposal{}, actualProposal)

	// Set and get new proposal
	suite.Keeper.SetProposal(suite.Ctx, proposal)
	actualProposal, found = suite.Keeper.GetProposal(suite.Ctx, proposal.ID)
	suite.Require().True(found)
	suite.Require().Equal(proposal, actualProposal)
}

func (suite *keeperTestSuite) TestGetProposalsByCommittee() {
	committee := mustNewTestMemberCommittee(suite.Addresses)
	proposalsCount := 4
	for i := 0; i < proposalsCount; i++ {
		proposal := mustNewTestProposal()
		proposal.ID = uint64(i)
		proposal.CommitteeID = committee.ID
		suite.Keeper.SetProposal(suite.Ctx, proposal)
	}
	proposal := mustNewTestProposal()
	proposal.ID = uint64(proposalsCount)
	proposal.CommitteeID = committee.ID + 1
	suite.Keeper.SetProposal(suite.Ctx, proposal)

	// No proposals
	actualProposals := suite.Keeper.GetProposalsByCommittee(suite.Ctx, committee.ID+2)
	suite.Require().Len(actualProposals, 0)

	// Proposals for existing committees
	actualProposals = suite.Keeper.GetProposalsByCommittee(suite.Ctx, committee.ID)
	suite.Require().Len(actualProposals, proposalsCount)
	actualProposals = suite.Keeper.GetProposalsByCommittee(suite.Ctx, committee.ID+1)
	suite.Require().Len(actualProposals, 1)

	// Make sure proposals have expected data
	suite.Require().Equal(proposal, actualProposals[0])
}

func (suite *keeperTestSuite) TestGetVotesByProposal() {
	proposal := mustNewTestProposal()
	suite.Keeper.SetProposal(suite.Ctx, proposal)
	votes := []types.Vote{
		types.NewVote(proposal.ID, suite.Addresses[0], types.VOTE_TYPE_NO),
		types.NewVote(proposal.ID, suite.Addresses[1], types.VOTE_TYPE_ABSTAIN),
		types.NewVote(proposal.ID, suite.Addresses[1], types.VOTE_TYPE_YES),
	}
	expectedVotes := []types.Vote{votes[0], votes[2]}
	for _, vote := range votes {
		suite.Keeper.SetVote(suite.Ctx, vote)
	}
	actualVotes := suite.Keeper.GetVotesByProposal(suite.Ctx, proposal.ID)
	suite.Require().Len(actualVotes, len(expectedVotes))
	suite.Require().ElementsMatch(expectedVotes, actualVotes)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}
