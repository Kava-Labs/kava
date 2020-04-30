package types_test

import (
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/cdp/types"
)

type ParamsTestSuite struct {
	suite.Suite
}

func (suite *ParamsTestSuite) SetupTest() {
}

func (suite *ParamsTestSuite) TestParamValidation() {
	type args struct {
		globalDebtLimit  sdk.Coin
		collateralParams types.CollateralParams
		debtParam        types.DebtParam
		surplusThreshold sdk.Int
		debtThreshold    sdk.Int
		distributionFreq time.Duration
		breaker          bool
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}

	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			name: "default",
			args: args{
				globalDebtLimit:  types.DefaultGlobalDebt,
				collateralParams: types.DefaultCollateralParams,
				debtParam:        types.DefaultDebtParam,
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "valid single-collateral",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 4000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "invalid single-collateral mismatched debt denoms",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 4000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "susd",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "does not match global debt denom",
			},
		},
		{
			name: "invalid single-collateral over debt limit",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 1000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "exceeds global debt limit",
			},
		},
		{
			name: "valid multi-collateral",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 4000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
					{
						Denom:              "xrp",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x21,
						MarketID:           "xrp:usd",
						ConversionFactor:   sdk.NewInt(6),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "invalid multi-collateral over debt limit",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
					{
						Denom:              "xrp",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x21,
						MarketID:           "xrp:usd",
						ConversionFactor:   sdk.NewInt(6),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "sum of collateral debt limits",
			},
		},
		{
			name: "invalid multi-collateral multiple debt denoms",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 4000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
					{
						Denom:              "xrp",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("susd", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x21,
						MarketID:           "xrp:usd",
						ConversionFactor:   sdk.NewInt(6),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "does not match global debt limit denom",
			},
		},
		{
			name: "invalid collateral params empty denom",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "collateral denom invalid",
			},
		},
		{
			name: "invalid collateral params empty market id",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "market id cannot be blank",
			},
		},
		{
			name: "invalid collateral params duplicate denom",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x21,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "duplicate collateral denom",
			},
		},
		{
			name: "invalid collateral params duplicate prefix",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
					{
						Denom:              "xrp",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "xrp:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "duplicate prefix for collateral denom",
			},
		},
		{
			name: "invalid collateral params nil debt limit",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.Coin{},
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "debt limit for all collaterals should be positive",
			},
		},
		{
			name: "invalid collateral params liquidation ratio out of range",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("1.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "liquidation penalty should be between 0 and 1",
			},
		},
		{
			name: "invalid collateral params auction size zero",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.ZeroInt(),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "auction size should be positive",
			},
		},
		{
			name: "invalid collateral params stability fee out of range",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 1000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.1"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "stability fee must be ≥ 1.0",
			},
		},
		{
			name: "invalid debt param empty denom",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "debt denom invalid",
			},
		},
		{
			name: "invalid debt param savings rate out of range",
			args: args{
				globalDebtLimit: sdk.NewInt64Coin("usdx", 2000000000000),
				collateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewInt64Coin("usdx", 2000000000000),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000000),
						Prefix:             0x20,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				debtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("1.05"),
				},
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "savings rate should be between 0 and 1",
			},
		},
		{
			name: "nil debt limit",
			args: args{
				globalDebtLimit:  sdk.Coin{},
				collateralParams: types.DefaultCollateralParams,
				debtParam:        types.DefaultDebtParam,
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: global debt limit",
			},
		},
		{
			name: "zero savings distribution frequency",
			args: args{
				globalDebtLimit:  types.DefaultGlobalDebt,
				collateralParams: types.DefaultCollateralParams,
				debtParam:        types.DefaultDebtParam,
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: time.Second * 0,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "savings distribution frequency should be positive",
			},
		},
		{
			name: "zero surplus auction",
			args: args{
				globalDebtLimit:  types.DefaultGlobalDebt,
				collateralParams: types.DefaultCollateralParams,
				debtParam:        types.DefaultDebtParam,
				surplusThreshold: sdk.ZeroInt(),
				debtThreshold:    types.DefaultDebtThreshold,
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "surplus auction threshold should be positive",
			},
		},
		{
			name: "zero debt auction",
			args: args{
				globalDebtLimit:  types.DefaultGlobalDebt,
				collateralParams: types.DefaultCollateralParams,
				debtParam:        types.DefaultDebtParam,
				surplusThreshold: types.DefaultSurplusThreshold,
				debtThreshold:    sdk.ZeroInt(),
				distributionFreq: types.DefaultSavingsDistributionFrequency,
				breaker:          types.DefaultCircuitBreaker,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "debt auction threshold should be positive",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := types.NewParams(tc.args.globalDebtLimit, tc.args.collateralParams, tc.args.debtParam, tc.args.surplusThreshold, tc.args.debtThreshold, tc.args.distributionFreq, tc.args.breaker)
			err := params.Validate()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
