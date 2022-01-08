package v0_16

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	app "github.com/kava-labs/kava/app"
	v015issuance "github.com/kava-labs/kava/x/issuance/legacy/v0_15"
	v016issuance "github.com/kava-labs/kava/x/issuance/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses   []sdk.AccAddress
	v15genstate v015issuance.GenesisState
	cdc         codec.Codec
	legacyCdc   *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v15genstate = v015issuance.GenesisState{
		Params:   v015issuance.Params{},
		Supplies: v015issuance.AssetSupplies{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	// Migrate v15 issuance to v16
	data := `{
		"params": {
			"assets": [
				{
					"blockable": true,
					"blocked_addresses": null,
					"denom": "hbtc",
					"owner": "kava1dmm9zpdnm6mfhywzt9sstm4p33y0cnsd0m673z",
					"paused": false,
					"rate_limit": {
						"active": false,
						"limit": "0",
						"time_period": "0"
					}
				}
			]
		},
		"supplies": [
			{
				"current_supply": { "denom": "ukava", "amount": "100" },
				"time_elapsed": "3600000000000"
			},
			{
				"current_supply": { "denom": "bnb", "amount": "300" },
				"time_elapsed": "300000000000"
			}
		]
	}`
	err := s.legacyCdc.UnmarshalJSON([]byte(data), &s.v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(s.v15genstate)

	// Compare expect v16 issuance json with migrated json
	expected := `{
		"params": {
			"assets": [
				{
					"blockable": true,
					"blocked_addresses": [],
					"denom": "hbtc",
					"owner": "kava1dmm9zpdnm6mfhywzt9sstm4p33y0cnsd0m673z",
					"paused": false,
					"rate_limit": {
						"active": false,
						"limit": "0",
						"time_period": "0s"
					}
				}
			]
		},
		"supplies": [
			{
				"current_supply": { "denom": "ukava", "amount": "100" },
				"time_elapsed": "3600s"
			},
			{
				"current_supply": { "denom": "bnb", "amount": "300" },
				"time_elapsed": "300s"
			}
		]
	}`
	actual := s.cdc.MustMarshalJSON(genstate)
	s.Require().NoError(err)
	s.Require().JSONEq(expected, string(actual))
}

func (s *migrateTestSuite) TestMigrate_Params() {
	s.v15genstate.Params = v015issuance.Params{
		Assets: v015issuance.Assets{
			{
				Owner:            s.addresses[0],
				Denom:            "ukava",
				BlockedAddresses: s.addresses[1:2],
				Paused:           true,
				Blockable:        true,
				RateLimit: v015issuance.RateLimit{
					Active:     true,
					Limit:      sdk.NewInt(10),
					TimePeriod: 1 * time.Hour,
				},
			},
		},
	}
	expectedParams := v016issuance.Params{
		Assets: []v016issuance.Asset{
			{
				Owner:            s.addresses[0].String(),
				Denom:            "ukava",
				BlockedAddresses: []string{s.addresses[1].String()},
				Paused:           true,
				Blockable:        true,
				RateLimit: v016issuance.RateLimit{
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

func (s *migrateTestSuite) TestMigrate_Supplies() {
	s.v15genstate.Supplies = v015issuance.AssetSupplies{
		{
			CurrentSupply: sdk.NewCoin("ukava", sdk.NewInt(100)),
			TimeElapsed:   time.Duration(1 * time.Hour),
		},
		{
			CurrentSupply: sdk.NewCoin("bnb", sdk.NewInt(300)),
			TimeElapsed:   time.Duration(5 * time.Minute),
		},
	}
	expected := []v016issuance.AssetSupply{
		{
			CurrentSupply: sdk.NewCoin("ukava", sdk.NewInt(100)),
			TimeElapsed:   time.Duration(1 * time.Hour),
		},
		{
			CurrentSupply: sdk.NewCoin("bnb", sdk.NewInt(300)),
			TimeElapsed:   time.Duration(5 * time.Minute),
		},
	}
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expected, genState.Supplies)
}

func TestIssuanceMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
