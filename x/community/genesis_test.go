package community_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/x/community"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
)

type genesisTestSuite struct {
	testutil.Suite
}

func (suite *genesisTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) TestInitGenesis() {

	accountKeeper := suite.App.GetAccountKeeper()

	genesisState := types.NewGenesisState(
		types.NewParams(
			time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			sdkmath.LegacyNewDec(1000),
			sdkmath.LegacyNewDec(1000),
		),
		types.NewStakingRewardsState(
			time.Date(1997, 1, 1, 0, 0, 0, 0, time.UTC),
			sdkmath.LegacyMustNewDecFromStr("0.100000000000000000"),
		),
	)

	suite.NotPanics(func() {
		community.InitGenesis(suite.Ctx, suite.Keeper, accountKeeper, genesisState)
	})

	// check for module account this way b/c GetModuleAccount creates if not existing.
	acc := accountKeeper.GetAccount(suite.Ctx, suite.MaccAddress)
	suite.NotNil(acc)
	_, ok := acc.(authtypes.ModuleAccountI)
	suite.True(ok)

	keeper := suite.App.GetCommunityKeeper()
	storedParams, found := keeper.GetParams(suite.Ctx)
	suite.True(found)
	suite.Equal(genesisState.Params, storedParams)

	stakingRewardsState := keeper.GetStakingRewardsState(suite.Ctx)
	suite.Equal(genesisState.StakingRewardsState, stakingRewardsState)
}

func (suite *genesisTestSuite) TestExportGenesis() {
	params := types.NewParams(
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		sdkmath.LegacyNewDec(1000),
		sdkmath.LegacyNewDec(1000),
	)
	suite.Keeper.SetParams(suite.Ctx, params)

	stakingRewardsState := types.NewStakingRewardsState(
		time.Date(1997, 1, 1, 0, 0, 0, 0, time.UTC),
		sdkmath.LegacyMustNewDecFromStr("0.100000000000000000"),
	)
	suite.Keeper.SetStakingRewardsState(suite.Ctx, stakingRewardsState)

	genesisState := community.ExportGenesis(suite.Ctx, suite.Keeper)

	suite.Equal(params, genesisState.Params)
	suite.Equal(stakingRewardsState, genesisState.StakingRewardsState)
}

func (suite *genesisTestSuite) TestInitExportIsLossless() {
	genesisState := types.NewGenesisState(
		types.NewParams(
			time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			sdkmath.LegacyNewDec(1000),
			sdkmath.LegacyNewDec(1000),
		),
		types.NewStakingRewardsState(
			time.Date(1997, 1, 1, 0, 0, 0, 0, time.UTC),
			sdkmath.LegacyMustNewDecFromStr("0.100000000000000000"),
		),
	)

	community.InitGenesis(suite.Ctx, suite.Keeper, suite.App.GetAccountKeeper(), genesisState)
	exportedState := community.ExportGenesis(suite.Ctx, suite.Keeper)

	suite.Equal(genesisState, exportedState)
}
