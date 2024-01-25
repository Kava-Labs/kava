package testutil

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

func NewCDPGenState(cdc codec.JSONCodec, denom, asset string, liquidationRatio sdk.Dec) app.GenesisState {
	cdpGenesis := cdptypes.GenesisState{
		Params: cdptypes.Params{
			GlobalDebtLimit:         sdk.NewInt64Coin("usdx", 1000000000000),
			SurplusAuctionThreshold: cdptypes.DefaultSurplusThreshold,
			SurplusAuctionLot:       cdptypes.DefaultSurplusLot,
			DebtAuctionThreshold:    cdptypes.DefaultDebtThreshold,
			DebtAuctionLot:          cdptypes.DefaultDebtLot,
			CollateralParams: cdptypes.CollateralParams{
				{
					Denom:                            denom,
					Type:                             asset + "-a",
					LiquidationRatio:                 liquidationRatio,
					DebtLimit:                        sdk.NewInt64Coin("usdx", 1000000000000),
					StabilityFee:                     sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:               sdk.MustNewDecFromStr("0.05"),
					AuctionSize:                      sdk.NewInt(100),
					SpotMarketID:                     asset + ":usd",
					LiquidationMarketID:              asset + ":usd",
					KeeperRewardPercentage:           sdk.MustNewDecFromStr("0.01"),
					CheckCollateralizationIndexCount: sdk.NewInt(10),
					ConversionFactor:                 sdk.NewInt(6),
				},
			},
			DebtParam: cdptypes.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: sdk.NewInt(6),
				DebtFloor:        sdk.NewInt(10000000),
			},
		},
		StartingCdpID: cdptypes.DefaultCdpStartingID,
		DebtDenom:     cdptypes.DefaultDebtDenom,
		GovDenom:      cdptypes.DefaultGovDenom,
		CDPs:          cdptypes.CDPs{},
		PreviousAccumulationTimes: cdptypes.GenesisAccumulationTimes{
			cdptypes.NewGenesisAccumulationTime(asset+"-a", time.Time{}, sdk.OneDec()),
		},
		TotalPrincipals: cdptypes.GenesisTotalPrincipals{
			cdptypes.NewGenesisTotalPrincipal(asset+"-a", sdk.ZeroInt()),
		},
	}
	return app.GenesisState{cdptypes.ModuleName: cdc.MustMarshalJSON(&cdpGenesis)}
}
