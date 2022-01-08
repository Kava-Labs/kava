package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

type grpcQueryTestSuite struct {
	testutil.Suite
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *grpcQueryTestSuite) TestVote() {
	ctx, keeper, queryClient := suite.Ctx, suite.Keeper, suite.QueryClient
	vote := types.Vote{
		ProposalID: 1,
		Voter:      suite.Addresses[0],
		VoteType:   types.VOTE_TYPE_ABSTAIN,
	}
	keeper.SetVote(ctx, vote)

	req := types.QueryVoteRequest{
		ProposalId: vote.ProposalID,
		Voter:      vote.Voter.String(),
	}
	res, err := queryClient.Vote(context.Background(), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(vote.ProposalID, res.ProposalID)
	suite.Require().Equal(vote.VoteType, res.VoteType)
	suite.Require().Equal(vote.Voter.String(), res.Voter)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
