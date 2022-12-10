package keeper_test

import (
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
}

// The default state used by each test
func (suite *HooksTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	suite.app = app.NewTestApp()
	ctx := suite.app.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	keeper := suite.app.GetHardKeeper()
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs

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
}

func (suite *HooksTestSuite) TestHooks_DepositBorrowAndWithdraw() {
	suite.keeper.ClearHooks()
	hardHooks := mocks.NewHARDHooks(suite.T())
	suite.keeper.SetHooks(hardHooks)

	tokenA := "ukava"
	tokenB := "bnb"

	suite.keeper.SetParams(suite.ctx, types.NewParams(
		types.MoneyMarkets{
			types.NewMoneyMarket("ukava",
				types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
				"kava:usd",          // Market ID
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
			types.NewMoneyMarket("bnb",
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
				sdk.ZeroDec()), // Keeper Reward Percentage
		},
		sdk.NewDec(10),
	))

	balance := sdk.NewCoins(
		sdk.NewCoin(tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(tokenB, sdk.NewInt(1000e6)),
	)

	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	suite.Require().NoError(suite.app.FundAccount(suite.ctx, addrs[0], balance))
	suite.Require().NoError(suite.app.FundAccount(suite.ctx, addrs[1], balance))

	depositor_1 := addrs[0]
	depositor_2 := addrs[1]

	depositA := sdk.NewCoin(tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(tokenB, sdk.NewInt(50e6))

	suite.Run("deposit 1", func() {
		// first deposit creates deposit - calls AfterDepositCreated with initial shares
		hardHooks.On("AfterDepositCreated", suite.ctx, types.NewDeposit(depositor_1, cs(depositA), nil)).Once()
		err := suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositA))
		suite.Require().NoError(err)

		// second deposit adds to deposit - calls BeforeDepositModified
		// shares given are the initial shares, along with a slice that includes new deposit denoms
		hardHooks.On("BeforeDepositModified", suite.ctx,
			types.NewDeposit(depositor_1, cs(depositA), nil), // old deposit
			[]string{depositB.Denom},                         // new deposit denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositB))
		suite.Require().NoError(err)

		// get the shares from the store from the last deposit
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_1)
		suite.Require().True(found)

		// third deposit adds to deposit - calls BeforeDepositModified
		// shares given are the shares added in previous deposit, not the shares added to the deposit now
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,    // previous deposit
			[]string{}, // no new denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositB))
		suite.Require().NoError(err)
	})

	suite.Run("deposit 2", func() {
		// first deposit creates deposit - calls BeforeDepositModified with initial shares
		hardHooks.On("AfterDepositCreated", suite.ctx, types.NewDeposit(depositor_2, cs(depositA), nil)).Once()
		err := suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositA))
		suite.Require().NoError(err)

		// second deposit adds to deposit - calls BeforeDepositModified
		// shares given are the initial shares, along with a slice that includes new deposit denoms
		hardHooks.On("BeforeDepositModified", suite.ctx,
			types.NewDeposit(depositor_2, cs(depositA), nil), // old deposit
			[]string{depositB.Denom},                         // new deposit denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositB))
		suite.Require().NoError(err)

		// get the shares from the store from the last deposit
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		// third deposit adds to deposit - calls BeforeDepositModified
		// shares given are the shares added in previous deposit, not the shares added to the deposit now
		hardHooks.On("BeforeDepositModified", suite.ctx,
			deposit,    // previous deposit
			[]string{}, // no new denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositB))
		suite.Require().NoError(err)
	})

	suite.Run("borrow", func() {

	})

	suite.Run("repay", func() {

	})

	suite.Run("withdraw full", func() {
		// test hooks with a full withdraw of all shares
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_1)
		suite.Require().True(found)
		// all shares given to BeforeDepositModified
		hardHooks.On("BeforeDepositModified", suite.ctx, deposit, []string{}).Once()
		err := suite.keeper.Withdraw(suite.ctx, depositor_1, deposit.Amount)
		suite.Require().NoError(err)
	})

	suite.Run("withdraw partial", func() {
		// test hooks on partial withdraw
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		partialWithdraw := sdk.NewCoin(deposit.Amount[0].Denom, deposit.Amount[0].Amount.QuoRaw(3))
		// all shares given to before deposit modified even with partial withdraw
		hardHooks.On("BeforeDepositModified", suite.ctx, deposit, nil).Once()
		err := suite.keeper.Withdraw(suite.ctx, depositor_2, cs(partialWithdraw))
		suite.Require().NoError(err)

		// test hooks on second partial withdraw
		deposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)
		partialWithdraw = sdk.NewCoin(deposit.Amount[1].Denom, deposit.Amount[1].Amount.QuoRaw(2))
		// all shares given to before deposit modified even with partial withdraw
		hardHooks.On("BeforeDepositModified", suite.ctx, deposit, nil).Once()
		err = suite.keeper.Withdraw(suite.ctx, depositor_2, cs(partialWithdraw))
		suite.Require().NoError(err)

		// test hooks withdraw all shares with second depositor
		deposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		// all shares given to before deposit modified even with partial withdraw
		hardHooks.On("BeforeDepositModified", suite.ctx, deposit, nil).Once()
		err = suite.keeper.Withdraw(suite.ctx, depositor_2, deposit.Amount)
		suite.Require().NoError(err)
	})
}

func (suite *HooksTestSuite) TestHooks_NoPanicsOnNilHooks() {
	suite.keeper.ClearHooks()

	tokenA := "ukava"
	tokenB := "bnb"

	suite.keeper.SetParams(suite.ctx, types.NewParams(
		types.MoneyMarkets{
			types.NewMoneyMarket("ukava",
				types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
				"kava:usd",          // Market ID
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
			types.NewMoneyMarket("bnb",
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
				sdk.ZeroDec()), // Keeper Reward Percentage
		},
		sdk.NewDec(10),
	))

	balance := sdk.NewCoins(
		sdk.NewCoin(tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(tokenB, sdk.NewInt(1000e6)),
	)

	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	suite.Require().NoError(suite.app.FundAccount(suite.ctx, addrs[0], balance))

	depositA := sdk.NewCoin(tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(tokenB, sdk.NewInt(50e6))

	// deposit create pool should not panic when hooks are not set
	err := suite.keeper.Deposit(suite.ctx, addrs[0], cs(depositA, depositB))
	suite.Require().NoError(err)

	// existing deposit should not panic with hooks are not set
	err = suite.keeper.Deposit(suite.ctx, addrs[0], cs(depositB))
	suite.Require().NoError(err)

	// withdraw of shares should not panic when hooks are not set
	shareRecord, found := suite.keeper.GetDeposit(suite.ctx, addrs[0])
	suite.Require().True(found)

	err = suite.keeper.Withdraw(suite.ctx, addrs[0], shareRecord.Amount)
	suite.Require().NoError(err)
}

func (suite *HooksTestSuite) TestHooks_HookOrdering() {
	suite.keeper.ClearHooks()

	hardHooks := mocks.NewHARDHooks(suite.T())
	suite.keeper.SetHooks(hardHooks)

	tokenA := "ukava"
	tokenB := "bnb"

	balance := sdk.NewCoins(
		sdk.NewCoin(tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(tokenB, sdk.NewInt(1000e6)),
	)

	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	suite.Require().NoError(suite.app.FundAccount(suite.ctx, addrs[0], balance))

	depositA := sdk.NewCoin(tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(tokenB, sdk.NewInt(50e6))

	interestFactorValue, foundValue := suite.keeper.GetSupplyInterestFactor(suite.ctx, depositA.Denom)
	suite.Require().True(foundValue)

	interestFactors := types.SupplyInterestFactors{}
	interestFactors = interestFactors.SetInterestFactor(depositA.Denom, interestFactorValue)

	hardHooks.On("AfterDepositCreated", suite.ctx,
		types.NewDeposit(addrs[0], cs(depositA), interestFactors), // new deposit created
	).Run(func(args mock.Arguments) {
		_, found := suite.keeper.GetDeposit(suite.ctx, addrs[0])
		suite.Require().True(found, "expected after hook to be called after deposit is updated")
	})
	err := suite.keeper.Deposit(suite.ctx, addrs[0], cs(depositA))
	suite.Require().NoError(err)

	hardHooks.On("BeforeDepositModified", suite.ctx,
		types.NewDeposit(addrs[0], cs(depositA), interestFactors), // existing deposit modified
		[]string{depositB.Denom},
	).Run(func(args mock.Arguments) {
		deposit, found := suite.keeper.GetDeposit(suite.ctx, addrs[0])
		suite.Require().True(found, "expected deposit to exist")
		suite.Equal(cs(depositA), deposit.Amount, "expected hook to be called before deposit is updated")
	})
	err = suite.keeper.Deposit(suite.ctx, addrs[0], cs(depositB))
	suite.Require().NoError(err)

	deposit, found := suite.keeper.GetDeposit(suite.ctx, addrs[0])
	suite.Require().True(found)
	hardHooks.On("BeforeDepositModified", suite.ctx,
		deposit,    // existing deposit modified
		[]string{}, // no new denoms when withdrawing
	).Run(func(args mock.Arguments) {
		existingDeposit, found := suite.keeper.GetDeposit(suite.ctx, addrs[0])
		suite.Require().True(found, "expected share record to exist")
		suite.Equal(deposit, existingDeposit, "expected hook to be called before shares are updated")
	})
	err = suite.keeper.Withdraw(suite.ctx, addrs[0], deposit.Amount)
	suite.Require().NoError(err)
}
