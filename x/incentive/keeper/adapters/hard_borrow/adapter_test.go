package hard_borrow_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/hard_borrow"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/stretchr/testify/suite"
)

type HardBorrowAdapterTestSuite struct {
	suite.Suite

	app app.TestApp
	ctx sdk.Context

	denomA string

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

func TestHardAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(HardBorrowAdapterTestSuite))
}

func (suite *HardBorrowAdapterTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
	suite.app = app.NewTestApp()

	suite.ctx = suite.app.NewContext(true, tmprototypes.Header{Time: suite.genesisTime})

	suite.denomA = "usdx"

	err := suite.app.FundAccount(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(1000000000000)),
		),
	)
	suite.NoError(err)

	err = suite.app.FundAccount(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(1000000000000)),
		),
	)
	suite.NoError(err)

	hardGs := hardtypes.NewGenesisState(
		hardtypes.NewParams(
			hardtypes.MoneyMarkets{
				hardtypes.NewMoneyMarket(
					suite.denomA,
					hardtypes.NewBorrowLimit(
						true,
						sdk.NewDec(500000000000),
						sdk.MustNewDecFromStr("0.5"),
					),
					"ukava:usd:30",
					sdk.NewInt(1000000),
					hardtypes.NewInterestRateModel(
						sdk.MustNewDecFromStr("0"),
						sdk.MustNewDecFromStr("0.05"),
						sdk.MustNewDecFromStr("0.8"),
						sdk.NewDec(5),
					),
					sdk.MustNewDecFromStr("0.025"),
					sdk.MustNewDecFromStr("0.02"),
				),
			},
			sdk.ZeroDec(),
		),
		hardtypes.DefaultAccumulationTimes,
		nil,
		nil,
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
	)

	pricefeedGs := pricefeedtypes.NewGenesisState(
		pricefeedtypes.NewParams(
			[]pricefeedtypes.Market{
				pricefeedtypes.NewMarket(
					"ukava:usd:30",
					"ukava",
					"usd",
					nil,
					true,
				),
			},
		),
		[]pricefeedtypes.PostedPrice{
			pricefeedtypes.NewPostedPrice(
				"ukava:usd:30",
				suite.addrs[0],
				sdk.MustNewDecFromStr("1.5"),
				suite.ctx.BlockTime().Add(time.Hour),
			),
		},
	)

	cdc := suite.app.AppCodec()
	suite.app.InitializeFromGenesisStates(
		app.GenesisState{
			hardtypes.ModuleName:      cdc.MustMarshalJSON(&hardGs),
			pricefeedtypes.ModuleName: cdc.MustMarshalJSON(&pricefeedGs),
		},
	)
}

func (suite *HardBorrowAdapterTestSuite) TestHardAdapter_OwnerSharesBySource() {
	hardKeeper := suite.app.GetHardKeeper()

	// Need some deposits in order to borrow
	err := hardKeeper.Deposit(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(100_000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Deposit(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(100_000)),
		),
	)
	suite.NoError(err)

	// ------------------------------------------
	// Actual borrows now
	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(40_000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(25_000)),
		),
	)
	suite.NoError(err)

	// ------------------------------------------
	// Accrue interest for hard borrows
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour))
	hard.BeginBlocker(suite.ctx, hardKeeper)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(24 * time.Hour))
	hard.BeginBlocker(suite.ctx, hardKeeper)

	// ------------------------------------------
	// Modify borrow to sync with new interest - Index field
	// Second borrow has no modifications
	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(1_000)),
		),
	)
	suite.NoError(err)

	interestFactor, found := hardKeeper.GetBorrowInterestFactor(suite.ctx, suite.denomA)
	suite.Require().True(found)
	suite.Require().True(interestFactor.GT(sdk.OneDec()), "interest factor should be greater than 1")

	adapter := hard_borrow.NewSourceAdapter(hardKeeper)

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
				suite.denomA,
			},
			map[string]sdk.Dec{
				// Normalized borrow amount - after interest
				// 41_000 - interest
				suite.denomA: sdk.MustNewDecFromStr("40999.189325524906210261"),
			},
		},
		{
			"depositor has shares - including empty deposits",
			suite.addrs[1],
			[]string{
				suite.denomA,
				"unknown",
			},
			map[string]sdk.Dec{
				suite.denomA: sdk.NewDecWithPrec(25_000, 0),
				"unknown":    sdk.ZeroDec(),
			},
		},
		{
			"non-depositor has zero shares",
			suite.addrs[2],
			[]string{
				suite.denomA,
			},
			map[string]sdk.Dec{
				suite.denomA: sdk.ZeroDec(),
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

func (suite *HardBorrowAdapterTestSuite) TestHardAdapter_TotalSharesBySource() {
	hardKeeper := suite.app.GetHardKeeper()

	err := hardKeeper.Deposit(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(100000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Deposit(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(100000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(40_000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(25_000)),
		),
	)
	suite.NoError(err)

	adapter := hard_borrow.NewSourceAdapter(suite.app.GetHardKeeper())

	tests := []struct {
		name         string
		giveSourceID string
		wantShares   sdk.Dec
	}{
		{
			"total shares",
			suite.denomA,
			sdk.NewDecWithPrec(40_000+25_000, 0),
		},
		{
			"empty or invalid denom empty",
			"denom2",
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

func (suite *HardBorrowAdapterTestSuite) TestHardAdapter_TotalSharesBySource_Interest() {
	hardKeeper := suite.app.GetHardKeeper()

	err := hardKeeper.Deposit(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(100000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Deposit(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(100000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(40_000)),
		),
	)
	suite.NoError(err)

	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(25_000)),
		),
	)
	suite.NoError(err)

	// ------------------------------------------
	// Accrue interest for hard borrows
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour))
	hard.BeginBlocker(suite.ctx, hardKeeper)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(24 * time.Hour))
	hard.BeginBlocker(suite.ctx, hardKeeper)

	// ------------------------------------------
	// Modify borrow to sync with new interest - Index field
	// Second borrow has no modifications
	err = hardKeeper.Borrow(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(suite.denomA, sdk.NewInt(1_000)),
		),
	)
	suite.NoError(err)

	adapter := hard_borrow.NewSourceAdapter(suite.app.GetHardKeeper())

	tests := []struct {
		name         string
		giveSourceID string
		wantShares   sdk.Dec
	}{
		{
			"total shares",
			suite.denomA,
			// 40_000 + 25_000 + 1_000
			// ~ 66_000 with some minor precision error after normalizing
			sdk.MustNewDecFromStr("65999.085238489179768533"),
		},
		{
			"empty or invalid denom empty",
			"denom2",
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
