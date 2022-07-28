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
		types.VaultRecords{
			{
				TotalShares: types.NewVaultShare("", sdk.NewInt(1)),
			},
		},
		types.VaultShareRecords{},
	)

	suite.Panics(func() {
		earn.InitGenesis(suite.Ctx, suite.Keeper, suite.AccountKeeper, invalidState)
	}, "expected init genesis to panic with invalid state")
}

func (suite *genesisTestSuite) Test_InitAndExportGenesis() {
	depositor_1, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	suite.Require().NoError(err)
	depositor_2, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
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
				TotalShares: types.NewVaultShare("ukava", sdk.NewInt(2000000)),
			},
			types.VaultRecord{
				TotalShares: types.NewVaultShare("usdx", sdk.NewInt(2000000)),
			},
		},
		types.VaultShareRecords{
			types.VaultShareRecord{
				Depositor: depositor_1,
				Shares: types.NewVaultShares(
					types.NewVaultShare("usdx", sdk.NewInt(500000)),
					types.NewVaultShare("ukava", sdk.NewInt(1900000)),
				),
			},
			types.VaultShareRecord{
				Depositor: depositor_2,
				Shares: types.NewVaultShares(
					types.NewVaultShare("usdx", sdk.NewInt(500000)),
					types.NewVaultShare("ukava", sdk.NewInt(1900000)),
				),
			},
		},
	)

	earn.InitGenesis(suite.Ctx, suite.Keeper, suite.AccountKeeper, state)
	suite.Equal(state.Params, suite.Keeper.GetParams(suite.Ctx))

	vaultRecord1, _ := suite.Keeper.GetVaultRecord(suite.Ctx, "ukava")
	vaultRecord2, _ := suite.Keeper.GetVaultRecord(suite.Ctx, "usdx")
	suite.Equal(state.VaultRecords[0], vaultRecord1)
	suite.Equal(state.VaultRecords[1], vaultRecord2)

	shareRecord1, _ := suite.Keeper.GetVaultShareRecord(suite.Ctx, depositor_1)
	shareRecord2, _ := suite.Keeper.GetVaultShareRecord(suite.Ctx, depositor_2)

	suite.Equal(state.VaultShareRecords[0], shareRecord1)
	suite.Equal(state.VaultShareRecords[1], shareRecord2)

	exportedState := earn.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(state, exportedState)
}

func (suite *genesisTestSuite) Test_Marshall() {
	depositor_1, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	suite.Require().NoError(err)
	depositor_2, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
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
				TotalShares: types.NewVaultShare("ukava", sdk.NewInt(2000000)),
			},
			types.VaultRecord{
				TotalShares: types.NewVaultShare("usdx", sdk.NewInt(2000000)),
			},
		},
		types.VaultShareRecords{
			types.VaultShareRecord{
				Depositor: depositor_1,
				Shares: types.NewVaultShares(
					types.NewVaultShare("usdx", sdk.NewInt(500000)),
					types.NewVaultShare("ukava", sdk.NewInt(1900000)),
				),
			},
			types.VaultShareRecord{
				Depositor: depositor_2,
				Shares: types.NewVaultShares(
					types.NewVaultShare("usdx", sdk.NewInt(500000)),
					types.NewVaultShare("ukava", sdk.NewInt(1900000)),
				),
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

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
