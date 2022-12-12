package keeper_test

import (
	"testing"
	"time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/hard/types/mocks"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HooksTestSuite struct {
	suite.Suite
	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress

	tokenA string
	tokenB string
}

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

// The default state used by each test
func (suite *HooksTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	suite.tokenA = "ukava"
	suite.tokenB = "bnb"

	suite.app = app.NewTestApp()
	_, suite.addrs = app.GeneratePrivKeyAddressPairs(2)
	keeper := suite.app.GetHardKeeper()
	suite.ctx = suite.app.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.keeper = keeper

	hardGS := types.NewGenesisState(types.NewParams(
		types.MoneyMarkets{
			types.NewMoneyMarket(
				"ukava",
				types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
				"ukava:usd",         // Market ID
				sdk.NewInt(KAVA_CF), // Conversion Factor
				types.NewInterestRateModel(
					sdk.MustNewDecFromStr("0.05"),
					sdk.MustNewDecFromStr("2"),
					sdk.MustNewDecFromStr("0.8"),
					sdk.MustNewDecFromStr("10"),
				),
				sdk.MustNewDecFromStr("0.05"),
				sdk.ZeroDec(), // Keeper Reward Percentage
			),
			types.NewMoneyMarket(
				"bnb",
				types.NewBorrowLimit(false, sdk.NewDec(100000000*BNB_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
				"bnb:usd",          // Market ID
				sdk.NewInt(BNB_CF), // Conversion Factor
				types.NewInterestRateModel(
					sdk.MustNewDecFromStr("0.05"),
					sdk.MustNewDecFromStr("2"),
					sdk.MustNewDecFromStr("0.8"),
					sdk.MustNewDecFromStr("10"),
				),
				sdk.MustNewDecFromStr("0.05"),
				sdk.ZeroDec(), // Keeper Reward Percentage
			),
		},
		sdk.NewDec(10),
	), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
		types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
	)

	pricefeedGS := pricefeedtypes.GenesisState{
		Params: pricefeedtypes.Params{
			Markets: []pricefeedtypes.Market{
				{MarketID: "ukava:usd", BaseAsset: "ukava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrpb:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "busd:usd", BaseAsset: "btcb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeedtypes.PostedPrice{
			{
				MarketID:      "ukava:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("1.5"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "busd:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("1.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrpb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("2.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("200.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}

	suite.app = suite.app.InitializeFromGenesisStates(
		app.GenesisState{
			pricefeedtypes.ModuleName: suite.app.AppCodec().MustMarshalJSON(&pricefeedGS),
			types.ModuleName:          suite.app.AppCodec().MustMarshalJSON(&hardGS),
		})

	balance := sdk.NewCoins(
		sdk.NewCoin(suite.tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(suite.tokenB, sdk.NewInt(1000e6)),
	)

	suite.Require().NoError(suite.app.FundAccount(suite.ctx, suite.addrs[0], balance))
}

func (suite *HooksTestSuite) TestHooks_DepositBorrowAndWithdraw() {
	suite.keeper.ClearHooks()
	hardHooks := mocks.NewHARDHooks(suite.T())
	suite.keeper.SetHooks(hardHooks)

	balance := sdk.NewCoins(
		sdk.NewCoin(suite.tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(suite.tokenB, sdk.NewInt(1000e6)),
	)

	suite.Require().NoError(suite.app.FundAccount(suite.ctx, suite.addrs[0], balance))
	suite.Require().NoError(suite.app.FundAccount(suite.ctx, suite.addrs[1], balance))

	depositor_1 := suite.addrs[0]
	depositor_2 := suite.addrs[1]

	depositA := sdk.NewCoin(suite.tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(suite.tokenB, sdk.NewInt(50e6))

	borrowA := sdk.NewCoin(suite.tokenB, sdk.NewInt(5e6))

	suite.Run("deposit 1", func() {
		interestFactors := types.SupplyInterestFactors{}
		interestFactors = interestFactors.SetInterestFactor(depositA.Denom, sdk.OneDec())
		expectedDeposit := types.NewDeposit(depositor_1, cs(depositA), interestFactors)

		// first deposit creates deposit - calls AfterDepositCreated with initial shares
		hardHooks.On("AfterDepositCreated", suite.ctx, expectedDeposit).Once()
		err := suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositA))
		suite.Require().NoError(err)

		// second deposit adds to deposit - calls BeforeDepositModified
		// shares given are the initial shares, along with a slice that includes new deposit denoms
		hardHooks.On("BeforeDepositModified", suite.ctx,
			expectedDeposit,          // old deposit
			[]string{depositB.Denom}, // new deposit denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositB))
		suite.Require().NoError(err)

		// get the shares from the store from the last deposit
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_1)
		suite.Require().True(found)

		// third deposit adds to deposit - calls BeforeDepositModified
		// shares given are the shares added in previous deposit, not the shares added to the deposit now
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,       // previous deposit
			[]string(nil), // no new denoms
		).Once()

		err = suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositB))
		suite.Require().NoError(err)
	})

	suite.Run("deposit 2", func() {
		interestFactors := types.SupplyInterestFactors{}
		interestFactors = interestFactors.SetInterestFactor(depositA.Denom, sdk.OneDec())
		expectedDeposit := types.NewDeposit(depositor_2, cs(depositA), interestFactors)

		// first deposit creates deposit - calls BeforeDepositModified with initial shares
		hardHooks.On("AfterDepositCreated", suite.ctx, expectedDeposit).Once()
		err := suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositA))
		suite.Require().NoError(err)

		// second deposit adds to deposit - calls BeforeDepositModified
		// shares given are the initial shares, along with a slice that includes new deposit denoms
		hardHooks.On("BeforeDepositModified", suite.ctx,
			expectedDeposit,          // old deposit
			[]string{depositB.Denom}, // new deposit denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositB))
		suite.Require().NoError(err)

		// get the shares from the store from the last deposit
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		// third deposit adds to deposit - calls BeforeDepositModified
		// shares given are the shares added in previous deposit, not the shares added to the deposit now
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,       // previous deposit
			[]string(nil), // no new denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositB))
		suite.Require().NoError(err)
	})

	suite.Run("borrow", func() {
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_1)
		suite.Require().True(found)

		suite.T().Logf("deposit: %v", deposit)

		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,       // previous deposit
			[]string(nil), // no new denoms when borrowing
		).Once()

		interestFactors := types.BorrowInterestFactors{}
		interestFactors = interestFactors.SetInterestFactor(borrowA.Denom, sdk.OneDec())
		expectedBorrow := types.NewBorrow(depositor_1, cs(borrowA), interestFactors)

		hardHooks.On("AfterBorrowCreated", suite.ctx,
			expectedBorrow, // new borrow
		).Once()
		err := suite.keeper.Borrow(suite.ctx, depositor_1, cs(borrowA))
		suite.Require().NoError(err)
	})

	// Depositor 2 borrows but does not repay
	suite.Run("borrow 2", func() {
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		suite.T().Logf("deposit: %v", deposit)

		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,       // previous deposit
			[]string(nil), // no new denoms when borrowing
		).Once()

		interestFactors := types.BorrowInterestFactors{}
		interestFactors = interestFactors.SetInterestFactor(borrowA.Denom, sdk.OneDec())
		expectedBorrow := types.NewBorrow(depositor_2, cs(borrowA), interestFactors)

		hardHooks.On("AfterBorrowCreated", suite.ctx,
			expectedBorrow, // new borrow
		).Once()
		err := suite.keeper.Borrow(suite.ctx, depositor_2, cs(borrowA))
		suite.Require().NoError(err)
	})

	suite.Run("repay 1", func() {
		borrow, found := suite.keeper.GetBorrow(suite.ctx, depositor_1)
		suite.Require().True(found)

		hardHooks.On("BeforeBorrowModified", suite.ctx,
			borrow,
			[]string(nil),
		).Once()

		err := suite.keeper.Repay(suite.ctx, depositor_1, depositor_1, cs(borrowA))
		suite.Require().NoError(err)
	})

	suite.Run("withdraw full", func() {
		// test hooks with a full withdraw of all shares
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_1)
		suite.Require().True(found)
		// all shares given to BeforeDepositModified
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,
			[]string(nil),
		).Once()
		err := suite.keeper.Withdraw(suite.ctx, depositor_1, deposit.Amount)
		suite.Require().NoError(err)
	})

	suite.Run("withdraw partial", func() {
		borrow, found := suite.keeper.GetBorrow(suite.ctx, depositor_2)
		suite.Require().True(found)

		hardHooks.On("BeforeBorrowModified", suite.ctx,
			borrow,
			[]string(nil),
		).Once()

		// test hooks on partial withdraw, WITH borrow still outstanding
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		partialWithdraw := sdk.NewCoin(deposit.Amount[0].Denom, deposit.Amount[0].Amount.QuoRaw(3))
		// all shares given to before deposit modified even with partial withdraw
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,
			[]string(nil),
		).Once()
		err := suite.keeper.Withdraw(suite.ctx, depositor_2, cs(partialWithdraw))
		suite.Require().NoError(err)

		hardHooks.On("BeforeBorrowModified", suite.ctx,
			borrow,
			[]string(nil),
		).Once()

		// test hooks on second partial withdraw
		deposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)
		partialWithdraw = sdk.NewCoin(deposit.Amount[1].Denom, deposit.Amount[1].Amount.QuoRaw(2))
		// all shares given to before deposit modified even with partial withdraw
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,
			[]string(nil),
		).Once()
		err = suite.keeper.Withdraw(suite.ctx, depositor_2, cs(partialWithdraw))
		suite.Require().NoError(err)

		// Repay borrow to before withdraw all shares
		hardHooks.On("BeforeBorrowModified", suite.ctx,
			borrow,
			[]string(nil),
		).Once()
		err = suite.keeper.Repay(suite.ctx, depositor_2, depositor_2, cs(borrowA))
		suite.Require().NoError(err)

		// test hooks withdraw all shares with second depositor
		deposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		// all shares given to before deposit modified even with partial withdraw
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,
			[]string(nil),
		).Once()
		err = suite.keeper.Withdraw(suite.ctx, depositor_2, deposit.Amount)
		suite.Require().NoError(err)
	})
}

func (suite *HooksTestSuite) TestHooks_NoPanicsOnNilHooks() {
	suite.keeper.ClearHooks()

	depositA := sdk.NewCoin(suite.tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(suite.tokenB, sdk.NewInt(50e6))

	// deposit create pool should not panic when hooks are not set
	err := suite.keeper.Deposit(suite.ctx, suite.addrs[0], cs(depositA, depositB))
	suite.Require().NoError(err)

	// existing deposit should not panic with hooks are not set
	err = suite.keeper.Deposit(suite.ctx, suite.addrs[0], cs(depositB))
	suite.Require().NoError(err)

	// withdraw of shares should not panic when hooks are not set
	shareRecord, found := suite.keeper.GetDeposit(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	err = suite.keeper.Withdraw(suite.ctx, suite.addrs[0], shareRecord.Amount)
	suite.Require().NoError(err)
}

func (suite *HooksTestSuite) TestHooks_HookOrdering() {
	suite.keeper.ClearHooks()

	hardHooks := mocks.NewHARDHooks(suite.T())
	suite.keeper.SetHooks(hardHooks)

	depositA := sdk.NewCoin(suite.tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(suite.tokenB, sdk.NewInt(50e6))

	_, foundValue := suite.keeper.GetSupplyInterestFactor(suite.ctx, depositA.Denom)
	suite.Require().False(foundValue)

	interestFactors := types.SupplyInterestFactors{}
	interestFactors = interestFactors.SetInterestFactor(depositA.Denom, sdk.OneDec())

	expectedDeposit := types.NewDeposit(suite.addrs[0], cs(depositA), interestFactors)

	hardHooks.On("AfterDepositCreated", suite.ctx,
		expectedDeposit, // new deposit created
	).Run(func(args mock.Arguments) {
		_, found := suite.keeper.GetDeposit(suite.ctx, suite.addrs[0])
		suite.Require().True(found, "expected after hook to be called after deposit is updated")
	})
	err := suite.keeper.Deposit(suite.ctx, suite.addrs[0], cs(depositA))
	suite.Require().NoError(err)

	hardHooks.On("BeforeDepositModified", suite.ctx,
		types.NewDeposit(suite.addrs[0], cs(depositA), interestFactors), // existing deposit modified
		[]string{depositB.Denom},
	).Run(func(args mock.Arguments) {
		deposit, found := suite.keeper.GetDeposit(suite.ctx, suite.addrs[0])
		suite.Require().True(found, "expected deposit to exist")
		suite.Equal(cs(depositA), deposit.Amount, "expected hook to be called before deposit is updated")
	})
	err = suite.keeper.Deposit(suite.ctx, suite.addrs[0], cs(depositB))
	suite.Require().NoError(err)

	deposit, found := suite.keeper.GetDeposit(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	hardHooks.
		On(
			"BeforeDepositModified",
			suite.ctx,
			deposit,       // existing deposit modified
			[]string(nil), // no new denoms when withdrawing
		).
		Run(func(args mock.Arguments) {
			existingDeposit, found := suite.keeper.GetDeposit(suite.ctx, suite.addrs[0])
			suite.Require().True(found, "expected share record to exist")
			suite.Equal(deposit, existingDeposit, "expected hook to be called before shares are updated")
		})

	err = suite.keeper.Withdraw(suite.ctx, suite.addrs[0], deposit.Amount)
	suite.Require().NoError(err)
}
