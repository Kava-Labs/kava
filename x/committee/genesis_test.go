package committee_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app    app.TestApp
	ctx    sdk.Context
	keeper committee.Keeper
}

func (suite *GenesisTestSuite) TestGenesis() {
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
			name: "invalid",
			genState: types.NewGenesisState(
				2,
				[]types.Committee{},
				[]types.Proposal{{ID: 1, CommitteeID: 57}},
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
