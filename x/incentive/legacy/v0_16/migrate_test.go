package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	app "github.com/kava-labs/kava/app"
	v015incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_15"
	v016incentive "github.com/kava-labs/kava/x/incentive/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses []sdk.AccAddress
	cdc       codec.Codec
	legacyCdc *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	file := filepath.Join("testdata", "v15-incentive.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	var v15genstate v015incentive.GenesisState
	err = s.legacyCdc.UnmarshalJSON(data, &v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(v15genstate)
	actual := s.cdc.MustMarshalJSON(genstate)

	file = filepath.Join("testdata", "v16-incentive.json")
	expected, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_GenState() {
	v15genstate := v015incentive.GenesisState{
		Params: v015incentive.Params{
			ClaimEnd: time.Date(2020, time.March, 1, 2, 0, 0, 0, time.UTC),
			USDXMintingRewardPeriods: []v015incentive.RewardPeriod{
				{
					Active:           true,
					CollateralType:   "usdx",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoin("usdx", sdk.NewInt(10)),
				},
			},
			HardSupplyRewardPeriods: v015incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "usdx",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(10))),
				},
			},
			HardBorrowRewardPeriods: v015incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "bnb",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10))),
				},
			},
			DelegatorRewardPeriods: v015incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "bnb",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10))),
				},
			},
			SwapRewardPeriods: v015incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "bnb",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10))),
				},
			},
			ClaimMultipliers: v015incentive.MultipliersPerDenom{
				{
					Denom: "usdx",
					Multipliers: v015incentive.Multipliers{
						{
							Name:         v015incentive.Small,
							MonthsLockup: 6,
							Factor:       sdk.MustNewDecFromStr("0.5"),
						},
						{
							Name:         v015incentive.Large,
							MonthsLockup: 12,
							Factor:       sdk.MustNewDecFromStr("0.8"),
						},
						{
							Name:         v015incentive.Medium,
							MonthsLockup: 9,
							Factor:       sdk.MustNewDecFromStr("0.7"),
						},
					},
				},
			},
		},
		USDXRewardState: v015incentive.GenesisRewardState{
			AccumulationTimes: v015incentive.AccumulationTimes{
				{
					CollateralType:           "usdx",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v015incentive.MultiRewardIndexes{
				{
					CollateralType: "usdx",
					RewardIndexes: []v015incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		USDXMintingClaims: v015incentive.USDXMintingClaims{
			{
				BaseClaim: v015incentive.BaseClaim{
					Owner:  s.addresses[0],
					Reward: sdk.NewCoin("usdx", sdk.NewInt(100)),
				},
				RewardIndexes: v015incentive.RewardIndexes{
					{
						CollateralType: "kava",
						RewardFactor:   sdk.MustNewDecFromStr("0.5"),
					},
				},
			},
		},
		HardSupplyRewardState: v015incentive.GenesisRewardState{
			AccumulationTimes: v015incentive.AccumulationTimes{
				{
					CollateralType:           "usdx",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v015incentive.MultiRewardIndexes{
				{
					CollateralType: "usdx",
					RewardIndexes: []v015incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		HardBorrowRewardState: v015incentive.GenesisRewardState{
			AccumulationTimes: v015incentive.AccumulationTimes{
				{
					CollateralType:           "hard",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v015incentive.MultiRewardIndexes{
				{
					CollateralType: "hard",
					RewardIndexes: []v015incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		DelegatorRewardState: v015incentive.GenesisRewardState{
			AccumulationTimes: v015incentive.AccumulationTimes{
				{
					CollateralType:           "usdx",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v015incentive.MultiRewardIndexes{
				{
					CollateralType: "usdx",
					RewardIndexes: []v015incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		SwapRewardState: v015incentive.GenesisRewardState{
			AccumulationTimes: v015incentive.AccumulationTimes{
				{
					CollateralType:           "swap",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v015incentive.MultiRewardIndexes{
				{
					CollateralType: "swap",
					RewardIndexes: []v015incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.25"),
						},
					},
				},
			},
		},
		HardLiquidityProviderClaims: v015incentive.HardLiquidityProviderClaims{
			{
				BaseMultiClaim: v015incentive.BaseMultiClaim{
					Owner:  s.addresses[1],
					Reward: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
				},
				SupplyRewardIndexes: v015incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v015incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
				BorrowRewardIndexes: v015incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v015incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
			},
		},
		DelegatorClaims: v015incentive.DelegatorClaims{
			{
				BaseMultiClaim: v015incentive.BaseMultiClaim{
					Owner:  s.addresses[1],
					Reward: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
				},
				RewardIndexes: v015incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v015incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
			},
		},
		SwapClaims: v015incentive.SwapClaims{
			{
				BaseMultiClaim: v015incentive.BaseMultiClaim{
					Owner:  s.addresses[1],
					Reward: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
				},
				RewardIndexes: v015incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v015incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
			},
		},
	}
	expected := v016incentive.GenesisState{
		USDXRewardState: v016incentive.GenesisRewardState{
			AccumulationTimes: v016incentive.AccumulationTimes{
				{
					CollateralType:           "usdx",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v016incentive.MultiRewardIndexes{
				{
					CollateralType: "usdx",
					RewardIndexes: []v016incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		Params: v016incentive.Params{
			ClaimEnd: time.Date(2020, time.March, 1, 2, 0, 0, 0, time.UTC),
			USDXMintingRewardPeriods: []v016incentive.RewardPeriod{
				{
					Active:           true,
					CollateralType:   "usdx",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoin("usdx", sdk.NewInt(10)),
				},
			},
			HardSupplyRewardPeriods: v016incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "usdx",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(10))),
				},
			},
			HardBorrowRewardPeriods: v016incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "bnb",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10))),
				},
			},
			DelegatorRewardPeriods: v016incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "bnb",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10))),
				},
			},
			SwapRewardPeriods: v016incentive.MultiRewardPeriods{
				{
					Active:           true,
					CollateralType:   "bnb",
					Start:            time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					End:              time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					RewardsPerSecond: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10))),
				},
			},
			ClaimMultipliers: []v016incentive.MultipliersPerDenom{
				{
					Denom: "usdx",
					Multipliers: v016incentive.Multipliers{
						{
							Name:         "small",
							MonthsLockup: 6,
							Factor:       sdk.MustNewDecFromStr("0.5"),
						},
						{
							Name:         "large",
							MonthsLockup: 12,
							Factor:       sdk.MustNewDecFromStr("0.8"),
						},
						{
							Name:         "medium",
							MonthsLockup: 9,
							Factor:       sdk.MustNewDecFromStr("0.7"),
						},
					},
				},
			},
		},
		USDXMintingClaims: v016incentive.USDXMintingClaims{
			{
				BaseClaim: v016incentive.BaseClaim{
					Owner:  s.addresses[0],
					Reward: sdk.NewCoin("usdx", sdk.NewInt(100)),
				},
				RewardIndexes: v016incentive.RewardIndexes{
					{
						CollateralType: "kava",
						RewardFactor:   sdk.MustNewDecFromStr("0.5"),
					},
				},
			},
		},
		HardSupplyRewardState: v016incentive.GenesisRewardState{
			AccumulationTimes: v016incentive.AccumulationTimes{
				{
					CollateralType:           "usdx",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v016incentive.MultiRewardIndexes{
				{
					CollateralType: "usdx",
					RewardIndexes: []v016incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		HardBorrowRewardState: v016incentive.GenesisRewardState{
			AccumulationTimes: v016incentive.AccumulationTimes{
				{
					CollateralType:           "hard",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v016incentive.MultiRewardIndexes{
				{
					CollateralType: "hard",
					RewardIndexes: []v016incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		DelegatorRewardState: v016incentive.GenesisRewardState{
			AccumulationTimes: v016incentive.AccumulationTimes{
				{
					CollateralType:           "usdx",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v016incentive.MultiRewardIndexes{
				{
					CollateralType: "usdx",
					RewardIndexes: []v016incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.15"),
						},
					},
				},
			},
		},
		SwapRewardState: v016incentive.GenesisRewardState{
			AccumulationTimes: v016incentive.AccumulationTimes{
				{
					CollateralType:           "swap",
					PreviousAccumulationTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			MultiRewardIndexes: v016incentive.MultiRewardIndexes{
				{
					CollateralType: "swap",
					RewardIndexes: []v016incentive.RewardIndex{
						{
							CollateralType: "bnb",
							RewardFactor:   sdk.MustNewDecFromStr("0.25"),
						},
					},
				},
			},
		},
		HardLiquidityProviderClaims: v016incentive.HardLiquidityProviderClaims{
			{
				BaseMultiClaim: v016incentive.BaseMultiClaim{
					Owner:  s.addresses[1],
					Reward: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
				},
				SupplyRewardIndexes: v016incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v016incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
				BorrowRewardIndexes: v016incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v016incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
			},
		},
		DelegatorClaims: v016incentive.DelegatorClaims{
			{
				BaseMultiClaim: v016incentive.BaseMultiClaim{
					Owner:  s.addresses[1],
					Reward: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
				},
				RewardIndexes: v016incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v016incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
			},
		},
		SwapClaims: v016incentive.SwapClaims{
			{
				BaseMultiClaim: v016incentive.BaseMultiClaim{
					Owner:  s.addresses[1],
					Reward: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
				},
				RewardIndexes: v016incentive.MultiRewardIndexes{
					{
						CollateralType: "bnb",
						RewardIndexes: []v016incentive.RewardIndex{
							{
								CollateralType: "bnb",
								RewardFactor:   sdk.MustNewDecFromStr("0.25"),
							},
						},
					},
				},
			},
		},
	}
	genState := Migrate(v15genstate)
	s.Require().Equal(expected, *genState)
}

func TestIncentiveMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
