package keeper_test

import (
	"fmt"
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

func (suite *sendIntegrationTestSuite) TestSendCoinsFromAccountToModule_MatchingErrors() {
	// No specific errors for SendCoinsFromAccountToModule, only 1 panic if
	// the module account does not exist

	tests := []struct {
		name            string
		sender          sdk.AccAddress
		recipientModule string
		sendAmount      sdk.Coins
		wantPanic       string
	}{
		// SendCoinsFromAccountToModule specific errors/panics
		{
			"missing module account - passthrough",
			sdk.AccAddress([]byte{2}),
			"cat",
			cs(c("usdc", 1000)),
			"module account cat does not exist: unknown address",
		},
		{
			"missing module account - extended",
			sdk.AccAddress([]byte{2}),
			"cat",
			cs(c(types.ExtendedCoinDenom, 1000)),
			"module account cat does not exist: unknown address",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset
			suite.SetupTest()

			suite.Require().NotEmpty(tt.wantPanic, "test case must have a wantPanic")

			suite.Require().PanicsWithError(tt.wantPanic, func() {
				suite.BankKeeper.SendCoinsFromAccountToModule(suite.Ctx, tt.sender, tt.recipientModule, tt.sendAmount)
			}, "wantPanic should match x/bank SendCoinsFromAccountToModule panic")

			suite.Require().PanicsWithError(tt.wantPanic, func() {
				suite.Keeper.SendCoinsFromAccountToModule(suite.Ctx, tt.sender, tt.recipientModule, tt.sendAmount)
			}, "x/precisebank panic should match x/bank SendCoinsFromAccountToModule panic")
		})
	}
}

func (suite *sendIntegrationTestSuite) TestSendCoinsFromModuleToAccount_MatchingErrors() {
	// Ensure errors match x/bank errors AND panics. This needs to be well
	// tested before SendCoins as all send tests rely on this to initialize
	// account balances.
	// No unit test with mock x/bank for SendCoinsFromModuleToAccount since
	// we only are testing the errors/panics specific to the method and
	// remaining logic is the same as SendCoins.

	blockedMacAddrs := suite.App.GetBlockedMaccAddrs()

	var blockedAddr sdk.AccAddress
	// Get the first blocked address
	for addr, isBlocked := range blockedMacAddrs {
		if isBlocked {
			blockedAddr = sdk.MustAccAddressFromBech32(addr)
			break
		}
	}

	suite.Require().NotEmpty(blockedAddr, "no blocked addresses found")

	tests := []struct {
		name         string
		senderModule string
		recipient    sdk.AccAddress
		sendAmount   sdk.Coins
		wantErr      string
		wantPanic    string
	}{
		// SendCoinsFromModuleToAccount specific errors/panics
		{
			"missing module account - passthrough",
			"cat",
			sdk.AccAddress([]byte{2}),
			cs(c("usdc", 1000)),
			"",
			"module account cat does not exist: unknown address",
		},
		{
			"missing module account - extended",
			"cat",
			sdk.AccAddress([]byte{2}),
			cs(c(types.ExtendedCoinDenom, 1000)),
			"",
			"module account cat does not exist: unknown address",
		},
		{
			"blocked recipient address - passthrough",
			types.ModuleName,
			blockedAddr,
			cs(c("usdc", 1000)),
			fmt.Sprintf("%s is not allowed to receive funds: unauthorized", blockedAddr.String()),
			"",
		},
		{
			"blocked recipient address - extended",
			types.ModuleName,
			blockedAddr,
			cs(c(types.ExtendedCoinDenom, 1000)),
			fmt.Sprintf("%s is not allowed to receive funds: unauthorized", blockedAddr.String()),
			"",
		},
		// SendCoins specific errors/panics
		{
			"invalid coins",
			types.ModuleName,
			sdk.AccAddress([]byte{2}),
			sdk.Coins{sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)}},
			"-1ukava: invalid coins",
			"",
		},
		{
			"insufficient balance - passthrough",
			types.ModuleName,
			sdk.AccAddress([]byte{2}),
			cs(c(types.IntegerCoinDenom, 1000)),
			"spendable balance  is smaller than 1000ukava: insufficient funds",
			"",
		},
		{
			"insufficient balance - extended",
			types.ModuleName,
			sdk.AccAddress([]byte{2}),
			// We can still test insufficient bal errors with "akava" since
			// we also expect it to not exist in x/bank
			cs(c(types.ExtendedCoinDenom, 1000)),
			"spendable balance  is smaller than 1000akava: insufficient funds",
			"",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset
			suite.SetupTest()

			if tt.wantPanic == "" && tt.wantErr == "" {
				suite.FailNow("test case must have a wantErr or wantPanic")
			}

			if tt.wantPanic != "" {
				suite.Require().Empty(tt.wantErr, "test case must not have a wantErr if wantPanic is set")

				suite.Require().PanicsWithError(tt.wantPanic, func() {
					suite.BankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, tt.senderModule, tt.recipient, tt.sendAmount)
				}, "wantPanic should match x/bank SendCoinsFromModuleToAccount panic")

				suite.Require().PanicsWithError(tt.wantPanic, func() {
					suite.Keeper.SendCoinsFromModuleToAccount(suite.Ctx, tt.senderModule, tt.recipient, tt.sendAmount)
				}, "x/precisebank panic should match x/bank SendCoinsFromModuleToAccount panic")
			}

			if tt.wantErr != "" {
				bankErr := suite.BankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, tt.senderModule, tt.recipient, tt.sendAmount)
				suite.Require().Error(bankErr)
				suite.Require().EqualError(bankErr, tt.wantErr, "expected error should match x/bank SendCoins error")

				pbankErr := suite.Keeper.SendCoinsFromModuleToAccount(suite.Ctx, tt.senderModule, tt.recipient, tt.sendAmount)
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
			"-1ukava: invalid coins",
		},
		{
			"insufficient balance - passthrough",
			cs(c(types.IntegerCoinDenom, 1000)),
			"spendable balance  is smaller than 1000ukava: insufficient funds",
		},
		{
			"insufficient balance - extended",
			// We can still test insufficient bal errors with "akava" since
			// we also expect it to not exist in x/bank
			cs(c(types.ExtendedCoinDenom, 1000)),
			"spendable balance  is smaller than 1000akava: insufficient funds",
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
