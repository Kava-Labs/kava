package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/suite"
)

type mintIntegrationTestSuite struct {
	testutil.Suite
}

func (suite *mintIntegrationTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestMintIntegrationTest(t *testing.T) {
	suite.Run(t, new(mintIntegrationTestSuite))
}

func (suite *mintIntegrationTestSuite) TestBlockedRecipient() {
	// Tests that sending funds to x/precisebank is disallowed.
	// x/precisebank balance is used as the reserve funds and should not be
	// directly interacted with by external modules or users.
	msgServer := bankkeeper.NewMsgServerImpl(suite.BankKeeper)

	fromAddr := sdk.AccAddress{1}

	// To x/precisebank
	toAddr := suite.AccountKeeper.GetModuleAddress(types.ModuleName)
	amount := cs(c("ukava", 1000))

	msg := banktypes.NewMsgSend(fromAddr, toAddr, amount)

	_, err := msgServer.Send(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().Error(err)

	suite.Require().EqualError(
		err,
		fmt.Sprintf("%s is not allowed to receive funds: unauthorized", toAddr.String()),
	)
}

func (suite *mintIntegrationTestSuite) TestMintCoins_MatchingErrors() {
	// x/precisebank MintCoins should be identical to x/bank MintCoins to
	// consumers. This test ensures that the panics & errors returned by
	// x/precisebank are identical to x/bank.

	tests := []struct {
		name            string
		recipientModule string
		mintAmount      sdk.Coins
		wantErr         string
		wantPanic       string
	}{
		{
			"invalid module",
			"notamodule",
			cs(c("ukava", 1000)),
			"",
			"module account notamodule does not exist: unknown address",
		},
		{
			"no mint permissions",
			// Check app.go to ensure this module has no mint permissions
			authtypes.FeeCollectorName,
			cs(c("ukava", 1000)),
			"",
			"module account fee_collector does not have permissions to mint tokens: unauthorized",
		},
		{
			"invalid amount",
			minttypes.ModuleName,
			sdk.Coins{sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(-100)}},
			"-100ukava: invalid coins",
			"",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset
			suite.SetupTest()

			if tt.wantErr == "" && tt.wantPanic == "" {
				suite.Fail("test must specify either wantErr or wantPanic")
			}

			if tt.wantErr != "" {
				// Check x/bank MintCoins for identical error
				bankErr := suite.BankKeeper.MintCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				suite.Require().Error(bankErr)
				suite.Require().EqualError(bankErr, tt.wantErr, "expected error should match x/bank MintCoins error")

				pbankErr := suite.Keeper.MintCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				suite.Require().Error(pbankErr)
				// Compare strings instead of errors, as error stack is still different
				suite.Require().Equal(
					bankErr.Error(),
					pbankErr.Error(),
					"x/precisebank error should match x/bank MintCoins error",
				)
			}

			if tt.wantPanic != "" {
				// First check the wantPanic string is correct.
				// Actually specify the panic string in the test since it makes
				// it more clear we are testing specific and different cases.
				suite.Require().PanicsWithError(tt.wantPanic, func() {
					_ = suite.BankKeeper.MintCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				}, "expected panic error should match x/bank MintCoins")

				suite.Require().PanicsWithError(tt.wantPanic, func() {
					_ = suite.Keeper.MintCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				}, "x/precisebank panic should match x/bank MintCoins")
			}
		})
	}
}

func (suite *mintIntegrationTestSuite) TestMintCoins() {
	type mintTest struct {
		mintAmount sdk.Coins
		// Expected **full** balances after MintCoins(mintAmount)
		wantBalance sdk.Coins
	}

	tests := []struct {
		name            string
		recipientModule string
		// Instead of having a start balance, we just have a list of mints to
		// both test & get into desired non-default states.
		mints []mintTest
	}{
		{
			"passthrough - unrelated",
			minttypes.ModuleName,
			[]mintTest{
				{
					mintAmount:  cs(c("busd", 1000)),
					wantBalance: cs(c("busd", 1000)),
				},
			},
		},
		{
			"passthrough - integer denom",
			minttypes.ModuleName,
			[]mintTest{
				{
					mintAmount:  cs(c(types.IntegerCoinDenom, 1000)),
					wantBalance: cs(c(types.ExtendedCoinDenom, 1000000000000000)),
				},
			},
		},
		{
			"fractional only",
			minttypes.ModuleName,
			[]mintTest{
				{
					mintAmount:  cs(c(types.ExtendedCoinDenom, 1000)),
					wantBalance: cs(c(types.ExtendedCoinDenom, 1000)),
				},
				{
					mintAmount:  cs(c(types.ExtendedCoinDenom, 1000)),
					wantBalance: cs(c(types.ExtendedCoinDenom, 2000)),
				},
			},
		},
		{
			"fractional only with carry",
			minttypes.ModuleName,
			[]mintTest{
				{
					// Start with (1/4 * 3) = 0.75
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().QuoRaw(4).MulRaw(3))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().QuoRaw(4).MulRaw(3))),
				},
				{
					// Add another 0.50 to incur carry to test reserve on carry
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().QuoRaw(2))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().QuoRaw(4).MulRaw(5))),
				},
			},
		},
		{
			"fractional only, resulting in exact carry and 0 remainder",
			minttypes.ModuleName,
			[]mintTest{
				// mint 0.5, acc = 0.5, reserve = 1
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().QuoRaw(2))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().QuoRaw(2))),
				},
				// mint another 0.5, acc = 1, reserve = 0
				// Reserve actually goes down by 1 for integer carry
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().QuoRaw(2))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor())),
				},
			},
		},
		{
			"exact carry",
			minttypes.ModuleName,
			[]mintTest{
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor())),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor())),
				},
				// Carry again - exact amount
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor())),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(2))),
				},
			},
		},
		{
			"carry with extra",
			minttypes.ModuleName,
			[]mintTest{
				// MintCoins(C + 100)
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().AddRaw(100))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().AddRaw(100))),
				},
				// MintCoins(C + 5), total = 2C + 105
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().AddRaw(5))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(2).AddRaw(105))),
				},
			},
		},
		{
			"integer with fractional",
			minttypes.ModuleName,
			[]mintTest{
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(5).AddRaw(100))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(5).AddRaw(100))),
				},
				{
					mintAmount:  cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(2).AddRaw(5))),
					wantBalance: cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(7).AddRaw(105))),
				},
			},
		},
		{
			"with passthrough",
			minttypes.ModuleName,
			[]mintTest{
				{
					mintAmount: cs(
						ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(5).AddRaw(100)),
						c("busd", 1000),
					),
					wantBalance: cs(
						ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(5).AddRaw(100)),
						c("busd", 1000),
					),
				},
				{
					mintAmount: cs(
						ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(2).AddRaw(5)),
						c("meow", 40),
					),
					wantBalance: cs(
						ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(7).AddRaw(105)),
						c("busd", 1000),
						c("meow", 40),
					),
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset
			suite.SetupTest()

			recipientAddr := suite.AccountKeeper.GetModuleAddress(tt.recipientModule)

			for _, mt := range tt.mints {
				err := suite.Keeper.MintCoins(suite.Ctx, tt.recipientModule, mt.mintAmount)
				suite.Require().NoError(err)

				// -------------------------------------------------------------
				// Check FULL balances
				// x/bank balances + x/precisebank balance
				// Exclude "ukava" as x/precisebank balance will include it
				bankCoins := suite.BankKeeper.GetAllBalances(suite.Ctx, recipientAddr)

				// Only use x/bank balances for non-ukava denoms
				var denoms []string
				for _, coin := range bankCoins {
					// Ignore integer coins, query the extended denom instead
					if coin.Denom == types.IntegerCoinDenom {
						continue
					}

					denoms = append(denoms, coin.Denom)
				}

				// Add the extended denom to the list of denoms to balance check
				// Will be included in balance check even if x/bank doesn't have
				// ukava.
				denoms = append(denoms, types.ExtendedCoinDenom)

				// All balance queries through x/precisebank
				afterBalance := sdk.NewCoins()
				for _, denom := range denoms {
					coin := suite.Keeper.GetBalance(suite.Ctx, recipientAddr, denom)
					afterBalance = afterBalance.Add(coin)
				}

				suite.Require().Equal(
					mt.wantBalance.String(),
					afterBalance.String(),
					"unexpected balance after minting %s to %s",
				)

				// Ensure reserve is backing all minted fractions
				allInvariantsFn := keeper.AllInvariants(suite.Keeper)
				msg, stop := allInvariantsFn(suite.Ctx)
				suite.Require().Falsef(stop, "invariant should not be broken: %s", msg)
			}
		})
	}
}

func FuzzMintCoins(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(100))
	f.Add(types.ConversionFactor().Int64())
	f.Add(types.ConversionFactor().QuoRaw(2).Int64())
	f.Add(types.ConversionFactor().MulRaw(5).Int64())
	f.Add(types.ConversionFactor().MulRaw(2).AddRaw(123948723).Int64())

	f.Fuzz(func(t *testing.T, amount int64) {
		// No negative amounts
		if amount < 0 {
			amount = -amount
		}

		// Manually setup test suite since no direct Fuzz support in test suites
		suite := new(mintIntegrationTestSuite)
		suite.SetT(t)
		suite.SetS(suite)
		suite.SetupTest()

		mintCount := int64(10)

		suite.T().Logf("minting %d %d times", amount, mintCount)

		// Mint 10 times to include mints from non-zero balances
		for i := int64(0); i < mintCount; i++ {
			err := suite.Keeper.MintCoins(
				suite.Ctx,
				minttypes.ModuleName,
				cs(c(types.ExtendedCoinDenom, amount)),
			)
			suite.Require().NoError(err)
		}

		// Check FULL balances
		recipientAddr := suite.AccountKeeper.GetModuleAddress(minttypes.ModuleName)
		bal := suite.Keeper.GetBalance(suite.Ctx, recipientAddr, types.ExtendedCoinDenom)

		suite.Require().Equalf(
			amount*mintCount,
			bal.Amount.Int64(),
			"unexpected balance after minting %d %d times",
			amount,
			mintCount,
		)

		// Run Invariants to ensure remainder is backing all minted fractions
		// and in a valid state
		res, stop := keeper.AllInvariants(suite.Keeper)(suite.Ctx)
		suite.False(stop, "invariant should not be broken")
		suite.Empty(res, "unexpected invariant message")
	})
}
