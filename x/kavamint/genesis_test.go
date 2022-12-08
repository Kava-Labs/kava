package kavamint_test

import (
	"testing"
	"time"

	"github.com/kava-labs/kava/x/kavamint"
	"github.com/kava-labs/kava/x/kavamint/testutil"
	"github.com/kava-labs/kava/x/kavamint/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type genesisTestSuite struct {
	testutil.KavamintTestSuite
}

func (suite *genesisTestSuite) Test_InitGenesis_ValidationPanic() {
	invalidState := types.NewGenesisState(
		types.NewParams(sdk.OneDec(), types.MaxMintingRate.Add(sdk.OneDec())), // rate over max
		time.Time{},
	)

	suite.Panics(func() {
		kavamint.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), invalidState)
	}, "expected init genesis to panic with invalid state")
}

func (suite *genesisTestSuite) Test_InitGenesis_ModuleAccountDoesNotHaveMinterPerms() {
	gs := types.DefaultGenesisState()

	ak := suite.App.GetAccountKeeper()
	macc := ak.GetModuleAccount(suite.Ctx, types.ModuleName)
	suite.Require().NotNil(macc)

	m, ok := macc.(*authtypes.ModuleAccount)
	suite.Require().True(ok)

	m.Permissions = []string{}
	ak.SetAccount(suite.Ctx, m)

	suite.PanicsWithValue("kavamint module account does not have minter permissions", func() {
		kavamint.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), gs)
	})
}

func (suite *genesisTestSuite) Test_InitGenesis_CreatesModuleAccountWithPermissions() {
	gs := types.DefaultGenesisState()
	ak := suite.App.GetAccountKeeper()

	kavamint.InitGenesis(suite.Ctx, suite.Keeper, ak, gs)

	// by pass auto creation of module accounts
	addr, _ := ak.GetModuleAddressAndPermissions(types.ModuleName)
	acc := suite.App.GetAccountKeeper().GetAccount(suite.Ctx, addr)
	suite.Require().NotNil(acc)

	macc, ok := acc.(authtypes.ModuleAccountI)
	suite.Require().True(ok)

	suite.True(macc.HasPermission(authtypes.Minter))
}

func (suite *genesisTestSuite) Test_InitAndExportGenesis_DefaultValues() {
	state := types.DefaultGenesisState()

	kavamint.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), state)

	suite.Equal(types.DefaultParams(), suite.Keeper.GetParams(suite.Ctx), "expected default params to bet set in store")

	storeTime := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
	suite.Equal(types.DefaultPreviousBlockTime, storeTime, "expected default previous block time to be set in store")

	exportedState := kavamint.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(state, exportedState, "expected exported state to match imported state")
}

func (suite *genesisTestSuite) Test_InitAndExportGenesis_SetValues() {
	prevBlockTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	state := types.NewGenesisState(
		types.NewParams(sdk.MustNewDecFromStr("0.000000000000000001"), sdk.MustNewDecFromStr("0.000000000000000002")),
		prevBlockTime,
	)

	kavamint.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), state)

	suite.Equal(state.Params, suite.Keeper.GetParams(suite.Ctx), "expected params to bet set in store")

	storeTime := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
	suite.Equal(state.PreviousBlockTime, storeTime, "expected previous block time to be set in store")

	exportedState := kavamint.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(state, exportedState, "expected exported state to match imported state")
}

func (suite *genesisTestSuite) Test_InitAndExportGenesis_ZeroValues() {
	state := types.NewGenesisState(
		types.NewParams(sdk.ZeroDec(), sdk.ZeroDec()),
		time.Time{},
	)

	kavamint.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), state)

	suite.Equal(state.Params, suite.Keeper.GetParams(suite.Ctx), "expected params to bet set in store")

	storeTime := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
	suite.Equal(state.PreviousBlockTime, storeTime, "expected previous block time to be set in store")

	exportedState := kavamint.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(state, exportedState, "expected exported state to match imported state")
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
