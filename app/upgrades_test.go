package app_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type UpgradeTestSuite struct {
	suite.Suite
	App app.TestApp
	Ctx sdk.Context
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	suite.App = app.NewTestApp()

	cdc := suite.App.AppCodec()

	suite.App = suite.App.InitializeFromGenesisStates(
		app.GenesisState{
			minttypes.ModuleName: cdc.MustMarshalJSON(minttypes.NewGenesisState(
				minttypes.DefaultInitialMinter(),
				// Params reflect mainnet params
				minttypes.Params{
					MintDenom:           "ukava",
					InflationRateChange: sdk.NewDecWithPrec(13, 2),
					InflationMax:        sdk.OneDec(),
					InflationMin:        sdk.OneDec(),
					GoalBonded:          sdk.NewDecWithPrec(67, 2),
					BlocksPerYear:       5256000,
				},
			)),
		},
	)

	suite.Ctx = suite.App.NewContext(false, tmproto.Header{Height: 1})
}

func (suite *UpgradeTestSuite) TestUpdateCosmosMintInflation() {
	mintKeeper := suite.App.GetMintKeeper()
	oldParams := mintKeeper.GetParams(suite.Ctx)
	suite.Equal(sdk.OneDec(), oldParams.InflationMin, "initial InflationMin should be 1")
	suite.Equal(sdk.OneDec(), oldParams.InflationMax, "initial InflationMax should be 1")

	// Run migration
	app.UpdateCosmosMintInflation(suite.Ctx, mintKeeper)

	newParams := mintKeeper.GetParams(suite.Ctx)
	suite.NotEqual(oldParams, newParams, "params should be changed after migration")

	suite.Equal(sdk.MustNewDecFromStr("0.75"), sdk.NewDecWithPrec(75, 2), "sdk.NewDecWithPrec(75, 2) should be 0.75")

	suite.Equal(sdk.MustNewDecFromStr("0.75"), newParams.InflationMin, "InflationMin should changed to 0.75")
	suite.Equal(sdk.MustNewDecFromStr("0.75"), newParams.InflationMax, "InflationMax should changed to 0.75")

	// Other parameters should be unchanged
	suite.Equal(oldParams.MintDenom, newParams.MintDenom)
	suite.Equal(oldParams.InflationRateChange, newParams.InflationRateChange)
	suite.Equal(oldParams.GoalBonded, newParams.GoalBonded)
	suite.Equal(oldParams.BlocksPerYear, newParams.BlocksPerYear)
}

func (suite *UpgradeTestSuite) TestUpdateSavingsParams() {
	savingsKeeper := suite.App.GetSavingsKeeper()
	oldParams := savingsKeeper.GetParams(suite.Ctx)
	suite.Empty(oldParams.SupportedDenoms, "initial SupportedDenoms should be empty")

	// Run migration
	app.UpdateSavingsParams(suite.Ctx, savingsKeeper)

	newParams := savingsKeeper.GetParams(suite.Ctx)
	suite.NotEqual(oldParams, newParams, "params should be changed after migration")

	suite.ElementsMatch(
		[]string{
			"ukava",
			"bkava",
			"erc20/multichain/usdc",
		},
		newParams.SupportedDenoms,
		"SupportedDenoms should be updated to include ukava",
	)
}
