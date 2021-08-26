package hard_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
)

type GenesisTestSuite struct {
	suite.Suite

	app     app.TestApp
	genTime time.Time
	ctx     sdk.Context
	keeper  hard.Keeper
	addrs   []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.genTime = tmtime.Canonical(time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC))
	suite.ctx = tApp.NewContext(true, abci.Header{Height: 1, Time: suite.genTime})
	suite.keeper = tApp.GetHardKeeper()
	suite.app = tApp

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.addrs = addrs
}

func (suite *GenesisTestSuite) Test_InitExportGenesis() {

	loanToValue, _ := sdk.NewDecFromStr("0.6")
	params := hard.NewParams(
		hard.MoneyMarkets{
			hard.NewMoneyMarket("ukava", hard.NewBorrowLimit(false, sdk.NewDec(1e15), loanToValue), "kava:usd", sdk.NewInt(1e6), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
		},
		sdk.NewDec(10),
	)

	deposits := hard.Deposits{
		hard.NewDeposit(
			suite.addrs[0],
			sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e8))), // 100 ukava
			hard.SupplyInterestFactors{
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

	borrows := hard.Borrows{
		hard.NewBorrow(
			suite.addrs[1],
			sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e7))), // 10 ukava
			hard.BorrowInterestFactors{
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
	accuralTimes := hard.GenesisAccumulationTimes{
		hard.NewGenesisAccumulationTime("ukava", suite.genTime, supplyInterestFactor, borrowInterestFactor),
	}

	hardGenesis := hard.NewGenesisState(
		params,
		accuralTimes,
		deposits,
		borrows,
		totalSupplied,
		totalBorrowed,
		nil,
	)

	suite.True(
		suite.NotPanics(
			func() {
				suite.app.InitializeFromGenesisStatesWithTime(
					suite.genTime,
					app.GenesisState{hard.ModuleName: hard.ModuleCdc.MustMarshalJSON(hardGenesis)},
				)
			},
		),
	)

	var expectedDeposits hard.Deposits
	for _, deposit := range deposits {
		// Calculate expected coin amount post sync
		var depositAmount sdk.Coins
		for _, coin := range deposit.Amount {
			expectedAmt := supplyInterestFactor.MulInt(coin.Amount).RoundInt()
			depositAmount = depositAmount.Add(sdk.NewCoin(coin.Denom, expectedAmt))
		}
		deposit.Amount = depositAmount
		// Calculate expected indexes post sync
		var indexes hard.SupplyInterestFactors
		for _, index := range deposit.Index {
			index.Value = supplyInterestFactor.Mul(index.Value)
			indexes = append(indexes, index)
		}
		deposit.Index = indexes
		expectedDeposits = append(expectedDeposits, deposit)
	}

	var expectedBorrows hard.Borrows
	for _, borrow := range borrows {
		// Calculate expected coin amount post sync
		var borrowAmount sdk.Coins
		for _, coin := range borrow.Amount {
			expectedAmt := borrowInterestFactor.MulInt(coin.Amount).RoundInt()
			borrowAmount = borrowAmount.Add(sdk.NewCoin(coin.Denom, expectedAmt))
		}
		borrow.Amount = borrowAmount
		// Calculate expected indexes post sync
		var indexes hard.BorrowInterestFactors
		for _, index := range borrow.Index {
			index.Value = borrowInterestFactor.Mul(index.Value)
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

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
