package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	app "github.com/kava-labs/kava/app"
	v015cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_15"
	v016cdp "github.com/kava-labs/kava/x/cdp/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses   []sdk.AccAddress
	v15genstate v015cdp.GenesisState
	cdc         codec.Codec
	legacyCdc   *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v15genstate = v015cdp.GenesisState{
		Params:                    v015cdp.Params{},
		CDPs:                      v015cdp.CDPs{},
		Deposits:                  v015cdp.Deposits{},
		StartingCdpID:             1,
		DebtDenom:                 "usdx",
		GovDenom:                  "ukava",
		PreviousAccumulationTimes: v015cdp.GenesisAccumulationTimes{},
		TotalPrincipals:           v015cdp.GenesisTotalPrincipals{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	file := filepath.Join("testdata", "v15-cdp.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	err = s.legacyCdc.UnmarshalJSON(data, &s.v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(s.v15genstate)
	actual := s.cdc.MustMarshalJSON(genstate)

	file = filepath.Join("testdata", "v16-cdp.json")
	expected, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_GenState() {
	s.v15genstate = v015cdp.GenesisState{
		StartingCdpID: 2,
		DebtDenom:     "usdx",
		GovDenom:      "ukava",
		Params: v015cdp.Params{
			CollateralParams: v015cdp.CollateralParams{
				{
					Denom:                            "xrp",
					Type:                             "xrp-a",
					LiquidationRatio:                 sdk.MustNewDecFromStr("2.0"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.012"),
					LiquidationPenalty:               sdk.MustNewDecFromStr("0.05"),
					AuctionSize:                      sdkmath.NewInt(70),
					SpotMarketID:                     "xrp:usd",
					LiquidationMarketID:              "xrp:usd",
					KeeperRewardPercentage:           sdk.MustNewDecFromStr("0.01"),
					CheckCollateralizationIndexCount: sdkmath.NewInt(10),
					ConversionFactor:                 sdkmath.NewInt(6),
				},
			},
			DebtParam: v015cdp.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: sdkmath.NewInt(6),
				DebtFloor:        sdkmath.NewInt(100),
			},
			GlobalDebtLimit:         sdk.NewInt64Coin("usdx", 1000000000000),
			SurplusAuctionThreshold: sdkmath.NewInt(6),
			SurplusAuctionLot:       sdkmath.NewInt(7),
			DebtAuctionThreshold:    sdkmath.NewInt(8),
			DebtAuctionLot:          sdkmath.NewInt(9),
		},
		CDPs: v015cdp.CDPs{
			{
				ID:              2,
				Owner:           s.addresses[0],
				Type:            "xrp-a",
				Collateral:      sdk.NewCoin("xrp", sdkmath.NewInt(2123)),
				Principal:       sdk.NewCoin("usdx", sdkmath.NewInt(100)),
				AccumulatedFees: sdk.NewCoin("usdx", sdk.ZeroInt()),
				FeesUpdated:     time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				InterestFactor:  sdk.NewDec(1),
			},
		},
		Deposits: v015cdp.Deposits{
			{
				CdpID:     1,
				Depositor: s.addresses[0],
				Amount:    sdk.NewCoin("usdx", sdkmath.NewInt(100)),
			},
			{
				CdpID:     2,
				Depositor: s.addresses[1],
				Amount:    sdk.NewCoin("ukava", sdkmath.NewInt(1200)),
			},
		},
		PreviousAccumulationTimes: v015cdp.GenesisAccumulationTimes{
			{
				CollateralType:           "usdx",
				PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				InterestFactor:           sdk.MustNewDecFromStr("0.02"),
			},
		},
		TotalPrincipals: v015cdp.GenesisTotalPrincipals{
			{
				CollateralType: "usdx",
				TotalPrincipal: sdkmath.NewInt(1200),
			},
		},
	}
	expected := v016cdp.GenesisState{
		StartingCdpID: 2,
		DebtDenom:     "usdx",
		GovDenom:      "ukava",
		Params: v016cdp.Params{
			CollateralParams: v016cdp.CollateralParams{
				{
					Denom:                            "xrp",
					Type:                             "xrp-a",
					LiquidationRatio:                 sdk.MustNewDecFromStr("2.0"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.012"),
					LiquidationPenalty:               sdk.MustNewDecFromStr("0.05"),
					AuctionSize:                      sdkmath.NewInt(70),
					SpotMarketID:                     "xrp:usd",
					LiquidationMarketID:              "xrp:usd",
					KeeperRewardPercentage:           sdk.MustNewDecFromStr("0.01"),
					CheckCollateralizationIndexCount: sdkmath.NewInt(10),
					ConversionFactor:                 sdkmath.NewInt(6),
				},
			},
			DebtParam: v016cdp.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: sdkmath.NewInt(6),
				DebtFloor:        sdkmath.NewInt(100),
			},
			GlobalDebtLimit:                    sdk.NewInt64Coin("usdx", 1000000000000),
			SurplusAuctionThreshold:            sdkmath.NewInt(6),
			SurplusAuctionLot:                  sdkmath.NewInt(7),
			DebtAuctionThreshold:               sdkmath.NewInt(8),
			DebtAuctionLot:                     sdkmath.NewInt(9),
			BeginBlockerExecutionBlockInterval: v016cdp.DefaultBeginBlockerExecutionBlockInterval,
		},
		CDPs: v016cdp.CDPs{
			{
				ID:              2,
				Owner:           s.addresses[0],
				Type:            "xrp-a",
				Collateral:      sdk.NewCoin("xrp", sdkmath.NewInt(2123)),
				Principal:       sdk.NewCoin("usdx", sdkmath.NewInt(100)),
				AccumulatedFees: sdk.NewCoin("usdx", sdk.ZeroInt()),
				FeesUpdated:     time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				InterestFactor:  sdk.NewDec(1),
			},
		},
		Deposits: v016cdp.Deposits{
			{
				CdpID:     1,
				Depositor: s.addresses[0],
				Amount:    sdk.NewCoin("usdx", sdkmath.NewInt(100)),
			},
			{
				CdpID:     2,
				Depositor: s.addresses[1],
				Amount:    sdk.NewCoin("ukava", sdkmath.NewInt(1200)),
			},
		},
		PreviousAccumulationTimes: v016cdp.GenesisAccumulationTimes{
			{
				CollateralType:           "usdx",
				PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				InterestFactor:           sdk.MustNewDecFromStr("0.02"),
			},
		},
		TotalPrincipals: v016cdp.GenesisTotalPrincipals{
			{
				CollateralType: "usdx",
				TotalPrincipal: sdkmath.NewInt(1200),
			},
		},
	}
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expected, *genState)
}

func TestCdpMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
