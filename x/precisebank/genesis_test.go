package precisebank_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank"
	"github.com/kava-labs/kava/x/precisebank/keeper"
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
		panicMsg     string
	}{
		{
			"valid - default genesisState",
			func() {},
			types.DefaultGenesisState(),
			"",
		},
		{
			"valid - empty genesisState",
			func() {},
			&types.GenesisState{},
			"failed to validate precisebank genesis state: nil remainder amount",
		},
		{
			"valid - module balance matches non-zero amount",
			func() {
				// Set module account balance to expected amount
				err := suite.BankKeeper.MintCoins(
					suite.Ctx,
					types.ModuleName,
					sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.NewInt(2))),
				)
				suite.Require().NoError(err)
			},
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), types.ConversionFactor().SubRaw(1)),
					types.NewFractionalBalance(sdk.AccAddress{2}.String(), types.ConversionFactor().SubRaw(1)),
				},
				// 2 leftover from 0.999... + 0.999...
				sdkmath.NewInt(2),
			),
			"",
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
			"failed to validate precisebank genesis state: invalid balances: duplicate address kava1qy0xn7za",
		},
		{
			"invalid - module balance insufficient",
			func() {},
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), types.ConversionFactor().SubRaw(1)),
					types.NewFractionalBalance(sdk.AccAddress{2}.String(), types.ConversionFactor().SubRaw(1)),
				},
				// 2 leftover from 0.999... + 0.999...
				sdkmath.NewInt(2),
			),
			"module account balance does not match sum of fractional balances and remainder, balance is 0ukava but expected 2000000000000akava (2ukava)",
		},
		{
			"invalid - module balance excessive",
			func() {
				// Set module account balance to greater than expected amount
				err := suite.BankKeeper.MintCoins(
					suite.Ctx,
					types.ModuleName,
					sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.NewInt(100))),
				)
				suite.Require().NoError(err)
			},
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), types.ConversionFactor().SubRaw(1)),
					types.NewFractionalBalance(sdk.AccAddress{2}.String(), types.ConversionFactor().SubRaw(1)),
				},
				sdkmath.NewInt(2),
			),
			"module account balance does not match sum of fractional balances and remainder, balance is 100ukava but expected 2000000000000akava (2ukava)",
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
			"",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setupFn()

			if tc.panicMsg != "" {
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

			// Verify balances are set in state, get full list of balances in
			// state to ensure they are set AND no extra balances are set
			var bals []types.FractionalBalance
			suite.Keeper.IterateFractionalBalances(suite.Ctx, func(addr sdk.AccAddress, bal sdkmath.Int) bool {
				bals = append(bals, types.NewFractionalBalance(addr.String(), bal))

				return false
			})

			suite.Require().ElementsMatch(tc.genesisState.Balances, bals, "balances should be set in state")

			remainder := suite.Keeper.GetRemainderAmount(suite.Ctx)
			suite.Require().Equal(tc.genesisState.Remainder, remainder, "remainder should be set in state")

			// Additional verification of state via invariants
			invariantFn := keeper.AllInvariants(suite.Keeper)
			msg, broken := invariantFn(suite.Ctx)
			suite.Require().False(broken, "invariants should not be broken after InitGenesis")
			suite.Require().Empty(msg, "invariants should not return a message after InitGenesis")
		})
	}
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	// ExportGenesis(InitGenesis(genesisState)) == genesisState
	// Must also be valid.

	tests := []struct {
		name             string
		initGenesisState func() *types.GenesisState
	}{
		{
			"InitGenesis(DefaultGenesisState)",
			func() *types.GenesisState {
				return types.DefaultGenesisState()
			},
		},
		{
			"balances, no remainder",
			func() *types.GenesisState {
				err := suite.BankKeeper.MintCoins(
					suite.Ctx,
					types.ModuleName,
					sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.NewInt(1))),
				)
				suite.Require().NoError(err)

				return types.NewGenesisState(
					types.FractionalBalances{
						types.NewFractionalBalance(sdk.AccAddress{1}.String(), types.ConversionFactor().QuoRaw(2)),
						types.NewFractionalBalance(sdk.AccAddress{2}.String(), types.ConversionFactor().QuoRaw(2)),
					},
					sdkmath.ZeroInt(),
				)
			},
		},
		{
			"balances, remainder",
			func() *types.GenesisState {
				err := suite.BankKeeper.MintCoins(
					suite.Ctx,
					types.ModuleName,
					sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdkmath.NewInt(1))),
				)
				suite.Require().NoError(err)

				return types.NewGenesisState(
					types.FractionalBalances{
						types.NewFractionalBalance(sdk.AccAddress{1}.String(), types.ConversionFactor().QuoRaw(2)),
						types.NewFractionalBalance(sdk.AccAddress{2}.String(), types.ConversionFactor().QuoRaw(2).SubRaw(1)),
					},
					sdkmath.OneInt(),
				)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Reset state
			suite.SetupTest()

			initGs := tc.initGenesisState()

			suite.Require().NotPanics(func() {
				precisebank.InitGenesis(
					suite.Ctx,
					suite.Keeper,
					suite.AccountKeeper,
					suite.BankKeeper,
					initGs,
				)
			})

			genesisState := precisebank.ExportGenesis(suite.Ctx, suite.Keeper)
			suite.Require().NoError(genesisState.Validate(), "exported genesis state should be valid")

			suite.Require().Equal(
				initGs,
				genesisState,
				"exported genesis state should equal initial genesis state",
			)
		})
	}
}
