package keeper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

const (
	custom = "custom"
)

type QuerierTestSuite struct {
	suite.Suite

	keeper      keeper.Keeper
	app         app.TestApp
	ctx         sdk.Context
	cdc         codec.Codec
	legacyAmino *codec.LegacyAmino

	querier sdk.Querier

	addresses   []sdk.AccAddress
	testGenesis *types.GenesisState
	votes       map[uint64]([]types.Vote)
}

func (suite *QuerierTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, tmproto.Header{})
	suite.cdc = suite.app.AppCodec()
	suite.legacyAmino = suite.app.LegacyAmino()
	suite.querier = keeper.NewQuerier(suite.keeper, suite.legacyAmino)

	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
	memberCommittee := mustNewTestMemberCommittee(suite.addresses[:3])
	memberCommittee.ID = 1
	noPermCommittee := mustNewTestMemberCommittee(suite.addresses[2:])
	noPermCommittee.ID = 2
	noPermCommittee.SetPermissions([]types.Permission{})
	proposalOne := mustNewTestProposal()
	proposalTwo := mustNewTestProposal()
	proposalTwo.ID = 2
	suite.testGenesis = types.NewGenesisState(
		3,
		[]types.Committee{memberCommittee, noPermCommittee},
		types.Proposals{proposalOne, proposalTwo},
		[]types.Vote{
			{ProposalID: 1, Voter: suite.addresses[0], VoteType: types.VOTE_TYPE_YES},
			{ProposalID: 1, Voter: suite.addresses[1], VoteType: types.VOTE_TYPE_YES},
			{ProposalID: 2, Voter: suite.addresses[2], VoteType: types.VOTE_TYPE_YES},
		},
	)
	genState := NewCommitteeGenesisState(suite.cdc, suite.testGenesis)
	suite.app.InitializeFromGenesisStates(genState)

	suite.votes = getProposalVoteMap(suite.keeper, suite.ctx)
}

func (suite *QuerierTestSuite) assertQuerierResponse(expected interface{}, actual []byte) {
	expectedJson, err := suite.legacyAmino.MarshalJSONIndent(expected, "", "  ")
	suite.Require().NoError(err)
	suite.Require().Equal(string(expectedJson), string(actual))
}

func (suite *QuerierTestSuite) TestQueryCommittees() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryCommittees}, "/"),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryCommittees}, query)
	suite.Require().NoError(err)
	suite.Require().NotNil(bz)
	suite.assertQuerierResponse(suite.testGenesis.GetCommittees(), bz)
}

func (suite *QuerierTestSuite) TestQueryCommittee() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	targetCommittee := suite.testGenesis.GetCommittees()[0]
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryCommittee}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryCommitteeParams(targetCommittee.GetID())),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryCommittee}, query)
	suite.NoError(err)
	suite.NotNil(bz)
	suite.assertQuerierResponse(targetCommittee, bz)
}

func (suite *QuerierTestSuite) TestQueryProposals() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	comID := suite.testGenesis.Proposals[0].CommitteeID
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryProposals}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryCommitteeParams(comID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryProposals}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Check
	expectedProposals := types.Proposals{}
	for _, p := range suite.testGenesis.Proposals {
		if p.CommitteeID == comID {
			expectedProposals = append(expectedProposals, p)
		}
	}
	suite.assertQuerierResponse(expectedProposals, bz)
}

func (suite *QuerierTestSuite) TestQueryProposal() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryProposal}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryProposalParams(suite.testGenesis.Proposals[0].ID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryProposal}, query)
	suite.NoError(err)
	suite.NotNil(bz)
	suite.assertQuerierResponse(suite.testGenesis.Proposals[0], bz)
}

func (suite *QuerierTestSuite) TestQueryNextProposalID() {
	bz, err := suite.querier(suite.ctx, []string{types.QueryNextProposalID}, abci.RequestQuery{})
	suite.Require().NoError(err)
	suite.Require().NotNil(bz)

	var nextProposalID uint64
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(bz, &nextProposalID))

	expectedID, _ := suite.keeper.GetNextProposalID(suite.ctx)
	suite.Require().Equal(expectedID, nextProposalID)
}

func (suite *QuerierTestSuite) TestQueryVotes() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	propID := suite.testGenesis.Proposals[0].ID
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVotes}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryProposalParams(propID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryVotes}, query)
	suite.NoError(err)
	suite.NotNil(bz)
	suite.assertQuerierResponse(suite.votes[propID], bz)
}

func (suite *QuerierTestSuite) TestQueryVote() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	propID := suite.testGenesis.Proposals[0].ID
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVote}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryVoteParams(propID, suite.votes[propID][0].Voter)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryVote}, query)
	suite.NoError(err)
	suite.NotNil(bz)
	suite.assertQuerierResponse(suite.votes[propID][0], bz)
}

func (suite *QuerierTestSuite) TestQueryTally() {

	ctx := suite.ctx.WithIsCheckTx(false)

	// Expected result
	propID := suite.testGenesis.Proposals[0].ID
	expectedPollingStatus := types.QueryTallyResponse{
		ProposalID:    1,
		YesVotes:      sdk.NewDec(int64(len(suite.votes[propID]))),
		NoVotes:       sdk.ZeroDec(),
		CurrentVotes:  sdk.NewDec(int64(len(suite.votes[propID]))),
		PossibleVotes: testutil.D("3.0"),
		VoteThreshold: testutil.D("0.667"),
		Quorum:        testutil.D("0"),
	}

	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryTally}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryProposalParams(propID)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryTally}, query)
	suite.NoError(err)
	suite.NotNil(bz)
	suite.assertQuerierResponse(expectedPollingStatus, bz)
}

type TestSubParam struct {
	Some   string
	Test   sdk.Dec
	Params []types.Vote
}
type TestParams struct {
	TestKey TestSubParam
}

const paramKey = "TestKey"

func (p *TestParams) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair([]byte(paramKey), &p.TestKey, func(interface{}) error { return nil }),
	}
}
func (suite *QuerierTestSuite) TestQueryRawParams() {
	ctx := suite.ctx.WithIsCheckTx(false)

	// Create a new param subspace to avoid adding dependency to another module. Set a test param value.
	subspaceName := "test"
	subspace := suite.app.GetParamsKeeper().Subspace(subspaceName)
	subspace = subspace.WithKeyTable(paramstypes.NewKeyTable().RegisterParamSet(&TestParams{}))

	paramValue := TestSubParam{
		Some: "test",
		Test: testutil.D("1000000000000.000000000000000001"),
		Params: []types.Vote{
			types.NewVote(1, suite.addresses[0], types.VOTE_TYPE_YES),
			types.NewVote(12, suite.addresses[1], types.VOTE_TYPE_YES),
		},
	}
	subspace.Set(ctx, []byte(paramKey), paramValue)

	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryRawParams}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryRawParamsParams(subspaceName, paramKey)),
	}

	// Execute query
	bz, err := suite.querier(ctx, []string{types.QueryRawParams}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes
	var returnedParamValue []byte
	suite.NoError(suite.legacyAmino.UnmarshalJSON(bz, &returnedParamValue))

	// Check
	suite.Equal(suite.legacyAmino.MustMarshalJSON(paramValue), returnedParamValue)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}
