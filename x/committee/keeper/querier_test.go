package keeper_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

const (
	custom = "custom"
)

var testTime time.Time = time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)

func NewCommitteeGenesisState(cdc *codec.Codec, gs committee.GenesisState) app.GenesisState {
	return app.GenesisState{committee.ModuleName: cdc.MustMarshalJSON(gs)}
}

type QuerierTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	cdc    *codec.Codec

	querier sdk.Querier

	addresses   []sdk.AccAddress
	testGenesis types.GenesisState
	votes       map[uint64]([]types.Vote)
}

func (suite *QuerierTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, abci.Header{})
	suite.cdc = suite.app.Codec()
	suite.querier = keeper.NewQuerier(suite.keeper)

	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
	suite.testGenesis = types.NewGenesisState(
		3,
		[]types.Committee{
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
		[]types.Proposal{
			{ID: 1, CommitteeID: 1, PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."), Deadline: testTime.Add(7 * 24 * time.Hour)},
			{ID: 2, CommitteeID: 1, PubProposal: gov.NewTextProposal("Another Title", "A description of this other proposal."), Deadline: testTime.Add(21 * 24 * time.Hour)},
		},
		[]types.Vote{
			{ProposalID: 1, Voter: suite.addresses[0]},
			{ProposalID: 1, Voter: suite.addresses[1]},
			{ProposalID: 2, Voter: suite.addresses[2]},
		},
	)
	suite.app.InitializeFromGenesisStates(
		NewCommitteeGenesisState(suite.cdc, suite.testGenesis),
	)

	suite.votes = getProposalVoteMap(suite.keeper, suite.ctx)
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
	suite.Equal(suite.testGenesis.Committees, committees)
}

func (suite *QuerierTestSuite) TestQueryCommittee() {
	ctx := suite.ctx.WithIsCheckTx(false) // ?
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryCommittee}, "/"),
		Data: suite.cdc.MustMarshalJSON(types.NewQueryCommitteeParams(suite.testGenesis.Committees[0].ID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryCommittee}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var committee types.Committee
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &committee))

	// Check
	suite.Equal(suite.testGenesis.Committees[0], committee)
}

func (suite *QuerierTestSuite) TestQueryProposals() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	comID := suite.testGenesis.Proposals[0].CommitteeID
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
	for _, p := range suite.testGenesis.Proposals {
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
		Data: suite.cdc.MustMarshalJSON(types.NewQueryProposalParams(suite.testGenesis.Proposals[0].ID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryProposal}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var proposal types.Proposal
	suite.NoError(suite.cdc.UnmarshalJSON(bz, &proposal))

	// Check
	suite.Equal(suite.testGenesis.Proposals[0], proposal)
}

func (suite *QuerierTestSuite) TestQueryVotes() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	propID := suite.testGenesis.Proposals[0].ID
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
	propID := suite.testGenesis.Proposals[0].ID
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
	propID := suite.testGenesis.Proposals[0].ID
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
