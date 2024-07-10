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
	extCoin := sdk.NewCoin(types.ExtendedCoinDenom, types.ConversionFactor().AddRaw(1000))
	unrelatedCoin := sdk.NewCoin("unrelated", sdk.NewInt(1000))
	suite.MintToAccount(
		addr1,
		sdk.NewCoins(
			extCoin,
			unrelatedCoin,
		),
	)

	// Check underlying x/bank balance for reserve
	reserveIntCoin := suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, types.IntegerCoinDenom)
	suite.Require().Equal(
		sdkmath.NewInt(1),
		reserveIntCoin.Amount,
		"reserve should hold 1 integer coin",
	)

	tests := []struct {
		name       string
		giveAddr   sdk.AccAddress
		giveDenom  string
		wantAmount sdkmath.Int
	}{
		{
			"reserve account - hidden extended denom",
			moduleAddr,
			types.ExtendedCoinDenom,
			sdkmath.ZeroInt(),
		},
		{
			"reserve account - visible integer denom",
			moduleAddr,
			types.IntegerCoinDenom,
			sdkmath.OneInt(),
		},
		{
			"user account - visible extended denom",
			addr1,
			types.ExtendedCoinDenom,
			extCoin.Amount,
		},
		{
			"user account - visible integer denom",
			addr1,
			types.IntegerCoinDenom,
			extCoin.Amount.Quo(types.ConversionFactor()),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			coin := suite.Keeper.GetBalance(suite.Ctx, tt.giveAddr, tt.giveDenom)
			suite.Require().Equal(tt.wantAmount.Int64(), coin.Amount.Int64())

			spendableCoin := suite.Keeper.SpendableCoin(suite.Ctx, tt.giveAddr, tt.giveDenom)
			suite.Require().Equal(tt.wantAmount.Int64(), spendableCoin.Amount.Int64())
		})
	}
}
