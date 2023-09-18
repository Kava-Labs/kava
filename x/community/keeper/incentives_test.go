package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Test suite used for all incentive tests
type IncentivesTestSuite struct {
	suite.Suite

	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper
}

// The default state used by each test
func (suite *IncentivesTestSuite) SetupTest() {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	suite.App = tApp
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.Keeper = suite.App.GetCommunityKeeper()

	// Set up x/mint and x/kavadist gen state
	mintGen := minttypes.DefaultGenesisState()
	mintGen.Params.InflationMax = sdk.NewDecWithPrec(595, 3)
	mintGen.Params.InflationMin = sdk.NewDecWithPrec(595, 3)
	kavadistGen := kavadisttypes.DefaultGenesisState()
	kavadistGen.Params.Active = true
	appCodec := tApp.AppCodec()
	suite.App.InitializeFromGenesisStates(
		app.GenesisState{minttypes.ModuleName: appCodec.MustMarshalJSON(mintGen)},
		app.GenesisState{kavadisttypes.ModuleName: appCodec.MustMarshalJSON(kavadistGen)},
	)
}

func TestIncentivesTestSuite(t *testing.T) {
	suite.Run(t, new(IncentivesTestSuite))
}

func (suite *IncentivesTestSuite) TestStartCommunityFundConsolidation() {
	suite.SetupTest()

	initialFeePool := distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(
			sdk.NewDecCoinFromDec("ukava", sdk.NewDecWithPrec(123456, 2)),
			sdk.NewDecCoinFromDec("usdx", sdk.NewDecWithPrec(654321, 3)),
		),
	}

	initialFeePoolCoins, initialFeePoolDust := initialFeePool.CommunityPool.TruncateDecimal()

	// More coins than initial feepool/communitypool
	fundCoins := sdk.NewCoins(
		sdk.NewInt64Coin("ukava", 10_000),
		sdk.NewInt64Coin("usdx", 10_000),
	)

	err := suite.App.FundModuleAccount(
		suite.Ctx,
		distrtypes.ModuleName,
		fundCoins,
	)
	suite.NoError(err, "x/distribution account should be funded without error")
	err = suite.App.FundModuleAccount(
		suite.Ctx,
		kavadisttypes.ModuleName,
		fundCoins,
	)
	suite.NoError(err, "x/kavadist account should be funded without error")

	suite.App.GetDistrKeeper().SetFeePool(suite.Ctx, initialFeePool)

	// Ensure the feepool was set before migration
	feePoolBefore := suite.App.GetDistrKeeper().GetFeePool(suite.Ctx)
	suite.Equal(initialFeePool, feePoolBefore, "initial feepool should be set")
	communityBalanceBefore := suite.App.GetCommunityKeeper().GetModuleAccountBalance(suite.Ctx)

	kavadistAcc := suite.App.GetAccountKeeper().GetModuleAccount(suite.Ctx, kavadisttypes.KavaDistMacc)
	kavaDistCoinsBefore := suite.App.GetBankKeeper().GetAllBalances(suite.Ctx, kavadistAcc.GetAddress())
	suite.Equal(
		fundCoins,
		kavaDistCoinsBefore,
		"x/kavadist balance should be funded",
	)

	// -------------
	// Run upgrade

	suite.setUpgradeTimeFromNow(-2 * time.Minute)
	err = suite.Keeper.StartCommunityFundConsolidation(suite.Ctx)
	suite.NoError(err, "consolidation should not error")

	// -------------
	// Check results

	feePoolAfter := suite.App.GetDistrKeeper().GetFeePool(suite.Ctx)
	suite.Equal(
		initialFeePoolDust,
		feePoolAfter.CommunityPool,
		"x/distribution community pool should be sent to x/community",
	)

	kavaDistCoinsAfter := suite.App.GetBankKeeper().GetAllBalances(suite.Ctx, kavadistAcc.GetAddress())
	suite.Equal(
		sdk.NewCoins(),
		kavaDistCoinsAfter,
		"x/kavadist balance should be empty",
	)

	totalExpectedCommunityPoolCoins := communityBalanceBefore.
		Add(initialFeePoolCoins...). // x/distribution fee pool
		Add(fundCoins...)            // x/kavadist module balance

	communityBalanceAfter := suite.App.GetCommunityKeeper().GetModuleAccountBalance(suite.Ctx)

	suite.Equal(
		totalExpectedCommunityPoolCoins,
		communityBalanceAfter,
		"x/community balance should be increased by the truncated x/distribution community pool",
	)
}

func (suite *IncentivesTestSuite) setUpgradeTimeFromNow(t time.Duration) {
	params, found := suite.Keeper.GetParams(suite.Ctx)
	suite.True(found)
	params.UpgradeTimeDisableInflation = suite.Ctx.BlockTime().Add(t)
	suite.Keeper.SetParams(suite.Ctx, params)
}
