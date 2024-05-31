package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/stretchr/testify/suite"
)

type burnIntegrationTestSuite struct {
	testutil.Suite
}

func (suite *burnIntegrationTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestBurnIntegrationTest(t *testing.T) {
	suite.Run(t, new(burnIntegrationTestSuite))
}

func (suite *burnIntegrationTestSuite) TestBurnCoins_MatchingErrors() {
	// x/precisebank BurnCoins should be identical to x/bank BurnCoins to
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
			"module account fee_collector does not have permissions to burn tokens: unauthorized",
		},
		{
			"invalid amount",
			// Has burn permissions so it goes to the amt check
			stakingtypes.BondedPoolName,
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
				// Check x/bank BurnCoins for identical error
				bankErr := suite.BankKeeper.BurnCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				suite.Require().Error(bankErr)
				suite.Require().EqualError(bankErr, tt.wantErr, "expected error should match x/bank BurnCoins error")

				pbankErr := suite.Keeper.BurnCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				suite.Require().Error(pbankErr)
				// Compare strings instead of errors, as error stack is still different
				suite.Require().Equal(
					bankErr.Error(),
					pbankErr.Error(),
					"x/precisebank error should match x/bank BurnCoins error",
				)
			}

			if tt.wantPanic != "" {
				// First check the wantPanic string is correct.
				// Actually specify the panic string in the test since it makes
				// it more clear we are testing specific and different cases.
				suite.Require().PanicsWithError(tt.wantPanic, func() {
					_ = suite.BankKeeper.BurnCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				}, "expected panic error should match x/bank BurnCoins")

				suite.Require().PanicsWithError(tt.wantPanic, func() {
					_ = suite.Keeper.BurnCoins(suite.Ctx, tt.recipientModule, tt.mintAmount)
				}, "x/precisebank panic should match x/bank BurnCoins")
			}
		})
	}
}
