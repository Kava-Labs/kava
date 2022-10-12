package app_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

func (suite *UpgradeTestSuite) TestAddKavadistFundAccount() {
	ak := suite.App.GetAccountKeeper()
	maccAddr := ak.GetModuleAddress(kavadisttypes.FundModuleAccount)

	dstk := suite.App.GetDistrKeeper()

	communityCoinsBefore := dstk.GetFeePoolCommunityCoins(suite.Ctx)

	acc := ak.NewAccountWithAddress(suite.Ctx, maccAddr)
	ak.SetAccount(suite.Ctx, acc)

	bal := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000000000)))
	suite.App.FundAccount(suite.Ctx, maccAddr, bal)

	// Ensure it is a module account prior to migration
	acc = ak.GetAccount(suite.Ctx, maccAddr)
	_, ok := acc.(authtypes.ModuleAccountI)
	suite.Require().Falsef(ok, "account should not a ModuleAccount: %T", acc)

	suite.Require().IsType(&authtypes.BaseAccount{}, acc)

	app.AddKavadistFundAccount(
		suite.Ctx,
		ak,
		suite.App.GetBankKeeper(),
		dstk,
	)

	acc = ak.GetAccount(suite.Ctx, maccAddr)
	suite.Require().Implements((*authtypes.ModuleAccountI)(nil), acc)

	communityCoinsAfter := dstk.GetFeePoolCommunityCoins(suite.Ctx)

	suite.Equal(
		communityCoinsBefore.Add(sdk.NewDecCoinsFromCoins(bal...)...),
		communityCoinsAfter,
	)
}
