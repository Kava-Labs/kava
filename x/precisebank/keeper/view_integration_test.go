package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/suite"
)

type viewIntegrationTestSuite struct {
	testutil.Suite
}

func (suite *viewIntegrationTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestViewIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(viewIntegrationTestSuite))
}

func (suite *viewIntegrationTestSuite) TestKeeper_SpendableCoin() {
	tests := []struct {
		name      string
		giveDenom string // queried denom for balance

		giveBankBal       sdk.Coins   // full balance
		giveFractionalBal sdkmath.Int // stored fractional balance for giveAddr
		giveLockedCoins   sdk.Coins   // locked coins

		wantSpendableBal sdk.Coin
	}{
		{
			"extended denom, no fractional - locked coins",
			types.ExtendedCoinDenom,
			// queried bank balance in ukava when querying for akava
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			sdkmath.ZeroInt(),
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(10))),
			// (integer + fractional) - locked
			sdk.NewCoin(
				types.ExtendedCoinDenom,
				types.ConversionFactor().MulRaw(1000-10),
			),
		},
		{
			"extended denom, with fractional - locked coins",
			types.ExtendedCoinDenom,
			// queried bank balance in ukava when querying for akava
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			sdkmath.NewInt(5000),
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(10))),
			sdk.NewCoin(
				types.ExtendedCoinDenom,
				// (integer - locked) + fractional
				types.ConversionFactor().MulRaw(1000-10).AddRaw(5000),
			),
		},
		{
			"non-extended denom - ukava returns ukava",
			types.IntegerCoinDenom,
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			sdkmath.ZeroInt(),
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(10))),
			sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(990)),
		},
		{
			"non-extended denom, with fractional - ukava returns ukava",
			types.IntegerCoinDenom,
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			// does not affect balance
			sdkmath.NewInt(100),
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(10))),
			sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(990)),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			addr := sdk.AccAddress([]byte("test-address"))

			suite.MintToAccount(addr, tt.giveBankBal)

			// Set fractional balance in store before query
			suite.Keeper.SetFractionalBalance(suite.Ctx, addr, tt.giveFractionalBal)

			// Add some locked coins
			acc := suite.AccountKeeper.GetAccount(suite.Ctx, addr)
			if acc == nil {
				acc = authtypes.NewBaseAccount(addr, nil, 0, 0)
			}

			vestingAcc := vestingtypes.NewPeriodicVestingAccount(
				acc.(*authtypes.BaseAccount),
				tt.giveLockedCoins,
				suite.Ctx.BlockTime().Unix(),
				vestingtypes.Periods{
					vestingtypes.Period{
						Length: 100,
						Amount: tt.giveLockedCoins,
					},
				},
			)
			suite.AccountKeeper.SetAccount(suite.Ctx, vestingAcc)

			fetchedLockedCoins := vestingAcc.LockedCoins(suite.Ctx.BlockTime())
			suite.Require().Equal(
				tt.giveLockedCoins,
				fetchedLockedCoins,
				"locked coins should be matching at current block time",
			)

			spendableCoinsWithLocked := suite.Keeper.SpendableCoin(suite.Ctx, addr, tt.giveDenom)

			suite.Require().Equalf(
				tt.wantSpendableBal,
				spendableCoinsWithLocked,
				"expected spendable coins of denom %s",
				tt.giveDenom,
			)
		})
	}
}

func (suite *viewIntegrationTestSuite) TestKeeper_HiddenReserve() {
	// Reserve balances should not be shown to consumers of x/precisebank, as it
	// represents the fractional balances of accounts.

	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	addr1 := sdk.AccAddress{1}

	// Make the reserve hold a non-zero balance
	// Mint fractional coins to an account, which should cause a mint of 1
	// integer coin to the reserve to back it.
	unrelatedCoin := sdk.NewCoin("unrelated", sdk.NewInt(1000))
	suite.MintToAccount(
		addr1,
		sdk.NewCoins(
			sdk.NewCoin(types.ExtendedCoinDenom, sdk.NewInt(1000)),
			unrelatedCoin,
		),
	)

	// Check underlying x/bank balance for reserve
	reserveIntCoin := suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, types.IntegerCoinDenom)
	suite.Require().Equal(sdkmath.NewInt(1), reserveIntCoin.Amount, "reserve should hold 1 integer coin")

	// x/precisebank queries for reserve show as 0
	denom := types.ExtendedCoinDenom

	suite.Run("GetBalance()", func() {
		coin := suite.Keeper.GetBalance(suite.Ctx, moduleAddr, denom)
		suite.Require().Equal(denom, coin.Denom)
		suite.Require().Equal(sdkmath.ZeroInt(), coin.Amount)
	})

	suite.Run("SpendableCoin()", func() {
		spendableCoin := suite.Keeper.SpendableCoin(suite.Ctx, moduleAddr, denom)
		suite.Require().Equal(denom, spendableCoin.Denom)
		suite.Require().Equal(sdkmath.ZeroInt(), spendableCoin.Amount)
	})

	suite.Run("GetBalance() unrelated denom", func() {
		// Not affecting module account
		moduleCoin := suite.Keeper.GetBalance(suite.Ctx, moduleAddr, "unrelated")
		suite.Require().Equal(unrelatedCoin.Denom, moduleCoin.Denom)
		suite.Require().Equal(sdkmath.ZeroInt(), moduleCoin.Amount)

		// Still visible in user account balance
		accCoin := suite.Keeper.GetBalance(suite.Ctx, addr1, "unrelated")
		suite.Require().Equal(unrelatedCoin, accCoin)
	})

	suite.Run("SpendableCoin() unrelated denom", func() {
		// Not affecting module account
		moduleCoin := suite.Keeper.SpendableCoin(suite.Ctx, moduleAddr, "unrelated")
		suite.Require().Equal(unrelatedCoin.Denom, moduleCoin.Denom)
		suite.Require().Equal(sdkmath.ZeroInt(), moduleCoin.Amount)

		// Still visible in user account balance
		accCoin := suite.Keeper.SpendableCoin(suite.Ctx, addr1, "unrelated")
		suite.Require().Equal(unrelatedCoin, accCoin)
	})
}
