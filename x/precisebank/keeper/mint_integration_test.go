package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

				// Check FULL balances
				bankCoins := suite.BankKeeper.GetAllBalances(suite.Ctx, recipientAddr)
				var denoms []string
				for _, coin := range bankCoins {
					// Ignore integer coins, query the extended denom instead
					if coin.Denom == types.IntegerCoinDenom {
						continue
					}

					denoms = append(denoms, coin.Denom)
				}

				// Add the extended denom to the list of denoms to balance check
				denoms = append(denoms, types.ExtendedCoinDenom)

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
				res, stop := allInvariantsFn(suite.Ctx)
				suite.Require().False(stop, "invariant broken: %s", res)
				suite.Require().Empty(res, "unexpected invariant message: %s", res)
			}
		})
	}
}

func FuzzMintCoins(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(100))
	f.Add(types.ConversionFactor().Int64())
	f.Add(types.ConversionFactor().MulRaw(5).Int64())
	f.Add(types.ConversionFactor().MulRaw(2).AddRaw(123948723).Int64())

	f.Fuzz(func(t *testing.T, amount int64) {
		// No negative amounts
		if amount < 0 {
			amount = -amount
		}

		suite := new(mintIntegrationTestSuite)
		suite.SetT(t)
		suite.SetS(suite)
		suite.SetupTest()

		// Mint 5 times to include mints from non-zero balances
		for i := 0; i < 5; i++ {
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
			amount*5,
			bal.Amount.Int64(),
			"unexpected balance after minting %d 5 times",
			amount,
		)

		// TODO: Check remainder

		allInvariantsFn := keeper.AllInvariants(suite.Keeper)
		res, stop := allInvariantsFn(suite.Ctx)
		suite.Require().False(stop, "invariant broken: %s", res)
		suite.Require().Empty(res, "unexpected invariant message: %s", res)
	})
}
