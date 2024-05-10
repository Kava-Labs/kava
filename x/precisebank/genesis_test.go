package precisebank_test

import (
	"testing"

	"github.com/kava-labs/kava/x/precisebank"
	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	testutil.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	tests := []struct {
		name         string
		genesisState *types.GenesisState
		shouldPanic  bool
		panicMsg     string
	}{
		{
			"default genesisState",
			types.DefaultGenesisState(),
			false,
			"",
		},
		{
			"empty genesisState",
			&types.GenesisState{},
			false,
			"",
		},
		{
			"TODO: invalid genesisState",
			&types.GenesisState{},
			false,
			"",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.shouldPanic {
				suite.Require().Panics(func() {
					precisebank.InitGenesis(suite.Ctx, suite.Keeper, suite.AccountKeeper, tc.genesisState)
				}, tc.panicMsg)

				return
			}

			suite.Require().NotPanics(func() {
				precisebank.InitGenesis(suite.Ctx, suite.Keeper, suite.AccountKeeper, tc.genesisState)
			})

			// Ensure module account is created
			moduleAcc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
			suite.NotNil(moduleAcc, "module account should be created")

			// TODO: Check module state once implemented

			// - Verify balances
			// - Ensure reserve account exists
			// - Ensure reserve balance matches sum of all fractional balances
			// - etc
		})
	}
}

func (suite *GenesisTestSuite) TestImportGenesis_ModuleAccountCreated() {
	suite.Require().NotPanics(func() {
		precisebank.InitGenesis(suite.Ctx, suite.Keeper, suite.AccountKeeper, types.DefaultGenesisState())
	})

	moduleAcc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	suite.NotNil(moduleAcc, "module account should be created")
}

func (suite *GenesisTestSuite) TestExportGenesis_Valid() {
	// ExportGenesis(moduleState) should return a valid genesis state

	tests := []struct {
		name    string
		maleate func()
	}{
		{
			"InitGenesis(DefaultGenesisState)",
			func() {
				precisebank.InitGenesis(
					suite.Ctx,
					suite.Keeper,
					suite.AccountKeeper,
					types.DefaultGenesisState(),
				)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.maleate()

			genesisState := precisebank.ExportGenesis(suite.Ctx, suite.Keeper)

			suite.Require().NoError(genesisState.Validate(), "exported genesis state should be valid")
		})
	}
}

func (suite *GenesisTestSuite) TestExportImportedState() {
	// ExportGenesis(InitGenesis(genesisState)) == genesisState

	tests := []struct {
		name             string
		initGenesisState *types.GenesisState
	}{
		{
			"InitGenesis(DefaultGenesisState)",
			types.DefaultGenesisState(),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Require().NotPanics(func() {
				precisebank.InitGenesis(
					suite.Ctx,
					suite.Keeper,
					suite.AccountKeeper,
					tc.initGenesisState,
				)
			})

			genesisState := precisebank.ExportGenesis(suite.Ctx, suite.Keeper)
			suite.Require().NoError(genesisState.Validate(), "exported genesis state should be valid")

			suite.Require().Equal(
				tc.initGenesisState,
				genesisState,
				"exported genesis state should equal initial genesis state",
			)
		})
	}
}
