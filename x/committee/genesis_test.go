package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app       app.TestApp
	ctx       sdk.Context
	keeper    committee.Keeper
	addresses []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, abci.Header{})
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(10)
}

func (suite *GenesisTestSuite) TestInitGenesis() {

	memberCom := types.MemberCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:               1,
			Description:      "This member committee is for testing.",
			Members:          suite.addresses[:2],
			Permissions:      []types.Permission{types.GodPermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      types.FirstPastThePost,
		},
	}

	tokenCom := types.TokenCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:               1,
			Description:      "This token committee is for testing.",
			Members:          suite.addresses[:2],
			Permissions:      []types.Permission{types.GodPermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      types.FirstPastThePost,
		},
		Quorum:     d("0.4"),
		TallyDenom: "hard",
	}

	// Most genesis validation tests are located in the types directory. The 'invalid' test cases are
	// randomly selected subset of those tests.
	testCases := []struct {
		name       string
		genState   types.GenesisState
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
				[]types.Vote{{Voter: suite.addresses[0], ProposalID: 1, VoteType: types.Yes}},
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
			suite.ctx = suite.app.NewContext(true, abci.Header{})

			// Run
			var exportedGenState types.GenesisState
			run := func() {
				committee.InitGenesis(suite.ctx, suite.keeper, tc.genState)
				exportedGenState = committee.ExportGenesis(suite.ctx, suite.keeper)
			}
			if tc.expectPass {
				suite.NotPanics(run)
			} else {
				suite.Panics(run)
			}

			// Check
			if tc.expectPass {
				suite.Equal(tc.genState, exportedGenState)
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
