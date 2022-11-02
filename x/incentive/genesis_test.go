package incentive_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

const (
	oneYear time.Duration = 365 * 24 * time.Hour
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	app    app.TestApp
	keeper keeper.Keeper
	addrs  []sdk.AccAddress

	genesisTime time.Time
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.app = tApp
	keeper := tApp.GetIncentiveKeeper()
	suite.genesisTime = time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	_, addrs := app.GeneratePrivKeyAddressPairs(5)

	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleAccount(addrs[0], cs(c("bnb", 1e10), c("ukava", 1e10))).
		WithSimpleModuleAccount(kavadisttypes.KavaDistMacc, cs(c("hard", 1e15), c("ukava", 1e15)))

	loanToValue, _ := sdk.NewDecFromStr("0.6")
	borrowLimit := sdk.NewDec(1000000000000000)
	hardGS := hardtypes.NewGenesisState(
		hardtypes.NewParams(
			hardtypes.MoneyMarkets{
				hardtypes.NewMoneyMarket("ukava", hardtypes.NewBorrowLimit(false, borrowLimit, loanToValue), "kava:usd", sdk.NewInt(1000000), hardtypes.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				hardtypes.NewMoneyMarket("bnb", hardtypes.NewBorrowLimit(false, borrowLimit, loanToValue), "bnb:usd", sdk.NewInt(1000000), hardtypes.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			},
			sdk.NewDec(10),
		),
		hardtypes.DefaultAccumulationTimes,
		hardtypes.DefaultDeposits,
		hardtypes.DefaultBorrows,
		hardtypes.DefaultTotalSupplied,
		hardtypes.DefaultTotalBorrowed,
		hardtypes.DefaultTotalReserves,
	)
	incentiveGS := types.NewGenesisState(
		types.NewParams(
			types.RewardPeriods{types.NewRewardPeriod(true, "bnb-a", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), c("ukava", 122354))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "bnb", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "bnb", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "btcb/usdx", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("swp", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultipliersPerDenoms{
				{
					Denom: "ukava",
					Multipliers: types.Multipliers{
						types.NewMultiplier("large", 12, d("1.0")),
					},
				},
				{
					Denom: "hard",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.25")),
						types.NewMultiplier("large", 12, d("1.0")),
					},
				},
				{
					Denom: "swp",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.25")),
						types.NewMultiplier("medium", 6, d("0.8")),
					},
				},
			},
			suite.genesisTime.Add(5*oneYear),
		),
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultClaims,
		types.DefaultUSDXClaims,
		types.DefaultHardClaims,
		types.DefaultDelegatorClaims,
		types.DefaultSwapClaims,
		types.DefaultSavingsClaims,
		types.DefaultEarnClaims,
	)

	cdc := suite.app.AppCodec()

	tApp.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authBuilder.BuildMarshalled(cdc),
		app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(&incentiveGS)},
		app.GenesisState{hardtypes.ModuleName: cdc.MustMarshalJSON(&hardGS)},
		NewCDPGenStateMulti(cdc),
		NewPricefeedGenStateMultiFromTime(cdc, suite.genesisTime),
	)

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: suite.genesisTime})

	suite.addrs = addrs
	suite.keeper = keeper
	suite.ctx = ctx
}

func (suite *GenesisTestSuite) TestExportedGenesisMatchesImported() {
	genesisTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	genesisState := types.NewGenesisState(
		types.NewParams(
			types.RewardPeriods{types.NewRewardPeriod(true, "bnb-a", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), c("ukava", 122354))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "bnb", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "bnb", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "btcb/usdx", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("swp", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultipliersPerDenoms{
				{
					Denom: "ukava",
					Multipliers: types.Multipliers{
						types.NewMultiplier("large", 12, d("1.0")),
					},
				},
				{
					Denom: "hard",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.25")),
						types.NewMultiplier("large", 12, d("1.0")),
					},
				},
				{
					Denom: "swp",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.25")),
						types.NewMultiplier("medium", 6, d("0.8")),
					},
				},
			},
			genesisTime.Add(5*oneYear),
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bnb-a", genesisTime),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("bnb-a", types.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.3")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bnb", genesisTime.Add(-1*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("bnb", types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bnb", genesisTime.Add(-2*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("bnb", types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.05")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("ukava", genesisTime.Add(-3*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("ukava", types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.2")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bctb/usdx", genesisTime.Add(-4*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("btcb/usdx", types.RewardIndexes{{CollateralType: "swap", RewardFactor: d("0.001")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("ukava", genesisTime.Add(-3*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("ukava", types.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.2")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("usdx", genesisTime.Add(-3*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("usdx", types.RewardIndexes{{CollateralType: "usdx", RewardFactor: d("0.2")}}),
			},
		),
		types.Claims{
			types.NewClaim(
				types.CLAIM_TYPE_USDX_MINTING,
				suite.addrs[3],
				nil,
				types.MultiRewardIndexes{{CollateralType: "btcb/usdx", RewardIndexes: types.RewardIndexes{{CollateralType: "swap", RewardFactor: d("0.0")}}}},
			),
		},
		types.USDXMintingClaims{
			types.NewUSDXMintingClaim(
				suite.addrs[0],
				c("ukava", 1e9),
				types.RewardIndexes{{CollateralType: "bnb-a", RewardFactor: d("0.3")}},
			),
			types.NewUSDXMintingClaim(
				suite.addrs[1],
				c("ukava", 1),
				types.RewardIndexes{{CollateralType: "bnb-a", RewardFactor: d("0.001")}},
			),
		},
		types.HardLiquidityProviderClaims{
			types.NewHardLiquidityProviderClaim(
				suite.addrs[0],
				cs(c("ukava", 1e9), c("hard", 1e9)),
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.01")}}}},
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.0")}}}},
			),
			types.NewHardLiquidityProviderClaim(
				suite.addrs[1],
				cs(c("hard", 1)),
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}}}},
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.0")}}}},
			),
		},
		types.DelegatorClaims{
			types.NewDelegatorClaim(
				suite.addrs[2],
				cs(c("hard", 5)),
				types.MultiRewardIndexes{{CollateralType: "ukava", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.2")}}}},
			),
		},
		types.SwapClaims{
			types.NewSwapClaim(
				suite.addrs[3],
				nil,
				types.MultiRewardIndexes{{CollateralType: "btcb/usdx", RewardIndexes: types.RewardIndexes{{CollateralType: "swap", RewardFactor: d("0.0")}}}},
			),
		},
		types.SavingsClaims{
			types.NewSavingsClaim(
				suite.addrs[3],
				nil,
				types.MultiRewardIndexes{{CollateralType: "ukava", RewardIndexes: types.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.0")}}}},
			),
		},
		types.EarnClaims{
			types.NewEarnClaim(
				suite.addrs[3],
				nil,
				types.MultiRewardIndexes{{CollateralType: "usdx", RewardIndexes: types.RewardIndexes{{CollateralType: "earn", RewardFactor: d("0.0")}}}},
			),
		},
	)

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 0, Time: genesisTime})

	// Incentive init genesis reads from the cdp keeper to check params are ok. So it needs to be initialized first.
	// Then the cdp keeper reads from pricefeed keeper to check its params are ok. So it also need initialization.
	tApp.InitializeFromGenesisStates(
		NewCDPGenStateMulti(tApp.AppCodec()),
		NewPricefeedGenStateMultiFromTime(tApp.AppCodec(), genesisTime),
	)

	incentive.InitGenesis(
		ctx,
		tApp.GetIncentiveKeeper(),
		tApp.GetAccountKeeper(),
		tApp.GetBankKeeper(),
		tApp.GetCDPKeeper(),
		genesisState,
	)

	exportedGenesisState := incentive.ExportGenesis(ctx, tApp.GetIncentiveKeeper())

	suite.Equal(genesisState, exportedGenesisState)
}

func (suite *GenesisTestSuite) TestInitGenesisPanicsWhenAccumulationTimesToLongAgo() {
	genesisTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	invalidRewardState := types.NewGenesisRewardState(
		types.AccumulationTimes{
			types.NewAccumulationTime(
				"bnb",
				genesisTime.Add(-23*incentive.EarliestValidAccumulationTime).Add(-time.Nanosecond),
			),
		},
		types.MultiRewardIndexes{},
	)
	minimalParams := types.Params{
		ClaimEnd: genesisTime.Add(5 * oneYear),
	}

	testCases := []struct {
		genesisState types.GenesisState
	}{
		{
			types.GenesisState{
				Params:          minimalParams,
				USDXRewardState: invalidRewardState,
			},
		},
		{
			types.GenesisState{
				Params:                minimalParams,
				HardSupplyRewardState: invalidRewardState,
			},
		},
		{
			types.GenesisState{
				Params:                minimalParams,
				HardBorrowRewardState: invalidRewardState,
			},
		},
		{
			types.GenesisState{
				Params:               minimalParams,
				DelegatorRewardState: invalidRewardState,
			},
		},
		{
			types.GenesisState{
				Params:          minimalParams,
				SwapRewardState: invalidRewardState,
			},
		},
		{
			types.GenesisState{
				Params:             minimalParams,
				SavingsRewardState: invalidRewardState,
			},
		},
	}

	for _, tc := range testCases {
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 0, Time: genesisTime})

		// Incentive init genesis reads from the cdp keeper to check params are ok. So it needs to be initialized first.
		// Then the cdp keeper reads from pricefeed keeper to check its params are ok. So it also need initialization.
		tApp.InitializeFromGenesisStates(
			NewCDPGenStateMulti(tApp.AppCodec()),
			NewPricefeedGenStateMultiFromTime(tApp.AppCodec(), genesisTime),
		)

		suite.PanicsWithValue(
			"found accumulation time '1975-01-06 23:59:59.999999999 +0000 UTC' more than '8760h0m0s' behind genesis time '1998-01-01 00:00:00 +0000 UTC'",
			func() {
				incentive.InitGenesis(
					ctx, tApp.GetIncentiveKeeper(),
					tApp.GetAccountKeeper(),
					tApp.GetBankKeeper(),
					tApp.GetCDPKeeper(),
					tc.genesisState,
				)
			},
		)
	}
}

func (suite *GenesisTestSuite) TestValidateAccumulationTime() {
	genTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	err := incentive.ValidateAccumulationTime(
		genTime.Add(-incentive.EarliestValidAccumulationTime).Add(-time.Nanosecond),
		genTime,
	)
	suite.Error(err)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
