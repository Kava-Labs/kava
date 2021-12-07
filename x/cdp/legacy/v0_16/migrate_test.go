package v0_16

import (
	"testing"
	"time"

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

// func (s *migrateTestSuite) TestMigrate_JSON() {
// 	// Migrate v15 cdp to v16
// 	data := `{
// 		"params": {
// 			"assets": [
// 				{
// 					"blockable": true,
// 					"blocked_addresses": null,
// 					"denom": "hbtc",
// 					"owner": "kava1dmm9zpdnm6mfhywzt9sstm4p33y0cnsd0m673z",
// 					"paused": false,
// 					"rate_limit": {
// 						"active": false,
// 						"limit": "0",
// 						"time_period": "0"
// 					}
// 				}
// 			]
// 		},
// 		"supplies": [
// 			{
// 				"current_supply": { "denom": "ukava", "amount": "100" },
// 				"time_elapsed": "3600000000000"
// 			},
// 			{
// 				"current_supply": { "denom": "bnb", "amount": "300" },
// 				"time_elapsed": "300000000000"
// 			}
// 		]
// 	}`
// 	err := s.legacyCdc.UnmarshalJSON([]byte(data), &s.v15genstate)
// 	s.Require().NoError(err)
// 	genstate := Migrate(s.v15genstate)

// 	// Compare expect v16 cdp json with migrated json
// 	expected := `{
// 		"params": {
// 			"assets": [
// 				{
// 					"blockable": true,
// 					"blocked_addresses": [],
// 					"denom": "hbtc",
// 					"owner": "kava1dmm9zpdnm6mfhywzt9sstm4p33y0cnsd0m673z",
// 					"paused": false,
// 					"rate_limit": {
// 						"active": false,
// 						"limit": "0",
// 						"time_period": "0s"
// 					}
// 				}
// 			]
// 		},
// 		"supplies": [
// 			{
// 				"current_supply": { "denom": "ukava", "amount": "100" },
// 				"time_elapsed": "3600s"
// 			},
// 			{
// 				"current_supply": { "denom": "bnb", "amount": "300" },
// 				"time_elapsed": "300s"
// 			}
// 		]
// 	}`
// 	actual := s.cdc.MustMarshalJSON(genstate)
// 	s.Require().NoError(err)
// 	s.Require().JSONEq(expected, string(actual))
// }

func (s *migrateTestSuite) TestMigrate_Params() {
	s.v15genstate.Params = v015cdp.Params{
		CollateralParams: v015cdp.CollateralParams{
			{
				Denom:                            "xrp",
				Type:                             "xrp-a",
				LiquidationRatio:                 sdk.MustNewDecFromStr("2.0"),
				DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
				StabilityFee:                     sdk.MustNewDecFromStr("1.012"),
				LiquidationPenalty:               sdk.MustNewDecFromStr("0.05"),
				AuctionSize:                      sdk.NewInt(70),
				SpotMarketID:                     "xrp:usd",
				LiquidationMarketID:              "xrp:usd",
				KeeperRewardPercentage:           sdk.MustNewDecFromStr("0.01"),
				CheckCollateralizationIndexCount: sdk.NewInt(10),
				ConversionFactor:                 sdk.NewInt(6),
			},
		},
		DebtParam: v015cdp.DebtParam{
			Denom:            "usdx",
			ReferenceAsset:   "usd",
			ConversionFactor: sdk.NewInt(6),
			DebtFloor:        sdk.NewInt(100),
		},
		GlobalDebtLimit:         sdk.NewInt64Coin("usdx", 1000000000000),
		SurplusAuctionThreshold: sdk.NewInt(6),
		SurplusAuctionLot:       sdk.NewInt(7),
		DebtAuctionThreshold:    sdk.NewInt(8),
		DebtAuctionLot:          sdk.NewInt(9),
	}
	expectedParams := v016cdp.Params{
		Assets: []v016cdp.Asset{
			{
				Owner:            s.addresses[0].String(),
				Denom:            "ukava",
				BlockedAddresses: []string{s.addresses[1].String()},
				Paused:           true,
				Blockable:        true,
				RateLimit: v016cdp.RateLimit{
					Active:     true,
					Limit:      sdk.NewInt(10),
					TimePeriod: 1 * time.Hour,
				},
			},
		},
	}
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expectedParams, genState.Params)
}

// func (s *migrateTestSuite) TestMigrate_Supplies() {
// 	s.v15genstate.Supplies = v015cdp.AssetSupplies{
// 		{
// 			CurrentSupply: sdk.NewCoin("ukava", sdk.NewInt(100)),
// 			TimeElapsed:   time.Duration(1 * time.Hour),
// 		},
// 		{
// 			CurrentSupply: sdk.NewCoin("bnb", sdk.NewInt(300)),
// 			TimeElapsed:   time.Duration(5 * time.Minute),
// 		},
// 	}
// 	expected := []v016cdp.AssetSupply{
// 		{
// 			CurrentSupply: sdk.NewCoin("ukava", sdk.NewInt(100)),
// 			TimeElapsed:   time.Duration(1 * time.Hour),
// 		},
// 		{
// 			CurrentSupply: sdk.NewCoin("bnb", sdk.NewInt(300)),
// 			TimeElapsed:   time.Duration(5 * time.Minute),
// 		},
// 	}
// 	genState := Migrate(s.v15genstate)
// 	s.Require().Equal(expected, genState.Supplies)
// }

func TestcdpMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
