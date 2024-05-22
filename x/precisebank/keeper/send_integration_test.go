package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/suite"
)

type sendIntegrationTestSuite struct {
	testutil.Suite
}

func (suite *sendIntegrationTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestSendIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(sendIntegrationTestSuite))
}

func (suite *sendIntegrationTestSuite) TestSendCoinsFromModuleToAccount_MatchingErrors() {
	// Ensure errors match x/bank errors AND panics. This needs to be well
	// tested before SendCoins as all send tests rely on this to initialize
	// account balances.

	tests := []struct {
		name         string
		senderModule string
		sendAmount   sdk.Coins
		wantErr      string
		wantPanic    string
	}{
		// SendCoinsFromModuleToAccount specific errors/panics
		{
			"missing module account - passthrough",
			"cat",
			cs(c("usdc", 1000)),
			"",
			"module account %s does not exist",
		},
		{
			"missing module account - extended",
			"cat",
			cs(c(types.ExtendedCoinDenom, 1000)),
			"",
			"module account %s does not exist",
		},
		{
			"blocked address - passthrough",
			types.ModuleName,
			cs(c("usdc", 1000)),
			"",
			"is not allowed to receive funds",
		},
		{
			"blocked address - extended",
			types.ModuleName,
			cs(c(types.ExtendedCoinDenom, 1000)),
			"",
			"is not allowed to receive funds",
		},
		// SendCoins specific errors/panics
		{
			"invalid coins",
			types.ModuleName,
			sdk.Coins{sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)}},
			"invalid coins",
			"",
		},
		{
			"insufficient balance - passthrough",
			types.ModuleName,
			cs(c(types.IntegerCoinDenom, 1000)),
			"insufficient account funds",
			"",
		},
		{
			"insufficient balance - extended",
			types.ModuleName,
			// We can still test insufficient bal errors with "akava" since
			// we also expect it to not exist in x/bank
			cs(c(types.ExtendedCoinDenom, 1000)),
			"insufficient account funds",
			"",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset
			suite.SetupTest()
			sender := sdk.AccAddress([]byte{1})
			recipient := sdk.AccAddress([]byte{2})

			if tt.wantPanic == "" && tt.wantErr == "" {
				suite.FailNow("test case must have a wantErr or wantPanic")
			}

			if tt.wantPanic != "" {
				suite.Require().Empty(tt.wantErr, "test case must not have a wantErr if wantPanic is set")

				suite.Require().PanicsWithValue(tt.wantPanic, func() {
					suite.BankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, types.ModuleName, recipient, tt.sendAmount)
				}, "wantPanic should match x/bank SendCoinsFromModuleToAccount panic")

				suite.Require().PanicsWithValue(tt.wantPanic, func() {
					suite.Keeper.SendCoinsFromModuleToAccount(suite.Ctx, types.ModuleName, recipient, tt.sendAmount)
				}, "x/precisebank panic should match x/bank SendCoinsFromModuleToAccount panic")
			}

			if tt.wantErr != "" {
				bankErr := suite.BankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, types.ModuleName, recipient, tt.sendAmount)
				suite.Require().Error(bankErr)
				suite.Require().EqualError(bankErr, tt.wantErr, "expected error should match x/bank SendCoins error")

				pbankErr := suite.Keeper.SendCoins(suite.Ctx, sender, recipient, tt.sendAmount)
				suite.Require().Error(pbankErr)
				// Compare strings instead of errors, as error stack is still different
				suite.Require().Equal(
					bankErr.Error(),
					pbankErr.Error(),
					"x/precisebank error should match x/bank SendCoins error",
				)
			}
		})
	}
}

func (suite *sendIntegrationTestSuite) TestSendCoins_MatchingErrors() {
	// Ensure errors match x/bank errors

	tests := []struct {
		name       string
		sendAmount sdk.Coins
		wantErr    string
	}{
		{
			"invalid coins",
			sdk.Coins{sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)}},
			"invalid coins",
		},
		{
			"insufficient balance - passthrough",
			cs(c(types.IntegerCoinDenom, 1000)),
			"insufficient account funds",
		},
		{
			"insufficient balance - extended",
			// We can still test insufficient bal errors with "akava" since
			// we also expect it to not exist in x/bank
			cs(c(types.ExtendedCoinDenom, 1000)),
			"insufficient account funds",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset
			suite.SetupTest()
			sender := sdk.AccAddress([]byte{1})
			recipient := sdk.AccAddress([]byte{2})

			suite.Require().NotEmpty(tt.wantErr, "test case must have a wantErr")

			bankErr := suite.BankKeeper.SendCoins(suite.Ctx, sender, recipient, tt.sendAmount)
			suite.Require().Error(bankErr)
			suite.Require().EqualError(bankErr, tt.wantErr, "expected error should match x/bank SendCoins error")

			pbankErr := suite.Keeper.SendCoins(suite.Ctx, sender, recipient, tt.sendAmount)
			suite.Require().Error(pbankErr)
			// Compare strings instead of errors, as error stack is still different
			suite.Require().Equal(
				bankErr.Error(),
				pbankErr.Error(),
				"x/precisebank error should match x/bank SendCoins error",
			)
		})
	}
}

func (suite *sendIntegrationTestSuite) TestSendCoins() {
	tests := []struct {
		name                  string
		giveStartBalSender    sdk.Coins
		giveStartBalRecipient sdk.Coins
		giveAmt               sdk.Coins
		wantErr               string
	}{}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			sender := sdk.AccAddress([]byte{1})
			recipient := sdk.AccAddress([]byte{2})

			// Initialize balances
			suite.MintToAccount(sender, tt.giveStartBalSender)
			suite.MintToAccount(recipient, tt.giveStartBalRecipient)

			err := suite.Keeper.SendCoins(suite.Ctx, sender, recipient, tt.giveAmt)
			suite.Require().Error(err)

		})
	}
}
