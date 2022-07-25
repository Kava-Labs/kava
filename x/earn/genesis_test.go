package earn_test

import (
	"testing"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/earn"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type genesisTestSuite struct {
	testutil.Suite
}

func (suite *genesisTestSuite) Test_InitGenesis_ValidationPanic() {
	invalidState := types.NewGenesisState(
		types.Params{
			AllowedVaults: types.AllowedVaults{
				types.NewAllowedVault("usdx", types.STRATEGY_TYPE_HARD),
			},
		},
		types.VaultRecords{},
		types.VaultShareRecords{},
	)

	suite.Panics(func() {
		earn.InitGenesis(suite.Ctx, suite.Keeper, suite.AccountKeeper, invalidState)
	}, "expected init genesis to panic with invalid state")
}

func (suite *genesisTestSuite) Test_InitAndExportGenesis() {
	depositor_1, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	suite.Require().NoError(err)
	depositor_2, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	suite.Require().NoError(err)

	// slices are sorted by key as stored in the data store, so init and export can be compared with equal
	state := types.NewGenesisState(
		types.Params{
			AllowedVaults: types.AllowedVaults{
				types.NewAllowedVault("usdx", types.STRATEGY_TYPE_HARD),
				types.NewAllowedVault("ukava", types.STRATEGY_TYPE_SAVINGS),
			},
		},
		types.VaultRecords{
			types.VaultRecord{
				Denom:       "usdx",
				TotalSupply: sdk.NewInt64Coin("usdx", 1000000),
			},
			types.VaultRecord{
				Denom:       "ukava",
				TotalSupply: sdk.NewInt64Coin("ukava", 2000000),
			},
		},
		types.VaultShareRecords{
			types.VaultShareRecord{
				Depositor:      depositor_1,
				AmountSupplied: sdk.NewCoins(sdk.NewInt64Coin("usdx", 500000), sdk.NewInt64Coin("ukava", 1900000)),
			},
			types.VaultShareRecord{
				Depositor:      depositor_2,
				AmountSupplied: sdk.NewCoins(sdk.NewInt64Coin("usdx", 500000), sdk.NewInt64Coin("ukava", 100000)),
			},
		},
	)

	earn.InitGenesis(suite.Ctx, suite.Keeper, suite.AccountKeeper, state)
	suite.Equal(state.Params, suite.Keeper.GetParams(suite.Ctx))

	vaultRecord1, _ := suite.Keeper.GetVaultRecord(suite.Ctx, "usdx")
	suite.Equal(state.VaultRecords[0], vaultRecord1)
	vaultRecord2, _ := suite.Keeper.GetVaultRecord(suite.Ctx, "kava")
	suite.Equal(state.VaultRecords[1], vaultRecord2)

	shareRecord1, _ := suite.Keeper.GetVaultShareRecord(suite.Ctx, depositor_2)
	suite.Equal(state.VaultShareRecords[0], shareRecord1)
	shareRecord2, _ := suite.Keeper.GetVaultShareRecord(suite.Ctx, depositor_1)
	suite.Equal(state.VaultShareRecords[2], shareRecord2)

	exportedState := earn.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(state, exportedState)
}

func (suite *genesisTestSuite) Test_Marshall() {
	depositor_1, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	suite.Require().NoError(err)
	depositor_2, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	suite.Require().NoError(err)

	// slices are sorted by key as stored in the data store, so init and export can be compared with equal
	state := types.NewGenesisState(
		types.Params{
			AllowedVaults: types.AllowedVaults{
				types.NewAllowedVault("usdx", types.STRATEGY_TYPE_HARD),
				types.NewAllowedVault("ukava", types.STRATEGY_TYPE_SAVINGS),
			},
		},
		types.VaultRecords{
			types.VaultRecord{
				Denom:       "usdx",
				TotalSupply: sdk.NewInt64Coin("usdx", 1000000),
			},
			types.VaultRecord{
				Denom:       "ukava",
				TotalSupply: sdk.NewInt64Coin("ukava", 2000000),
			},
		},
		types.VaultShareRecords{
			types.VaultShareRecord{
				Depositor:      depositor_1,
				AmountSupplied: sdk.NewCoins(sdk.NewInt64Coin("usdx", 500000), sdk.NewInt64Coin("ukava", 1900000)),
			},
			types.VaultShareRecord{
				Depositor:      depositor_2,
				AmountSupplied: sdk.NewCoins(sdk.NewInt64Coin("usdx", 500000), sdk.NewInt64Coin("ukava", 100000)),
			},
		},
	)

	encodingCfg := app.MakeEncodingConfig()
	cdc := encodingCfg.Marshaler

	bz, err := cdc.Marshal(&state)
	suite.Require().NoError(err, "expected genesis state to marshal without error")

	var decodedState types.GenesisState
	err = cdc.Unmarshal(bz, &decodedState)
	suite.Require().NoError(err, "expected genesis state to unmarshal without error")

	suite.Equal(state, decodedState)
}

func (suite *genesisTestSuite) Test_LegacyJSONConversion() {
	depositor_1, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	suite.Require().NoError(err)
	depositor_2, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	suite.Require().NoError(err)

	// slices are sorted by key as stored in the data store, so init and export can be compared with equal
	state := types.NewGenesisState(
		types.Params{
			AllowedVaults: types.AllowedVaults{
				types.NewAllowedVault("usdx", types.STRATEGY_TYPE_HARD),
				types.NewAllowedVault("ukava", types.STRATEGY_TYPE_SAVINGS),
			},
		},
		types.VaultRecords{
			types.VaultRecord{
				Denom:       "usdx",
				TotalSupply: sdk.NewInt64Coin("usdx", 1000000),
			},
			types.VaultRecord{
				Denom:       "ukava",
				TotalSupply: sdk.NewInt64Coin("ukava", 2000000),
			},
		},
		types.VaultShareRecords{
			types.VaultShareRecord{
				Depositor:      depositor_1,
				AmountSupplied: sdk.NewCoins(sdk.NewInt64Coin("usdx", 500000), sdk.NewInt64Coin("ukava", 1900000)),
			},
			types.VaultShareRecord{
				Depositor:      depositor_2,
				AmountSupplied: sdk.NewCoins(sdk.NewInt64Coin("usdx", 500000), sdk.NewInt64Coin("ukava", 100000)),
			},
		},
	)

	encodingCfg := app.MakeEncodingConfig()
	cdc := encodingCfg.Marshaler
	legacyCdc := encodingCfg.Amino

	protoJson, err := cdc.MarshalJSON(&state)
	suite.Require().NoError(err, "expected genesis state to marshal amino json without error")

	aminoJson, err := legacyCdc.MarshalJSON(&state)
	suite.Require().NoError(err, "expected genesis state to marshal amino json without error")

	suite.JSONEq(string(protoJson), string(aminoJson), "expected json outputs to be equal")

	var importedState types.GenesisState
	err = cdc.UnmarshalJSON(aminoJson, &importedState)
	suite.Require().NoError(err, "expected amino json to unmarshall to proto without error")

	suite.Equal(state, importedState, "expected genesis state to be equal")
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
