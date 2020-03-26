package keeper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

const (
	custom = "custom"
)

type QuerierTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	cdc    *codec.Codec

	querier sdk.Querier

	addresses  []sdk.AccAddress
	committees []types.Committee
	proposals  []types.Proposal
	votes      map[uint64]([]types.Vote)
}

func (suite *QuerierTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, abci.Header{})
	suite.cdc = suite.app.Codec()
	suite.querier = keeper.NewQuerier(suite.keeper)

	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
	suite.app.InitializeFromGenesisStates()
	// TODO replace below with genesis state
	normalCom := types.Committee{
		ID:          12,
		Members:     suite.addresses[:2],
		Permissions: []types.Permission{types.GodPermission{}},
	}
	suite.keeper.SetCommittee(suite.ctx, normalCom)

	pprop1 := gov.NewTextProposal("1A Title", "A description of this proposal.")
	id1, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop1)
	suite.NoError(err)

	pprop2 := gov.NewTextProposal("2A Title", "A description of this proposal.")
	id2, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop2)
	suite.NoError(err)

	err = suite.keeper.AddVote(suite.ctx, id1, normalCom.Members[0])
	suite.NoError(err)
	err = suite.keeper.AddVote(suite.ctx, id1, normalCom.Members[1])
	suite.NoError(err)
	err = suite.keeper.AddVote(suite.ctx, id2, normalCom.Members[1])
	suite.NoError(err)

	suite.committees = []types.Committee{}
	suite.committees = []types.Committee{normalCom} // TODO
	suite.proposals = []types.Proposal{}
	suite.keeper.IterateProposals(suite.ctx, func(p types.Proposal) bool {
		suite.proposals = append(suite.proposals, p)
		return false
	})
	suite.votes = map[uint64]([]types.Vote){}
	suite.keeper.IterateProposals(suite.ctx, func(p types.Proposal) bool {
		suite.keeper.IterateVotes(suite.ctx, p.ID, func(v types.Vote) bool {
			suite.votes[p.ID] = append(suite.votes[p.ID], v)
			return false
		})
		return false
	})
}

func (suite *QuerierTestSuite) TestQueryCommittees() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryCommittees}, "/"),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryCommittees}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var committees []types.Committee
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &committees))

	// Check
	suite.Equal(suite.committees, committees)
}

func (suite *QuerierTestSuite) TestQueryCommittee() {
	ctx := suite.ctx.WithIsCheckTx(false) // ?
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryCommittee}, "/"),
		Data: suite.cdc.MustMarshalJSON(types.NewQueryCommitteeParams(suite.committees[0].ID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryCommittee}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var committee types.Committee
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &committee))

	// Check
	suite.Equal(suite.committees[0], committee)
}

func (suite *QuerierTestSuite) TestQueryProposals() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	comID := suite.proposals[0].CommitteeID
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryProposals}, "/"),
		Data: suite.cdc.MustMarshalJSON(types.NewQueryCommitteeParams(comID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryProposals}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var proposals []types.Proposal
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &proposals))

	// Check
	expectedProposals := []types.Proposal{}
	for _, p := range suite.proposals {
		if p.CommitteeID == comID {
			expectedProposals = append(expectedProposals, p)
		}
	}
	suite.Equal(expectedProposals, proposals)
}

func (suite *QuerierTestSuite) TestQueryProposal() {
	ctx := suite.ctx.WithIsCheckTx(false) // ?
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryProposal}, "/"),
		Data: suite.cdc.MustMarshalJSON(types.NewQueryProposalParams(suite.proposals[0].ID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryProposal}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var proposal types.Proposal
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &proposal))

	// Check
	suite.Equal(suite.proposals[0], proposal)
}

func (suite *QuerierTestSuite) TestQueryVotes() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	propID := suite.proposals[0].ID
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVotes}, "/"),
		Data: suite.cdc.MustMarshalJSON(types.NewQueryProposalParams(propID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryVotes}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var votes []types.Vote
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &votes))

	// Check
	suite.Equal(suite.votes[propID], votes)
}

func (suite *QuerierTestSuite) TestQueryVote() {
	ctx := suite.ctx.WithIsCheckTx(false) // ?
	// Set up request query
	propID := suite.proposals[0].ID
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVote}, "/"),
		Data: suite.cdc.MustMarshalJSON(types.NewQueryVoteParams(propID, suite.votes[propID][0].Voter)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryVote}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var vote types.Vote
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &vote))

	// Check
	suite.Equal(suite.votes[propID][0], vote)
}

func (suite *QuerierTestSuite) TestQueryTally() {
	ctx := suite.ctx.WithIsCheckTx(false) // ?
	// Set up request query
	propID := suite.proposals[0].ID
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryTally}, "/"),
		Data: suite.cdc.MustMarshalJSON(types.NewQueryProposalParams(propID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryTally}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var tally int64
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &tally))

	// Check
	suite.Equal(int64(len(suite.votes[propID])), tally)
}
func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}
