package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

func d(s string) sdk.Dec { return sdk.MustNewDecFromStr(s) }

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

func (suite *KeeperTestSuite) TestGetSetDeleteCommittee() {
	// setup test
	com := types.Committee{
		ID:                  12,
		Description:         "This committee is for testing.",
		Members:             suite.addresses,
		Permissions:         []types.Permission{types.GodPermission{}},
		VoteThreshold:       d("0.667"),
		MaxProposalDuration: time.Hour * 24 * 7,
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

func (suite *KeeperTestSuite) TestGetSetDeleteProposal() {
	// test setup
	prop := types.Proposal{
		ID:          12,
		CommitteeID: 0,
		PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
		Deadline:    time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
