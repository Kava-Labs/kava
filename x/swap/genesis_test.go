package swap_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap"
	"github.com/kava-labs/kava/x/swap/testutil"
	"github.com/kava-labs/kava/x/swap/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type genesisTestSuite struct {
	testutil.Suite
}

func (suite *genesisTestSuite) Test_InitGenesis_ValidationPanic() {
	invalidState := types.NewGenesisState(
		types.Params{
			SwapFee: sdk.NewDec(-1),
		},
		types.PoolRecords{},
		types.ShareRecords{},
	)

	suite.Panics(func() {
		swap.InitGenesis(suite.Ctx, suite.Keeper, invalidState)
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
			AllowedPools: swap.AllowedPools{swap.NewAllowedPool("ukava", "usdx")},
			SwapFee:      sdk.MustNewDecFromStr("0.00255"),
		},
		types.PoolRecords{
			swap.NewPoolRecord(sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(2e6))), sdk.NewInt(1e6)),
			swap.NewPoolRecord(sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6))), sdk.NewInt(3e6)),
		},
		types.ShareRecords{
			types.NewShareRecord(depositor_2, "hard/usdx", sdk.NewInt(1e6)),
			types.NewShareRecord(depositor_1, "ukava/usdx", sdk.NewInt(3e6)),
		},
	)

	swap.InitGenesis(suite.Ctx, suite.Keeper, state)
	suite.Equal(state.Params, suite.Keeper.GetParams(suite.Ctx))

	poolRecord1, _ := suite.Keeper.GetPool(suite.Ctx, "hard/usdx")
	suite.Equal(state.PoolRecords[0], poolRecord1)
	poolRecord2, _ := suite.Keeper.GetPool(suite.Ctx, "ukava/usdx")
	suite.Equal(state.PoolRecords[1], poolRecord2)

	shareRecord1, _ := suite.Keeper.GetDepositorShares(suite.Ctx, depositor_2, "hard/usdx")
	suite.Equal(state.ShareRecords[0], shareRecord1)
	shareRecord2, _ := suite.Keeper.GetDepositorShares(suite.Ctx, depositor_1, "ukava/usdx")
	suite.Equal(state.ShareRecords[1], shareRecord2)

	exportedState := swap.ExportGenesis(suite.Ctx, suite.Keeper)
	suite.Equal(state, exportedState)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
