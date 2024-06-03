package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/suite"
)

type invariantsIntegrationTestSuite struct {
	testutil.Suite
}

func (suite *invariantsIntegrationTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestInvariantsIntegrationTest(t *testing.T) {
	suite.Run(t, new(invariantsIntegrationTestSuite))
}

func (suite *invariantsIntegrationTestSuite) FundReserve(amt sdkmath.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, amt))
	err := suite.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
}

func (suite *invariantsIntegrationTestSuite) TestReserveBackingFractionalInvariant() {
	tests := []struct {
		name       string
		setupFn    func(ctx sdk.Context, k keeper.Keeper)
		wantBroken bool
		wantMsg    string
	}{
		{
			"valid - empty state",
			func(_ sdk.Context, _ keeper.Keeper) {},
			false,
			"",
		},
		{
			"valid - fractional balances, no remainder",
			func(ctx sdk.Context, k keeper.Keeper) {
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, types.ConversionFactor().QuoRaw(2))
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, types.ConversionFactor().QuoRaw(2))
				// 1 integer backs same amount fractional
				suite.FundReserve(sdk.NewInt(1))
			},
			false,
			"",
		},
		{
			"valid - fractional balances, with remainder",
			func(ctx sdk.Context, k keeper.Keeper) {
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, types.ConversionFactor().QuoRaw(2))
				k.SetRemainderAmount(ctx, types.ConversionFactor().QuoRaw(2))
				// 1 integer backs same amount fractional including remainder
				suite.FundReserve(sdk.NewInt(1))
			},
			false,
			"",
		},
		{
			"invalid - insufficient reserve backing",
			func(ctx sdk.Context, k keeper.Keeper) {
				amt := types.ConversionFactor().QuoRaw(2)

				// 0.5 int coins x 4
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, amt)
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, amt)
				k.SetFractionalBalance(ctx, sdk.AccAddress{3}, amt)
				k.SetRemainderAmount(ctx, amt)

				// Needs 2 to back 0.5 x 4
				suite.FundReserve(sdk.NewInt(1))
			},
			true,
			"precisebank: reserve-backing-fractional invariant\nakava reserve balance 1000000000000 mismatches 2000000000000 (fractional balances 1500000000000 + remainder 500000000000)\n\n",
		},
		{
			"invalid - excess reserve backing",
			func(ctx sdk.Context, k keeper.Keeper) {
				amt := types.ConversionFactor().QuoRaw(2)

				// 0.5 int coins x 4
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, amt)
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, amt)
				k.SetFractionalBalance(ctx, sdk.AccAddress{3}, amt)
				k.SetRemainderAmount(ctx, amt)

				// Needs 2 to back 0.5 x 4
				suite.FundReserve(sdk.NewInt(3))
			},
			true,
			"precisebank: reserve-backing-fractional invariant\nakava reserve balance 3000000000000 mismatches 2000000000000 (fractional balances 1500000000000 + remainder 500000000000)\n\n",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset each time
			suite.SetupTest()

			tt.setupFn(suite.Ctx, suite.Keeper)

			invariantFn := keeper.ReserveBacksFractionsInvariant(suite.Keeper)
			msg, broken := invariantFn(suite.Ctx)

			if tt.wantBroken {
				suite.Require().True(broken, "invariant should be broken but is not")
				suite.Require().Equal(tt.wantMsg, msg)
			} else {
				suite.Require().Falsef(broken, "invariant should not be broken but is: %s", msg)
			}
		})
	}
}
