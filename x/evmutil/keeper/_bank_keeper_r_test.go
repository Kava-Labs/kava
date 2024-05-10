package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type evmBankKeeperTestSuite2 struct {
	testutil.Suite
}

func (suite *evmBankKeeperTestSuite2) SetupTest() {
	suite.Suite.SetupTest()
}

func TestEvmBankKeeperTestSuite2(t *testing.T) {
	suite.Run(t, new(evmBankKeeperTestSuite2))
}

func (suite *evmBankKeeperTestSuite2) TestGetBalance() {

}

func (suite *evmBankKeeperTestSuite2) TestSendCoins() {
	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	tests := []struct {
		name string
		// Balances are all in akava
		startSenderBalance    sdkmath.Int
		startRecipientBalance sdkmath.Int
		sendAmount            sdkmath.Int
	}{
		{
			"1 akava - no borrow",
			sdkmath.NewInt(1_000_000_000_000_000_001),
			sdkmath.NewInt(0),
			sdkmath.NewInt(1),
		},
		{
			"1 akava - sender borrows 1 ukava",
			sdkmath.NewInt(1_000_000_000_000_000_000),
			sdkmath.NewInt(0),
			sdkmath.NewInt(1),
		},
		{
			"1 ukava & 1 akava - sender borrows 1 ukava",
			sdkmath.NewInt(5_000_000_000_000_000_000),
			sdkmath.NewInt(0),
			sdkmath.NewInt(1_000_000_000_000_000_001),
		},
		{
			"1 akava - recipient carries 1ukava",
			sdkmath.NewInt(1_000_000_000_000_000_001),
			sdkmath.NewInt(999_999_999_999_999_999),
			sdkmath.NewInt(1),
		},
		{
			"1 ukava & 1 akava - recipient carries 1ukava",
			sdkmath.NewInt(5_000_000_000_000_000_001),
			sdkmath.NewInt(999_999_999_999_999_999),
			sdkmath.NewInt(1_000_000_000_000_000_001),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			// Avoid using MintCoins() to isolate test to only SendCoins behavior
			bk := keeper.NewEvmBankKeeper(
				suite.Keeper,
				suite.BankKeeper,
				suite.AccountKeeper,
			)

			// --- Setup test state ---

			// Uses x/bank MintCoins for regular ukava! Isolate test to only
			// SendCoins behavior, not custom MintCoins
			suite.MintAkavaToAccount(addr1, tt.startSenderBalance)
			suite.MintAkavaToAccount(addr2, tt.startRecipientBalance)

			// Double check starting balances are correct, as we are also using
			// the SendCoins() method to initialize account balances
			suite.Require().Equal(
				tt.startSenderBalance,
				bk.GetBalance(suite.Ctx, addr1, keeper.EvmDenom),
				"starting balance of sender should be correct",
			)
			suite.Require().Equal(
				tt.startRecipientBalance,
				bk.GetBalance(suite.Ctx, addr2, keeper.EvmDenom),
				"starting balance of recipient should be correct",
			)
			suite.VerifyReserveState()

			// --- Run test specific transfer ---
			sendCoins := sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, tt.sendAmount))
			err := bk.SendCoins(suite.Ctx, addr1, addr2, sendCoins)
			suite.Require().NoError(err)

			suite.VerifyReserveState()

			// --- Check final balances ---

			// Sender balance
			finalBal := bk.GetBalance(suite.Ctx, addr1, keeper.EvmDenom)
			suite.Require().Equal(
				tt.startSenderBalance.Sub(tt.sendAmount),
				finalBal.Amount,
				"final balance should be correct",
			)

			// Recipient balance
			recipientBal := bk.GetBalance(suite.Ctx, addr2, keeper.EvmDenom)
			suite.Require().Equal(
				tt.startRecipientBalance.Add(tt.sendAmount),
				recipientBal.Amount,
				"recipient balance should be correct",
			)
		})
	}
}

func (suite *evmBankKeeperTestSuite2) MintAkavaToAccount(
	recipient sdk.AccAddress,
	amount sdkmath.Int,
) {
	bk := keeper.NewEvmBankKeeper(
		suite.Keeper,
		suite.BankKeeper,
		suite.AccountKeeper,
	)

	ukavaCoins := sdk.NewCoins(sdk.NewCoin(
		keeper.CosmosDenom,
		// Add 1 to ensure we have enough to convert to akava, effectively
		// a round up to the nearest whole integer
		amount.Quo(keeper.ConversionMultiplier).AddRaw(1),
	))

	// Ensure x/bank exists
	_ = suite.AccountKeeper.GetModuleAccount(suite.Ctx, minttypes.ModuleName)

	err := suite.BankKeeper.MintCoins(
		suite.Ctx,
		minttypes.ModuleName,
		ukavaCoins,
	)
	suite.Require().NoError(err)

	// Initialize acc1 balance
	err = bk.SendCoinsFromModuleToAccount(
		suite.Ctx,
		minttypes.ModuleName,
		recipient,
		sdk.NewCoins(
			sdk.NewCoin(keeper.EvmDenom, amount),
		),
	)
	suite.Require().NoError(err)
}

func (suite *evmBankKeeperTestSuite2) VerifyReserveState() {
	// Returns full 18 decimal KAVA balance
	// reserveBal := suite.Keeper.GetReserveBalance()
	reserveBal := sdk.NewCoin(keeper.EvmDenom, sdkmath.NewInt(0))
	suite.Require().Equal(
		keeper.EvmDenom,
		reserveBal.Denom,
		"reserve balance should be in akava",
	)
	suite.Require().Equal(
		sdkmath.ZeroInt(),
		reserveBal.Amount.Quo(keeper.ConversionMultiplier),
		"reserve balance should not have any fractional parts",
	)

	totalFractionalBalances := sdk.ZeroInt()
	suite.Keeper.IterateAllAccounts(suite.Ctx, func(acc types.Account) bool {
		totalFractionalBalances = totalFractionalBalances.Add(acc.Balance)
		return false
	})

	suite.Require().Equal(
		reserveBal,
		totalFractionalBalances,
		"total fractional balances should equal reserve balance",
	)
}

func FuzzSendCoins(f *testing.F) {
	f.Add(int64(5_000_000_000_000_000_000), int64(1))
	f.Add(int64(3_000_000_000_000_000_000), int64(1_000_000_000_000_000_001))
	f.Add(int64(5_000_000_000_000_000_000), int64(5_000_000_000_000_000_000))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	f.Fuzz(func(t *testing.T, startBal int64, sendAmt int64) {
		suite := new(evmBankKeeperTestSuite2)
		suite.SetT(t)
		suite.SetS(suite)
		suite.SetupTest()

		suite.MintAkavaToAccount(addr1, sdkmath.NewInt(int64(startBal)))

		bk := keeper.NewEvmBankKeeper(
			suite.Keeper,
			suite.BankKeeper,
			suite.AccountKeeper,
		)

		sendCoins := sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, sdkmath.NewInt(int64(sendAmt))))
		err := bk.SendCoins(suite.Ctx, addr1, addr2, sendCoins)
		suite.Require().NoError(err)

		suite.VerifyReserveState()
	})
}
