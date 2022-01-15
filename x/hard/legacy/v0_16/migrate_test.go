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
	v015hard "github.com/kava-labs/kava/x/hard/legacy/v0_15"
	v016hard "github.com/kava-labs/kava/x/hard/types"
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
	file := filepath.Join("testdata", "v15-hard.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	var v15genstate v015hard.GenesisState
	err = s.legacyCdc.UnmarshalJSON(data, &v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(v15genstate)
	actual := s.cdc.MustMarshalJSON(genstate)

	file = filepath.Join("testdata", "v16-hard.json")
	expected, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_GenState() {
	v15genstate := v015hard.GenesisState{
		Params: v015hard.Params{
			MoneyMarkets: v015hard.MoneyMarkets{
				{
					Denom: "kava",
					BorrowLimit: v015hard.BorrowLimit{
						HasMaxLimit:  true,
						MaximumLimit: sdk.MustNewDecFromStr("0.1"),
						LoanToValue:  sdk.MustNewDecFromStr("0.2"),
					},
					SpotMarketID:     "spot-market-id",
					ConversionFactor: sdk.NewInt(110),
					InterestRateModel: v015hard.InterestRateModel{
						BaseRateAPY:    sdk.MustNewDecFromStr("0.1"),
						BaseMultiplier: sdk.MustNewDecFromStr("0.2"),
						Kink:           sdk.MustNewDecFromStr("0.3"),
						JumpMultiplier: sdk.MustNewDecFromStr("0.4"),
					},
					ReserveFactor:          sdk.MustNewDecFromStr("0.5"),
					KeeperRewardPercentage: sdk.MustNewDecFromStr("0.6"),
				},
			},
		},
		PreviousAccumulationTimes: v015hard.GenesisAccumulationTimes{
			{
				CollateralType:           "kava",
				PreviousAccumulationTime: time.Date(1998, time.January, 1, 12, 0, 0, 1, time.UTC),
				SupplyInterestFactor:     sdk.MustNewDecFromStr("0.1"),
				BorrowInterestFactor:     sdk.MustNewDecFromStr("0.2"),
			},
		},
		Deposits: v015hard.Deposits{
			{
				Depositor: s.addresses[0],
				Amount:    sdk.NewCoins(sdk.NewCoin("kava", sdk.NewInt(100))),
				Index: v015hard.SupplyInterestFactors{
					{
						Denom: "kava",
						Value: sdk.MustNewDecFromStr("1.12"),
					},
				},
			},
		},
		Borrows: v015hard.Borrows{
			{
				Borrower: s.addresses[1],
				Amount:   sdk.NewCoins(sdk.NewCoin("kava", sdk.NewInt(100))),
				Index: v015hard.BorrowInterestFactors{
					{
						Denom: "kava",
						Value: sdk.MustNewDecFromStr("1.12"),
					},
				},
			},
		},
		TotalSupplied: sdk.NewCoins(sdk.NewCoin("kava", sdk.NewInt(100))),
		TotalBorrowed: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
		TotalReserves: sdk.NewCoins(sdk.NewCoin("xrp", sdk.NewInt(300))),
	}
	expected := v016hard.GenesisState{
		Params: v016hard.Params{
			MoneyMarkets: v016hard.MoneyMarkets{
				{
					Denom: "kava",
					BorrowLimit: v016hard.BorrowLimit{
						HasMaxLimit:  true,
						MaximumLimit: sdk.MustNewDecFromStr("0.1"),
						LoanToValue:  sdk.MustNewDecFromStr("0.2"),
					},
					SpotMarketID:     "spot-market-id",
					ConversionFactor: sdk.NewInt(110),
					InterestRateModel: v016hard.InterestRateModel{
						BaseRateAPY:    sdk.MustNewDecFromStr("0.1"),
						BaseMultiplier: sdk.MustNewDecFromStr("0.2"),
						Kink:           sdk.MustNewDecFromStr("0.3"),
						JumpMultiplier: sdk.MustNewDecFromStr("0.4"),
					},
					ReserveFactor:          sdk.MustNewDecFromStr("0.5"),
					KeeperRewardPercentage: sdk.MustNewDecFromStr("0.6"),
				},
				{
					Denom: UATOM_IBC_DENOM,
					BorrowLimit: v016hard.BorrowLimit{
						HasMaxLimit:  true,
						MaximumLimit: sdk.NewDec(25000000000),
						LoanToValue:  sdk.MustNewDecFromStr("0.5"),
					},
					SpotMarketID:     "atom:usd:30",
					ConversionFactor: sdk.NewInt(1000000),
					InterestRateModel: v016hard.InterestRateModel{
						BaseRateAPY:    sdk.ZeroDec(),
						BaseMultiplier: sdk.MustNewDecFromStr("0.05"),
						Kink:           sdk.MustNewDecFromStr("0.8"),
						JumpMultiplier: sdk.NewDec(5),
					},
					ReserveFactor:          sdk.MustNewDecFromStr("0.025"),
					KeeperRewardPercentage: sdk.MustNewDecFromStr("0.02"),
				},
			},
		},
		PreviousAccumulationTimes: v016hard.GenesisAccumulationTimes{
			{
				CollateralType:           "kava",
				PreviousAccumulationTime: time.Date(1998, time.January, 1, 12, 0, 0, 1, time.UTC),
				SupplyInterestFactor:     sdk.MustNewDecFromStr("0.1"),
				BorrowInterestFactor:     sdk.MustNewDecFromStr("0.2"),
			},
		},
		Deposits: v016hard.Deposits{
			{
				Depositor: s.addresses[0],
				Amount:    sdk.NewCoins(sdk.NewCoin("kava", sdk.NewInt(100))),
				Index: v016hard.SupplyInterestFactors{
					{
						Denom: "kava",
						Value: sdk.MustNewDecFromStr("1.12"),
					},
				},
			},
		},
		Borrows: v016hard.Borrows{
			{
				Borrower: s.addresses[1],
				Amount:   sdk.NewCoins(sdk.NewCoin("kava", sdk.NewInt(100))),
				Index: v016hard.BorrowInterestFactors{
					{
						Denom: "kava",
						Value: sdk.MustNewDecFromStr("1.12"),
					},
				},
			},
		},
		TotalSupplied: sdk.NewCoins(sdk.NewCoin("kava", sdk.NewInt(100))),
		TotalBorrowed: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
		TotalReserves: sdk.NewCoins(sdk.NewCoin("xrp", sdk.NewInt(300))),
	}
	genState := Migrate(v15genstate)
	s.Require().Equal(expected, *genState)
}

func TestHardMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
