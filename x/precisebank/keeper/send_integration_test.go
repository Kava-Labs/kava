package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/keeper"
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
		name          string
		initialAmount sdk.Coins
		sendAmount    sdk.Coins
		wantErr       string
	}{
		{
			"invalid coins",
			cs(),
			sdk.Coins{sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)}},
			"-1ukava: invalid coins",
		},
		{
			"insufficient empty balance - passthrough",
			cs(),
			cs(c(types.IntegerCoinDenom, 1000)),
			"spendable balance  is smaller than 1000ukava: insufficient funds",
		},
		{
			"insufficient empty balance - extended",
			cs(),
			// We can still test insufficient bal errors with "akava" since
			// we also expect it to not exist in x/bank
			cs(c(types.ExtendedCoinDenom, 1000)),
			"spendable balance  is smaller than 1000akava: insufficient funds",
		},
		{
			"insufficient non-empty balance - passthrough",
			cs(c(types.IntegerCoinDenom, 100), c("usdc", 1000)),
			cs(c(types.IntegerCoinDenom, 1000)),
			"spendable balance 100ukava is smaller than 1000ukava: insufficient funds",
		},
		// non-empty akava transfer error is tested in SendCoins, not here since
		// x/bank doesn't hold akava
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset
			suite.SetupTest()
			sender := sdk.AccAddress([]byte{1})
			recipient := sdk.AccAddress([]byte{2})

			suite.Require().NotEmpty(tt.wantErr, "test case must have a wantErr")

			suite.MintToAccount(sender, tt.initialAmount)

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
	// SendCoins is tested mostly in this integration test, as a unit test with
	// mocked BankKeeper overcomplicates expected keepers and makes initializing
	// balances very complex.

	tests := []struct {
		name                  string
		giveStartBalSender    sdk.Coins
		giveStartBalRecipient sdk.Coins
		giveAmt               sdk.Coins
		wantErr               string
	}{
		{
			"insufficient balance error denom matches",
			cs(c(types.ExtendedCoinDenom, 10), c("usdc", 1000)),
			cs(),
			cs(c(types.ExtendedCoinDenom, 1000)),
			"spendable balance 10akava is smaller than 1000akava: insufficient funds",
		},
		{
			"passthrough - unrelated",
			cs(c("cats", 1000)),
			cs(),
			cs(c("cats", 1000)),
			"",
		},
		{
			"passthrough - integer denom",
			cs(c(types.IntegerCoinDenom, 1000)),
			cs(),
			cs(c(types.IntegerCoinDenom, 1000)),
			"",
		},
		{
			"akava send - 1akava to 0 balance",
			// Starting balances
			cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(5))),
			cs(),
			// Send amount
			cs(c(types.ExtendedCoinDenom, 1)), // akava
			"",
		},
		{
			"sender borrow from integer",
			// 1ukava, 0 fractional
			cs(ci(types.ExtendedCoinDenom, types.ConversionFactor())),
			cs(),
			// Send 1 with 0 fractional
			cs(c(types.ExtendedCoinDenom, 1)),
			"",
		},
		{
			"receiver carry",
			cs(c(types.ExtendedCoinDenom, 1000)),
			// max fractional amount, carries over to integer
			cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().SubRaw(1))),
			cs(c(types.ExtendedCoinDenom, 1)),
			"",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			sender := sdk.AccAddress([]byte{1})
			recipient := sdk.AccAddress([]byte{2})

			// Initialize balances
			suite.MintToAccount(sender, tt.giveStartBalSender)
			suite.MintToAccount(recipient, tt.giveStartBalRecipient)

			senderBalBefore := suite.GetAllBalances(sender)
			recipientBalBefore := suite.GetAllBalances(recipient)

			err := suite.Keeper.SendCoins(suite.Ctx, sender, recipient, tt.giveAmt)
			if tt.wantErr != "" {
				suite.Require().Error(err)
				suite.Require().EqualError(err, tt.wantErr)
				return
			}

			suite.Require().NoError(err)

			// Check balances
			senderBalAfter := suite.GetAllBalances(sender)
			recipientBalAfter := suite.GetAllBalances(recipient)

			// Convert send amount coins to extended coins. i.e. if send coins
			// includes ukava, convert it so that its the equivalent akava
			// amount so its easier to compare. Compare extended coins only.
			sendAmountExtended := tt.giveAmt
			sendAmountInteger := tt.giveAmt.AmountOf(types.IntegerCoinDenom)
			if !sendAmountInteger.IsZero() {
				integerCoin := sdk.NewCoin(types.IntegerCoinDenom, sendAmountInteger)
				sendAmountExtended = sendAmountExtended.Sub(integerCoin)

				// Add equivalent extended coin
				extendedCoinAmount := sendAmountInteger.Mul(types.ConversionFactor())
				extendedCoin := sdk.NewCoin(types.ExtendedCoinDenom, extendedCoinAmount)
				sendAmountExtended = sendAmountExtended.Add(extendedCoin)
			}

			suite.Require().Equal(
				senderBalBefore.Sub(sendAmountExtended...),
				senderBalAfter,
			)

			suite.Require().Equal(
				recipientBalBefore.Add(sendAmountExtended...),
				recipientBalAfter,
			)

			invariantFn := keeper.AllInvariants(suite.Keeper)
			res, stop := invariantFn(suite.Ctx)
			suite.Require().False(stop, "invariants should not stop")
			suite.Require().Empty(res, "invariants should not return any messages")
		})
	}
}

func FuzzSendCoins(f *testing.F) {
	f.Add(int64(100), int64(0), int64(2))
	f.Add(int64(100), int64(100), int64(5))
	f.Add(types.ConversionFactor().Int64(), int64(0), int64(500))
	f.Add(
		types.ConversionFactor().MulRaw(2).AddRaw(123948723).Int64(),
		types.ConversionFactor().MulRaw(2).Int64(),
		types.ConversionFactor().Int64(),
	)

	f.Fuzz(func(
		t *testing.T,
		startBalSender int64,
		startBalReceiver int64,
		sendAmount int64,
	) {
		// No negative amounts
		if startBalSender < 0 {
			startBalSender = -startBalSender
		}

		if startBalReceiver < 0 {
			startBalReceiver = -startBalReceiver
		}

		if sendAmount < 0 {
			sendAmount = -sendAmount
		}

		// Manually setup test suite since no direct Fuzz support in test suites
		suite := new(sendIntegrationTestSuite)
		suite.SetT(t)
		suite.SetS(suite)
		suite.SetupTest()

		sender := sdk.AccAddress([]byte{1})
		recipient := sdk.AccAddress([]byte{2})

		// Initial balances
		suite.MintToAccount(sender, cs(c(types.ExtendedCoinDenom, startBalSender)))
		suite.MintToAccount(recipient, cs(c(types.ExtendedCoinDenom, startBalReceiver)))

		// Send amount
		sendCoins := cs(c(types.ExtendedCoinDenom, sendAmount))
		err := suite.Keeper.SendCoins(suite.Ctx, sender, recipient, sendCoins)
		if startBalSender < sendAmount {
			suite.Require().Error(err, "expected insufficient funds error")
			return
		}

		suite.Require().NoError(err)

		// Check FULL balances
		balSender := suite.GetAllBalances(sender)
		balReceiver := suite.GetAllBalances(recipient)

		suite.Require().Equal(
			startBalSender-sendAmount,
			balSender.AmountOf(types.ExtendedCoinDenom).Int64(),
		)
		suite.Require().Equal(
			startBalReceiver+sendAmount,
			balReceiver.AmountOf(types.ExtendedCoinDenom).Int64(),
		)

		// Run Invariants to ensure remainder is backing all minted fractions
		// and in a valid state
		allInvariantsFn := keeper.AllInvariants(suite.Keeper)
		res, stop := allInvariantsFn(suite.Ctx)
		suite.Require().False(stop, "invariant should not be broken")
		suite.Require().Empty(res, "unexpected invariant message: %s", res)
	})
}
