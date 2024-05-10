package precisebank_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		setupFn      func()
		genesisState *types.GenesisState
		shouldPanic  bool
		panicMsg     string
	}{
		{
			"default genesisState",
			func() {},
			types.DefaultGenesisState(),
			false,
			"",
		},
		{
			"empty genesisState",
			func() {},
			&types.GenesisState{},
			true,
			"failed to validate precisebank genesis state: nil remainder amount",
		},
		{
			// Other GenesisState.Validate() tests are in types/genesis_test.go
			"invalid genesisState - GenesisState.Validate() is called",
			func() {},
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
				},
				sdkmath.ZeroInt(),
			),
			true,
			"failed to validate precisebank genesis state: invalid balances: duplicate address kava1qy0xn7za",
		},
		{
			"sets module account",
			func() {
				// Delete the module account first to ensure it's created here
				moduleAcc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
				suite.AccountKeeper.RemoveAccount(suite.Ctx, moduleAcc)

				// Ensure module account is deleted in state.
				// GetModuleAccount() will always return non-nil and does not
				// necessarily equate to the account being stored in the account store.
				suite.Require().Nil(suite.AccountKeeper.GetAccount(suite.Ctx, moduleAcc.GetAddress()))
			},
			types.DefaultGenesisState(),
			false,
			"",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.shouldPanic {
				suite.Require().PanicsWithValue(
					tc.panicMsg,
					func() {
						precisebank.InitGenesis(
							suite.Ctx,
							suite.Keeper,
							suite.AccountKeeper,
							suite.BankKeeper,
							tc.genesisState,
						)
					},
				)

				return
			}

			suite.Require().NotPanics(func() {
				precisebank.InitGenesis(
					suite.Ctx,
					suite.Keeper,
					suite.AccountKeeper,
					suite.BankKeeper,
					tc.genesisState,
				)
			})

			// Ensure module account is created
			moduleAcc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
			suite.NotNil(moduleAcc)
			suite.NotNil(
				suite.AccountKeeper.GetAccount(suite.Ctx, moduleAcc.GetAddress()),
				"module account should be created & stored in account store",
			)

			// TODO: Check module state once implemented

			// Verify balances
			// IterateBalances() or something

			// Ensure reserve balance matches sum of all fractional balances
			// sum up IterateBalances()

			// - etc
		})
	}
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
					suite.BankKeeper,
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
					suite.BankKeeper,
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
