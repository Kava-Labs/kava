package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app       app.TestApp
	ctx       sdk.Context
	keeper    keeper.Keeper
	addresses []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, tmproto.Header{})
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(10)
}

func (suite *GenesisTestSuite) TestInitGenesis() {

	memberCom := types.MustNewMemberCommittee(
		1,
		"This member committee is for testing.",
		suite.addresses[:2],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_FIRST_PAST_THE_POST,
	)

	tokenCom := types.MustNewTokenCommittee(
		1,
		"This token committee is for testing.",
		suite.addresses[:2],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_FIRST_PAST_THE_POST,
		testutil.D("0.4"),
		"hard",
	)

	// Most genesis validation tests are located in the types directory. The 'invalid' test cases are
	// randomly selected subset of those tests.
	testCases := []struct {
		name       string
		genState   *types.GenesisState
		expectPass bool
	}{
		{
			name:       "normal",
			genState:   types.DefaultGenesisState(),
			expectPass: true,
		},
		{
			name: "member committee is correctly validated",
			genState: types.NewGenesisState(
				1,
				[]types.Committee{memberCom},
				[]types.Proposal{},
				[]types.Vote{},
			),
			expectPass: true,
		},
		{
			name: "token committee is correctly validated",
			genState: types.NewGenesisState(
				1,
				[]types.Committee{tokenCom},
				[]types.Proposal{},
				[]types.Vote{},
			),
			expectPass: true,
		},
		{
			name: "invalid: duplicate committee ID",
			genState: types.NewGenesisState(
				1,
				[]types.Committee{memberCom, memberCom},
				[]types.Proposal{},
				[]types.Vote{},
			),
			expectPass: false,
		},
		{
			name: "invalid: proposal doesn't have committee",
			genState: types.NewGenesisState(
				2,
				[]types.Committee{},
				[]types.Proposal{{ID: 1, CommitteeID: 57}},
				[]types.Vote{},
			),
			expectPass: false,
		},
		{
			name: "invalid: vote doesn't have proposal",
			genState: types.NewGenesisState(
				1,
				[]types.Committee{},
				[]types.Proposal{},
				[]types.Vote{{Voter: suite.addresses[0], ProposalID: 1, VoteType: types.VOTE_TYPE_YES}},
			),
			expectPass: false,
		},
		{
			name: "invalid: next proposal ID isn't greater than proposal ID",
			genState: types.NewGenesisState(
				4,
				[]types.Committee{memberCom},
				[]types.Proposal{{ID: 3, CommitteeID: 1}, {ID: 4, CommitteeID: 1}},
				[]types.Vote{},
			),
			expectPass: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup (note: suite.SetupTest is not run before every suite.Run)
			suite.app = app.NewTestApp()
			suite.keeper = suite.app.GetCommitteeKeeper()
			suite.ctx = suite.app.NewContext(true, tmproto.Header{})

			// Run
			var exportedGenState *types.GenesisState
			run := func() {
				committee.InitGenesis(suite.ctx, suite.keeper, tc.genState)
				exportedGenState = committee.ExportGenesis(suite.ctx, suite.keeper)
			}
			if tc.expectPass {
				suite.Require().NotPanics(run)
			} else {
				suite.Require().Panics(run)
			}

			// Check
			if tc.expectPass {
				expectedJson, err := suite.app.AppCodec().MarshalJSON(tc.genState)
				suite.Require().NoError(err)
				actualJson, err := suite.app.AppCodec().MarshalJSON(exportedGenState)
				suite.Equal(expectedJson, actualJson)
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
