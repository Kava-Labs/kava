package kavamint_test

import (
	"testing"
	"time"

	"github.com/kava-labs/kava/x/kavamint"
	"github.com/kava-labs/kava/x/kavamint/testutil"
	"github.com/kava-labs/kava/x/kavamint/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type genesisTestSuite struct {
	testutil.KavamintTestSuite
}

func (suite *genesisTestSuite) Test_InitGenesis_NoTimeSetPanics() {
	invalidState := types.NewGenesisState(
		types.DefaultParams(),
		time.Time{},
	)

	suite.Panics(func() {
		kavamint.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), invalidState)
	}, "expected init genesis to panic with invalid state")
}

func (suite *genesisTestSuite) Test_InitAndExportGenesis() {
	prevBlockTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	state := types.NewGenesisState(
		types.NewParams(sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.15")),
		prevBlockTime,
	)

	kavamint.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), state)

	suite.Equal(state.Params, suite.Keeper.GetParams(suite.Ctx))
	storeTime, found := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
	suite.True(found)
	suite.True(state.PreviousBlockTime.Equal(storeTime))
	suite.True(prevBlockTime.Equal(storeTime))

	exportedState := kavamint.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(state, exportedState)
}

func (suite *genesisTestSuite) Test_InitAndExportGenesis_DefaultsToBlockTime() {
	// init genesis
	kavamint.InitGenesis(
		suite.Ctx,
		suite.Keeper,
		suite.App.GetAccountKeeper(),
		types.DefaultGenesisState(),
	)

	// unset time
	suite.Keeper.SetPreviousBlockTime(suite.Ctx, time.Time{})

	// check that block time defaults to block time
	exportedState := kavamint.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(types.DefaultParams(), exportedState.Params)
	suite.Equal(suite.Ctx.BlockTime(), exportedState.PreviousBlockTime)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
