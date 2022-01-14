package v0_16

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	app "github.com/kava-labs/kava/app"
	v015pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_15"
	v016pricefeed "github.com/kava-labs/kava/x/pricefeed/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses   []sdk.AccAddress
	v15genstate v015pricefeed.GenesisState
	cdc         codec.Codec
	legacyCdc   *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v15genstate = v015pricefeed.GenesisState{
		Params:       v015pricefeed.Params{},
		PostedPrices: v015pricefeed.PostedPrices{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	// Migrate v15 pricefeed to v16
	v15Params := `{
		"params": {
			"markets": [
				{
					"active": true,
					"base_asset": "bnb",
					"market_id": "bnb:usd",
					"oracles": ["kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"],
					"quote_asset": "usd"
				},
				{
					"active": true,
					"base_asset": "bnb",
					"market_id": "bnb:usd:30",
					"oracles": ["kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"],
					"quote_asset": "usd"
				}
			]
		},
		"posted_prices": [
			{
				"expiry": "2022-07-20T00:00:00Z",
				"market_id": "bnb:usd",
				"oracle_address": "kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em",
				"price": "215.962650000000001782"
			},
			{
				"expiry": "2022-07-20T00:00:00Z",
				"market_id": "bnb:usd:30",
				"oracle_address": "kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em",
				"price": "217.962650000000001782"
			}
		]
	}`

	expectedV16Params := `{
		"params": {
			"markets": [
				{
					"market_id": "bnb:usd",
					"base_asset": "bnb",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "bnb:usd:30",
					"base_asset": "bnb",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "atom:usd",
					"base_asset": "atom",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "atom:usd:30",
					"base_asset": "atom",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "akt:usd",
					"base_asset": "akt",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "akt:usd:30",
					"base_asset": "akt",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "luna:usd",
					"base_asset": "luna",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "luna:usd:30",
					"base_asset": "luna",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "osmo:usd",
					"base_asset": "osmo",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "osmo:usd:30",
					"base_asset": "osmo",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "ust:usd",
					"base_asset": "ust",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				},
				{
					"market_id": "ust:usd:30",
					"base_asset": "ust",
					"quote_asset": "usd",
					"oracles": [
						"kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em"
					],
					"active": true
				}
			]
		},
		"posted_prices": [
			{
				"market_id": "bnb:usd",
				"oracle_address": "kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em",
				"price": "215.962650000000001782",
				"expiry": "2022-07-20T00:00:00Z"
			},
			{
				"market_id": "bnb:usd:30",
				"oracle_address": "kava1acge4tcvhf3q6fh53fgwaa7vsq40wvx6wn50em",
				"price": "217.962650000000001782",
				"expiry": "2022-07-20T00:00:00Z"
			}
		]
	}`

	err := s.legacyCdc.UnmarshalJSON([]byte(v15Params), &s.v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(s.v15genstate)

	// v16 pricefeed json should be the same as v15 but with IBC markets added
	actual := s.cdc.MustMarshalJSON(genstate)

	s.Require().NoError(err)
	s.Require().JSONEq(expectedV16Params, string(actual))
}

func (s *migrateTestSuite) TestMigrate_Params() {
	s.v15genstate.Params = v015pricefeed.Params{
		Markets: v015pricefeed.Markets{
			{
				MarketID:   "market-1",
				BaseAsset:  "kava",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
		},
	}
	expectedParams := v016pricefeed.Params{
		Markets: v016pricefeed.Markets{
			{
				MarketID:   "market-1",
				BaseAsset:  "kava",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "atom:usd",
				BaseAsset:  "atom",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "atom:usd:30",
				BaseAsset:  "atom",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "akt:usd",
				BaseAsset:  "akt",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "akt:usd:30",
				BaseAsset:  "akt",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "luna:usd",
				BaseAsset:  "luna",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "luna:usd:30",
				BaseAsset:  "luna",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "osmo:usd",
				BaseAsset:  "osmo",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "osmo:usd:30",
				BaseAsset:  "osmo",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "ust:usd",
				BaseAsset:  "ust",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
			{
				MarketID:   "ust:usd:30",
				BaseAsset:  "ust",
				QuoteAsset: "usd",
				Oracles:    s.addresses,
				Active:     true,
			},
		},
	}
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expectedParams, genState.Params)
}

func (s *migrateTestSuite) TestMigrate_PostedPrices() {
	s.v15genstate.PostedPrices = v015pricefeed.PostedPrices{
		{
			MarketID:      "market-1",
			OracleAddress: s.addresses[0],
			Price:         sdk.MustNewDecFromStr("1.2"),
			Expiry:        time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			MarketID:      "market-2",
			OracleAddress: s.addresses[1],
			Price:         sdk.MustNewDecFromStr("1.899"),
			Expiry:        time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	expected := v016pricefeed.PostedPrices{
		{
			MarketID:      "market-1",
			OracleAddress: s.addresses[0],
			Price:         sdk.MustNewDecFromStr("1.2"),
			Expiry:        time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			MarketID:      "market-2",
			OracleAddress: s.addresses[1],
			Price:         sdk.MustNewDecFromStr("1.899"),
			Expiry:        time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expected, genState.PostedPrices)
}

func TestPriceFeedMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
