package cdp_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/cdp"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/stretchr/testify/suite"
)

type CDPAdapterTestSuite struct {
	suite.Suite

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
	denoms      []string
}

func TestCDPAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(CDPAdapterTestSuite))
}

func (suite *CDPAdapterTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
	suite.app = app.NewTestApp()
	cdc := suite.app.AppCodec()
	suite.denoms = []string{"xrp", "btc"}

	authGS := app.NewFundedGenStateWithSameCoins(
		cdc,
		cs(c(suite.denoms[0], 500000000), c(suite.denoms[1], 500000000)),
		suite.addrs[0:2],
	)

	suite.app.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(cdc),
		NewCDPGenStateMulti(cdc),
	)

	suite.ctx = suite.app.NewContext(true, tmprototypes.Header{Time: suite.genesisTime})
}

func (suite *CDPAdapterTestSuite) TestEarnAdapter_OwnerSharesBySource() {
	cdpKeeper := suite.app.GetCDPKeeper()
	adapter := cdp.NewSourceAdapter(cdpKeeper)

	err := cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], c(suite.denoms[0], 400000000), c("usdx", 10000000), "xrp-a")
	suite.NoError(err)

	err = cdpKeeper.AddCdp(suite.ctx, suite.addrs[1], c(suite.denoms[1], 400000000), c("usdx", 20000000), "btc-a")
	suite.NoError(err)

	tests := []struct {
		name          string
		giveOwner     sdk.AccAddress
		giveSourceIDs []string
		wantShares    map[string]sdk.Dec
	}{
		{
			"depositor has shares",
			suite.addrs[0],
			[]string{
				suite.denoms[0] + "-a",
			},
			map[string]sdk.Dec{
				suite.denoms[0] + "-a": sdk.NewDecWithPrec(10000000, 0),
			},
		},
		{
			"depositor has shares - including empty deposits",
			suite.addrs[1],
			[]string{
				suite.denoms[1] + "-a",
				"unknown",
			},
			map[string]sdk.Dec{
				suite.denoms[1] + "-a": sdk.NewDecWithPrec(20000000, 0),
				"unknown":              sdk.ZeroDec(),
			},
		},
		{
			"non-depositor has zero shares",
			suite.addrs[2],
			[]string{
				suite.denoms[0],
			},
			map[string]sdk.Dec{
				suite.denoms[0]: sdk.ZeroDec(),
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			shares := adapter.OwnerSharesBySource(suite.ctx, tt.giveOwner, tt.giveSourceIDs)

			suite.Equal(tt.wantShares, shares)
		})
	}
}

func (suite *CDPAdapterTestSuite) TestEarnAdapter_TotalSharesBySource() {
	cdpKeeper := suite.app.GetCDPKeeper()
	adapter := cdp.NewSourceAdapter(cdpKeeper)

	err := cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], c(suite.denoms[0], 400000000), c("usdx", 10000000), "xrp-a")
	suite.NoError(err)

	err = cdpKeeper.AddCdp(suite.ctx, suite.addrs[1], c(suite.denoms[0], 400000000), c("usdx", 20000000), "xrp-a")
	suite.NoError(err)

	tests := []struct {
		name         string
		giveSourceID string
		wantShares   sdk.Dec
	}{
		{
			"total shares",
			suite.denoms[0] + "-a",
			sdk.NewDecWithPrec(30000000, 0),
		},
		{
			"empty or invalid vault empty",
			suite.denoms[1] + "-a",
			sdk.ZeroDec(),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			shares := adapter.TotalSharesBySource(suite.ctx, tt.giveSourceID)

			suite.Equal(tt.wantShares, shares)
		})
	}
}

func NewPricefeedGenStateMulti(cdc codec.JSONCodec) app.GenesisState {
	pfGenesis := pricefeedtypes.GenesisState{
		Params: pricefeedtypes.Params{
			Markets: []pricefeedtypes.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "btc:usd:30", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd:30", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd:30", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "busd:usd", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "busd:usd:30", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeedtypes.PostedPrice{
			{
				MarketID:      "btc:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "btc:usd:30",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd:30",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("17.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd:30",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("17.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "busd:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.OneDec(),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "busd:usd:30",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.OneDec(),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeedtypes.ModuleName: cdc.MustMarshalJSON(&pfGenesis)}
}

func NewCDPGenStateMulti(cdc codec.JSONCodec) app.GenesisState {
	cdpGenesis := cdptypes.GenesisState{
		Params: cdptypes.Params{
			GlobalDebtLimit:         sdk.NewInt64Coin("usdx", 2000000000000),
			SurplusAuctionThreshold: cdptypes.DefaultSurplusThreshold,
			SurplusAuctionLot:       cdptypes.DefaultSurplusLot,
			DebtAuctionThreshold:    cdptypes.DefaultDebtThreshold,
			DebtAuctionLot:          cdptypes.DefaultDebtLot,
			CollateralParams: cdptypes.CollateralParams{
				{
					Denom:                            "xrp",
					Type:                             "xrp-a",
					LiquidationRatio:                 sdk.MustNewDecFromStr("2.0"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:               d("0.05"),
					AuctionSize:                      i(7000000000),
					SpotMarketID:                     "xrp:usd",
					LiquidationMarketID:              "xrp:usd:30",
					KeeperRewardPercentage:           d("0.01"),
					CheckCollateralizationIndexCount: i(10),
					ConversionFactor:                 i(6),
				},
				{
					Denom:                            "btc",
					Type:                             "btc-a",
					LiquidationRatio:                 sdk.MustNewDecFromStr("1.5"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.000000000782997609"), // %2.5 apr
					LiquidationPenalty:               d("0.025"),
					AuctionSize:                      i(10000000),
					SpotMarketID:                     "btc:usd",
					LiquidationMarketID:              "btc:usd:30",
					KeeperRewardPercentage:           d("0.01"),
					CheckCollateralizationIndexCount: i(10),
					ConversionFactor:                 i(8),
				},
				{
					Denom:                            "bnb",
					Type:                             "bnb-a",
					LiquidationRatio:                 sdk.MustNewDecFromStr("1.5"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:               d("0.05"),
					AuctionSize:                      i(50000000000),
					SpotMarketID:                     "bnb:usd",
					LiquidationMarketID:              "bnb:usd:30",
					KeeperRewardPercentage:           d("0.01"),
					CheckCollateralizationIndexCount: i(10),
					ConversionFactor:                 i(8),
				},
				{
					Denom:                            "busd",
					Type:                             "busd-a",
					LiquidationRatio:                 d("1.01"),
					DebtLimit:                        sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:                     sdk.OneDec(), // %0 apr
					LiquidationPenalty:               d("0.05"),
					AuctionSize:                      i(10000000000),
					SpotMarketID:                     "busd:usd",
					LiquidationMarketID:              "busd:usd:30",
					KeeperRewardPercentage:           d("0.01"),
					CheckCollateralizationIndexCount: i(10),
					ConversionFactor:                 i(8),
				},
			},
			DebtParam: types.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: i(6),
				DebtFloor:        i(10000000),
			},
		},
		StartingCdpID: types.DefaultCdpStartingID,
		DebtDenom:     types.DefaultDebtDenom,
		GovDenom:      types.DefaultGovDenom,
		CDPs:          types.CDPs{},
		PreviousAccumulationTimes: types.GenesisAccumulationTimes{
			types.NewGenesisAccumulationTime("btc-a", time.Time{}, sdk.OneDec()),
			types.NewGenesisAccumulationTime("xrp-a", time.Time{}, sdk.OneDec()),
			types.NewGenesisAccumulationTime("busd-a", time.Time{}, sdk.OneDec()),
			types.NewGenesisAccumulationTime("bnb-a", time.Time{}, sdk.OneDec()),
		},
		TotalPrincipals: types.GenesisTotalPrincipals{
			types.NewGenesisTotalPrincipal("btc-a", sdk.ZeroInt()),
			types.NewGenesisTotalPrincipal("xrp-a", sdk.ZeroInt()),
			types.NewGenesisTotalPrincipal("busd-a", sdk.ZeroInt()),
			types.NewGenesisTotalPrincipal("bnb-a", sdk.ZeroInt()),
		},
	}
	return app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(&cdpGenesis)}
}

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
