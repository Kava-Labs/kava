package cdp_test

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
)

type GenesisTestSuite struct {
	suite.Suite

	app     app.TestApp
	ctx     sdk.Context
	genTime time.Time
	keeper  cdp.Keeper
	addrs   []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.genTime = tmtime.Canonical(time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC))
	suite.ctx = tApp.NewContext(true, abci.Header{Height: 1, Time: suite.genTime})
	suite.keeper = tApp.GetCDPKeeper()
	suite.app = tApp

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.addrs = addrs
}

func (suite *GenesisTestSuite) TestInvalidGenState() {
	type args struct {
		params             cdp.Params
		cdps               cdp.CDPs
		deposits           cdp.Deposits
		startingID         uint64
		debtDenom          string
		govDenom           string
		genAccumTimes      cdp.GenesisAccumulationTimes
		genTotalPrincipals cdp.GenesisTotalPrincipals
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type genesisTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			name: "empty debt denom",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          "",
				govDenom:           cdp.DefaultGovDenom,
				genAccumTimes:      cdp.DefaultGenesisState().PreviousAccumulationTimes,
				genTotalPrincipals: cdp.DefaultGenesisState().TotalPrincipals,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "debt denom invalid",
			},
		},
		{
			name: "empty gov denom",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          cdp.DefaultDebtDenom,
				govDenom:           "",
				genAccumTimes:      cdp.DefaultGenesisState().PreviousAccumulationTimes,
				genTotalPrincipals: cdp.DefaultGenesisState().TotalPrincipals,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "gov denom invalid",
			},
		},
		{
			name: "interest factor below one",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          cdp.DefaultDebtDenom,
				govDenom:           cdp.DefaultGovDenom,
				genAccumTimes:      cdp.GenesisAccumulationTimes{cdp.NewGenesisAccumulationTime("bnb-a", time.Time{}, sdk.OneDec().Sub(sdk.SmallestDec()))},
				genTotalPrincipals: cdp.DefaultGenesisState().TotalPrincipals,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "interest factor should be ≥ 1.0",
			},
		},
		{
			name: "negative total principal",
			args: args{
				params:             cdp.DefaultParams(),
				cdps:               cdp.CDPs{},
				deposits:           cdp.Deposits{},
				debtDenom:          cdp.DefaultDebtDenom,
				govDenom:           cdp.DefaultGovDenom,
				genAccumTimes:      cdp.DefaultGenesisState().PreviousAccumulationTimes,
				genTotalPrincipals: cdp.GenesisTotalPrincipals{cdp.NewGenesisTotalPrincipal("bnb-a", sdk.NewInt(-1))},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "total principal should be positive",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			gs := cdp.NewGenesisState(tc.args.params, tc.args.cdps, tc.args.deposits, tc.args.startingID,
				tc.args.debtDenom, tc.args.govDenom, tc.args.genAccumTimes, tc.args.genTotalPrincipals)
			err := gs.Validate()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *GenesisTestSuite) TestValidGenState() {
	suite.NotPanics(func() {
		suite.app.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
			NewCDPGenStateMulti(),
		)
	})

	cdpGS := NewCDPGenStateMulti()
	gs := cdp.GenesisState{}
	cdp.ModuleCdc.UnmarshalJSON(cdpGS["cdp"], &gs)
	gs.CDPs = cdps()
	gs.StartingCdpID = uint64(5)
	appGS := app.GenesisState{"cdp": cdp.ModuleCdc.MustMarshalJSON(gs)}
	suite.NotPanics(func() {
		suite.app.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
			appGS,
		)
	})
}

func (suite *GenesisTestSuite) Test_InitExportGenesis() {

	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:         sdk.NewInt64Coin("usdx", 1000000000000),
			SurplusAuctionThreshold: cdp.DefaultSurplusThreshold,
			SurplusAuctionLot:       cdp.DefaultSurplusLot,
			DebtAuctionThreshold:    cdp.DefaultDebtThreshold,
			DebtAuctionLot:          cdp.DefaultDebtLot,
			CollateralParams: cdp.CollateralParams{
				{
					Denom:                            "xrp",
					Type:                             "xrp-a",
					LiquidationRatio:                 sdk.MustNewDecFromStr("2.0"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.000000001547125958"), // 5% apr
					LiquidationPenalty:               d("0.05"),
					AuctionSize:                      i(7000000000),
					Prefix:                           0x20,
					SpotMarketID:                     "xrp:usd",
					LiquidationMarketID:              "xrp:usd",
					KeeperRewardPercentage:           d("0.01"),
					CheckCollateralizationIndexCount: i(10),
					ConversionFactor:                 i(6),
				},
				{
					Denom:                            "btc",
					Type:                             "btc-a",
					LiquidationRatio:                 sdk.MustNewDecFromStr("1.5"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.000000000782997609"), // 2.5% apr
					LiquidationPenalty:               d("0.025"),
					AuctionSize:                      i(10000000),
					Prefix:                           0x21,
					SpotMarketID:                     "btc:usd",
					LiquidationMarketID:              "btc:usd",
					KeeperRewardPercentage:           d("0.01"),
					CheckCollateralizationIndexCount: i(10),
					ConversionFactor:                 i(8),
				},
			},
			DebtParam: cdp.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: i(6),
				DebtFloor:        i(10000000),
			},
		},
		StartingCdpID: cdp.DefaultCdpStartingID,
		DebtDenom:     cdp.DefaultDebtDenom,
		GovDenom:      cdp.DefaultGovDenom,
		CDPs:          cdp.CDPs{},
		Deposits:      cdp.Deposits{},
		PreviousAccumulationTimes: cdp.GenesisAccumulationTimes{
			cdp.NewGenesisAccumulationTime("btc-a", time.Time{}, sdk.OneDec()),
			cdp.NewGenesisAccumulationTime("xrp-a", time.Time{}, sdk.OneDec()),
		},
		TotalPrincipals: cdp.GenesisTotalPrincipals{
			cdp.NewGenesisTotalPrincipal("btc-a", sdk.ZeroInt()),
			cdp.NewGenesisTotalPrincipal("xrp-a", sdk.ZeroInt()),
		},
	}

	suite.NotPanics(func() {
		suite.app.InitializeFromGenesisStatesWithTime(
			suite.genTime,
			NewPricefeedGenStateMulti(),
			app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)},
		)
	})

	expectedGenesis := cdpGenesis

	var expectedPrevAccTimes cdp.GenesisAccumulationTimes
	for _, prevAccTime := range cdpGenesis.PreviousAccumulationTimes {
		prevAccTime.PreviousAccumulationTime = suite.genTime
		expectedPrevAccTimes = append(expectedPrevAccTimes, prevAccTime)
	}
	expectedGenesis.PreviousAccumulationTimes = expectedPrevAccTimes

	// TODO: export panics unless BeginBlocker is run at least once because previous accrual times are not set in state
	cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, suite.keeper)
	exportedGenesis := cdp.ExportGenesis(suite.ctx, suite.keeper)

	// Sort TotalPrincipals in both genesis files so slice order matches
	sort.SliceStable(expectedGenesis.TotalPrincipals, func(i, j int) bool {
		return expectedGenesis.TotalPrincipals[i].CollateralType < expectedGenesis.TotalPrincipals[j].CollateralType
	})
	sort.SliceStable(exportedGenesis.TotalPrincipals, func(i, j int) bool {
		return exportedGenesis.TotalPrincipals[i].CollateralType < exportedGenesis.TotalPrincipals[j].CollateralType
	})

	// Sort PreviousAccumulationTimes in both genesis files so slice order matches
	sort.SliceStable(expectedGenesis.PreviousAccumulationTimes, func(i, j int) bool {
		return expectedGenesis.PreviousAccumulationTimes[i].CollateralType < expectedGenesis.PreviousAccumulationTimes[j].CollateralType
	})
	sort.SliceStable(exportedGenesis.PreviousAccumulationTimes, func(i, j int) bool {
		return exportedGenesis.PreviousAccumulationTimes[i].CollateralType < exportedGenesis.PreviousAccumulationTimes[j].CollateralType
	})

	suite.Equal(expectedGenesis, exportedGenesis)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
