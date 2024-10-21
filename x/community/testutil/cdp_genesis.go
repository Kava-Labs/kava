package testutil

import (
	sdkmath "cosmossdk.io/math"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

func NewCDPGenState(cdc codec.JSONCodec, denom, asset string, liquidationRatio sdkmath.LegacyDec) app.GenesisState {
	cdpGenesis := cdptypes.GenesisState{
		Params: cdptypes.Params{
			GlobalDebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
			SurplusAuctionThreshold:  cdptypes.DefaultSurplusThreshold,
			SurplusAuctionLot:        cdptypes.DefaultSurplusLot,
			DebtAuctionThreshold:     cdptypes.DefaultDebtThreshold,
			DebtAuctionLot:           cdptypes.DefaultDebtLot,
			LiquidationBlockInterval: cdptypes.DefaultBeginBlockerExecutionBlockInterval,
			CollateralParams: cdptypes.CollateralParams{
				{
					Denom:                            denom,
					Type:                             asset + "-a",
					LiquidationRatio:                 liquidationRatio,
					DebtLimit:                        sdk.NewInt64Coin("usdx", 1000000000000),
					StabilityFee:                     sdkmath.LegacyMustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:               sdkmath.LegacyMustNewDecFromStr("0.05"),
					AuctionSize:                      sdkmath.NewInt(100),
					SpotMarketID:                     asset + ":usd",
					LiquidationMarketID:              asset + ":usd",
					KeeperRewardPercentage:           sdkmath.LegacyMustNewDecFromStr("0.01"),
					CheckCollateralizationIndexCount: sdkmath.NewInt(10),
					ConversionFactor:                 sdkmath.NewInt(6),
				},
			},
			DebtParam: cdptypes.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: sdkmath.NewInt(6),
				DebtFloor:        sdkmath.NewInt(10000000),
			},
		},
		StartingCdpID: cdptypes.DefaultCdpStartingID,
		DebtDenom:     cdptypes.DefaultDebtDenom,
		GovDenom:      cdptypes.DefaultGovDenom,
		CDPs:          cdptypes.CDPs{},
		PreviousAccumulationTimes: cdptypes.GenesisAccumulationTimes{
			cdptypes.NewGenesisAccumulationTime(asset+"-a", time.Time{}, sdkmath.LegacyOneDec()),
		},
		TotalPrincipals: cdptypes.GenesisTotalPrincipals{
			cdptypes.NewGenesisTotalPrincipal(asset+"-a", sdkmath.ZeroInt()),
		},
	}
	return app.GenesisState{cdptypes.ModuleName: cdc.MustMarshalJSON(&cdpGenesis)}
}
