package hard_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app     app.TestApp
	genTime time.Time
	ctx     sdk.Context
	keeper  keeper.Keeper
	addrs   []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.genTime = tmtime.Canonical(time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC))
	suite.ctx = tApp.NewContext(true, tmproto.Header{Height: 1, Time: suite.genTime})
	suite.keeper = tApp.GetHardKeeper()
	suite.app = tApp

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.addrs = addrs
}

func (suite *GenesisTestSuite) Test_InitExportGenesis() {
	loanToValue, _ := sdk.NewDecFromStr("0.6")
	params := types.NewParams(
		types.MoneyMarkets{
			types.NewMoneyMarket(
				"ukava",
				types.NewBorrowLimit(
					false,
					sdk.NewDec(1e15),
					loanToValue,
				),
				"kava:usd",
				sdkmath.NewInt(1e6),
				types.NewInterestRateModel(
					sdk.MustNewDecFromStr("0.05"),
					sdk.MustNewDecFromStr("2"),
					sdk.MustNewDecFromStr("0.8"),
					sdk.MustNewDecFromStr("10"),
				),
				sdk.MustNewDecFromStr("0.05"),
				sdk.ZeroDec(),
			),
		},
		sdk.NewDec(10),
	)

	deposits := types.Deposits{
		types.NewDeposit(
			suite.addrs[0],
			sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e8))), // 100 ukava
			types.SupplyInterestFactors{
				{
					Denom: "ukava",
					Value: sdk.NewDec(1),
				},
			},
		),
	}

	var totalSupplied sdk.Coins
	for _, deposit := range deposits {
		totalSupplied = totalSupplied.Add(deposit.Amount...)
	}

	borrows := types.Borrows{
		types.NewBorrow(
			suite.addrs[1],
			sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e7))), // 10 ukava
			types.BorrowInterestFactors{
				{
					Denom: "ukava",
					Value: sdk.NewDec(1),
				},
			},
		),
	}

	var totalBorrowed sdk.Coins
	for _, borrow := range borrows {
		totalBorrowed = totalBorrowed.Add(borrow.Amount...)
	}

	supplyInterestFactor := sdk.MustNewDecFromStr("1.0001")
	borrowInterestFactor := sdk.MustNewDecFromStr("1.1234")
	accuralTimes := types.GenesisAccumulationTimes{
		types.NewGenesisAccumulationTime("ukava", suite.genTime, supplyInterestFactor, borrowInterestFactor),
	}

	hardGenesis := types.NewGenesisState(
		params,
		accuralTimes,
		deposits,
		borrows,
		totalSupplied,
		totalBorrowed,
		sdk.Coins{},
	)

	suite.NotPanics(
		func() {
			suite.app.InitializeFromGenesisStatesWithTime(
				suite.genTime,
				app.GenesisState{types.ModuleName: suite.app.AppCodec().MustMarshalJSON(&hardGenesis)},
			)
		},
	)

	var expectedDeposits types.Deposits
	for _, deposit := range deposits {
		// Deposit coin amounts
		var depositAmount sdk.Coins
		for _, coin := range deposit.Amount {
			accrualTime, found := getGenesisAccumulationTime(coin.Denom, accuralTimes)
			if !found {
				panic(fmt.Sprintf("accrual time not found %s", coin.Denom))
			}
			expectedAmt := accrualTime.SupplyInterestFactor.MulInt(coin.Amount).RoundInt()
			depositAmount = depositAmount.Add(sdk.NewCoin(coin.Denom, expectedAmt))
		}
		deposit.Amount = depositAmount
		// Deposit interest factor indexes
		var indexes types.SupplyInterestFactors
		for _, index := range deposit.Index {
			accrualTime, found := getGenesisAccumulationTime(index.Denom, accuralTimes)
			if !found {
				panic(fmt.Sprintf("accrual time not found %s", index.Denom))
			}
			index.Value = accrualTime.SupplyInterestFactor
			indexes = append(indexes, index)
		}
		deposit.Index = indexes
		expectedDeposits = append(expectedDeposits, deposit)
	}

	var expectedBorrows types.Borrows
	for _, borrow := range borrows {
		// Borrow coin amounts
		var borrowAmount sdk.Coins
		for _, coin := range borrow.Amount {
			accrualTime, found := getGenesisAccumulationTime(coin.Denom, accuralTimes)
			if !found {
				panic(fmt.Sprintf("accrual time not found %s", coin.Denom))
			}
			expectedAmt := accrualTime.BorrowInterestFactor.MulInt(coin.Amount).RoundInt()
			borrowAmount = borrowAmount.Add(sdk.NewCoin(coin.Denom, expectedAmt))

		}
		borrow.Amount = borrowAmount
		// Borrow interest factor indexes
		var indexes types.BorrowInterestFactors
		for _, index := range borrow.Index {
			accrualTime, found := getGenesisAccumulationTime(index.Denom, accuralTimes)
			if !found {
				panic(fmt.Sprintf("accrual time not found %s", index.Denom))
			}
			index.Value = accrualTime.BorrowInterestFactor
			indexes = append(indexes, index)
		}
		borrow.Index = indexes
		expectedBorrows = append(expectedBorrows, borrow)
	}

	expectedGenesis := hardGenesis
	expectedGenesis.Deposits = expectedDeposits
	expectedGenesis.Borrows = expectedBorrows
	exportedGenesis := hard.ExportGenesis(suite.ctx, suite.keeper)
	suite.Equal(expectedGenesis, exportedGenesis)
}

func getGenesisAccumulationTime(denom string, ts types.GenesisAccumulationTimes) (types.GenesisAccumulationTime, bool) {
	for _, t := range ts {
		if t.CollateralType == denom {
			return t, true
		}
	}
	return types.GenesisAccumulationTime{}, false
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
