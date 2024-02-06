package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type ConversionTestSuite struct {
	testutil.Suite
}

func TestConversionTestSuite(t *testing.T) {
	suite.Run(t, new(ConversionTestSuite))
}

func (suite *ConversionTestSuite) TestMint() {
	pair := types.NewConversionPair(
		testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
		"erc20/usdc",
	)

	amount := big.NewInt(100)
	recipient := suite.Key1.PubKey().Address().Bytes()

	coin, err := suite.Keeper.MintConversionPairCoin(suite.Ctx, pair, amount, recipient)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(pair.Denom, sdkmath.NewIntFromBigInt(amount)), coin)

	bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, recipient, pair.Denom)
	suite.Require().Equal(amount, bal.Amount.BigInt(), "minted amount should increase balance")
}

func (suite *ConversionTestSuite) TestBurn_InsufficientBalance() {
	pair := types.NewConversionPair(
		testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
		"erc20/usdc",
	)

	amount := sdkmath.NewInt(100)
	recipient := suite.Key1.PubKey().Address().Bytes()

	err := suite.Keeper.BurnConversionPairCoin(suite.Ctx, pair, sdk.NewCoin(pair.Denom, amount), recipient)
	suite.Require().Error(err)
	suite.Require().Equal("spendable balance  is smaller than 100erc20/usdc: insufficient funds", err.Error())
}

func (suite *ConversionTestSuite) TestBurn() {
	pair := types.NewConversionPair(
		testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
		"erc20/usdc",
	)

	amount := sdkmath.NewInt(100)
	recipient := suite.Key1.PubKey().Address().Bytes()

	coin, err := suite.Keeper.MintConversionPairCoin(suite.Ctx, pair, amount.BigInt(), recipient)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(pair.Denom, amount), coin)

	bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, recipient, pair.Denom)
	suite.Require().Equal(amount, bal.Amount, "minted amount should increase balance")

	err = suite.Keeper.BurnConversionPairCoin(suite.Ctx, pair, sdk.NewCoin(pair.Denom, amount), recipient)
	suite.Require().NoError(err)

	bal = suite.App.GetBankKeeper().GetBalance(suite.Ctx, recipient, pair.Denom)
	suite.Require().Equal(sdk.ZeroInt(), bal.Amount, "balance should be zero after burn")
}

func (suite *ConversionTestSuite) TestUnlockERC20Tokens() {
	contractAddr := suite.DeployERC20()

	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	amount := big.NewInt(100)
	recipient := types.NewInternalEVMAddress(common.BytesToAddress(suite.Key1.PubKey().Address()))
	moduleAddr := types.NewInternalEVMAddress(types.ModuleEVMAddress)

	// Mint some initial balance for module account to transfer
	err := suite.Keeper.MintERC20(
		suite.Ctx,
		pair.GetAddress(), // contractAddr
		moduleAddr,        //receiver
		amount,
	)
	suite.Require().NoError(err)

	err = suite.Keeper.UnlockERC20Tokens(suite.Ctx, pair, amount, recipient)
	suite.Require().NoError(err)

	// Check balance of recipient
	bal := suite.GetERC20BalanceOf(
		types.ERC20MintableBurnableContract.ABI,
		pair.GetAddress(),
		recipient,
	)
	suite.Require().Equal(amount, bal, "balance should increase by unlock amount")

	// Check balance of module account
	bal = suite.GetERC20BalanceOf(
		types.ERC20MintableBurnableContract.ABI,
		pair.GetAddress(),
		moduleAddr,
	)
	suite.Require().Equal(
		// String() due to non-equal struct values for 0
		big.NewInt(0).String(),
		bal.String(),
		"balance should decrease module account by unlock amount",
	)
}

func (suite *ConversionTestSuite) TestUnlockERC20Tokens_Insufficient() {
	contractAddr := suite.DeployERC20()

	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	amount := big.NewInt(100)
	recipient := types.NewInternalEVMAddress(common.BytesToAddress(suite.Key1.PubKey().Address()))

	// Module account has 0 balance, cannot unlock
	err := suite.Keeper.UnlockERC20Tokens(suite.Ctx, pair, amount, recipient)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "execution reverted: ERC20: transfer amount exceeds balance")
}

func (suite *ConversionTestSuite) TestConvertCoinToERC20() {
	contractAddr := suite.DeployERC20()

	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	amount := big.NewInt(100)
	originAcc := sdk.AccAddress(suite.Key1.PubKey().Address().Bytes())
	recipientAcc := types.NewInternalEVMAddress(common.BytesToAddress(suite.Key2.PubKey().Address()))
	moduleAddr := types.NewInternalEVMAddress(types.ModuleEVMAddress)

	// Starting balance of origin account
	coin, err := suite.Keeper.MintConversionPairCoin(suite.Ctx, pair, amount, originAcc)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(pair.Denom, sdkmath.NewIntFromBigInt(amount)), coin)

	// Mint same initial balance for module account as backing erc20 supply
	err = suite.Keeper.MintERC20(
		suite.Ctx,
		pair.GetAddress(), // contractAddr
		moduleAddr,        //receiver
		amount,
	)
	suite.Require().NoError(err)

	// convert coin to erc20
	ctx := suite.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	err = suite.Keeper.ConvertCoinToERC20(
		ctx,
		originAcc,
		recipientAcc,
		sdk.NewCoin(pair.Denom, sdkmath.NewIntFromBigInt(amount)),
	)
	suite.Require().NoError(err)
	suite.Require().LessOrEqual(ctx.GasMeter().GasConsumed(), uint64(500000))
	suite.Require().GreaterOrEqual(ctx.GasMeter().GasConsumed(), uint64(50000))

	// Source should decrease
	bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, originAcc, pair.Denom)
	suite.Require().Equal(sdk.ZeroInt(), bal.Amount, "conversion should decrease source balance")

	// Module bal should also decrease
	moduleBal := suite.GetERC20BalanceOf(
		types.ERC20MintableBurnableContract.ABI,
		pair.GetAddress(),
		moduleAddr,
	)
	suite.Require().Equal(
		// String() due to non-equal struct values for 0
		big.NewInt(0).String(),
		moduleBal.String(),
		"balance should decrease module account by unlock amount",
	)

	// Recipient balance should increase by same amount
	recipientBal := suite.GetERC20BalanceOf(
		types.ERC20MintableBurnableContract.ABI,
		pair.GetAddress(),
		recipientAcc,
	)
	suite.Require().Equal(
		// String() due to non-equal struct values for 0
		amount,
		recipientBal,
		"recipient balance should increase",
	)

	suite.EventsContains(suite.GetEvents(),
		sdk.NewEvent(
			types.EventTypeConvertCoinToERC20,
			sdk.NewAttribute(types.AttributeKeyInitiator, originAcc.String()),
			sdk.NewAttribute(types.AttributeKeyReceiver, recipientAcc.String()),
			sdk.NewAttribute(types.AttributeKeyERC20Address, pair.GetAddress().String()),
			sdk.NewAttribute(types.AttributeKeyAmount, coin.String()),
		))
}

func (suite *ConversionTestSuite) TestConvertCoinToERC20_InsufficientBalance() {
	contractAddr := suite.DeployERC20()

	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	amount := big.NewInt(100)
	originAcc := sdk.AccAddress(suite.Key1.PubKey().Address().Bytes())
	recipientAcc := types.NewInternalEVMAddress(common.BytesToAddress(suite.Key2.PubKey().Address()))

	err := suite.Keeper.ConvertCoinToERC20(
		suite.Ctx,
		originAcc,
		recipientAcc,
		sdk.NewCoin(pair.Denom, sdkmath.NewIntFromBigInt(amount)),
	)

	suite.Require().Error(err)
	suite.Require().Equal("spendable balance  is smaller than 100erc20/usdc: insufficient funds", err.Error())
}

func (suite *ConversionTestSuite) TestConvertCoinToERC20_NotEnabled() {
	contractAddr := suite.DeployERC20()

	pair := types.NewConversionPair(
		contractAddr,
		"erc20/notenabled",
	)

	amount := big.NewInt(100)
	originAcc := sdk.AccAddress(suite.Key1.PubKey().Address().Bytes())
	recipientAcc := types.NewInternalEVMAddress(common.BytesToAddress(suite.Key2.PubKey().Address()))

	err := suite.Keeper.ConvertCoinToERC20(
		suite.Ctx,
		originAcc,
		recipientAcc,
		sdk.NewCoin(pair.Denom, sdkmath.NewIntFromBigInt(amount)),
	)

	suite.Require().Error(err)
	suite.Require().Equal("erc20/notenabled: ERC20 token not enabled to convert to sdk.Coin", err.Error())
}

func (suite *ConversionTestSuite) TestConvertERC20ToCoin() {
	contractAddr := suite.DeployERC20()

	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	totalAmt := big.NewInt(100)
	userAddr := sdk.AccAddress(suite.Key1.PubKey().Address().Bytes())
	userEvmAddr := types.NewInternalEVMAddress(common.BytesToAddress(suite.Key1.PubKey().Address()))

	// Mint same initial balance for user account
	err := suite.Keeper.MintERC20(
		suite.Ctx,
		pair.GetAddress(), // contractAddr
		userEvmAddr,       //receiver
		totalAmt,
	)
	suite.Require().NoError(err)

	ctx := suite.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	convertAmt := sdkmath.NewInt(50)
	err = suite.Keeper.ConvertERC20ToCoin(
		ctx,
		userEvmAddr,
		userAddr,
		pair.GetAddress(),
		convertAmt,
	)
	suite.Require().NoError(err)
	suite.Require().LessOrEqual(ctx.GasMeter().GasConsumed(), uint64(500000))
	suite.Require().GreaterOrEqual(ctx.GasMeter().GasConsumed(), uint64(50000))

	// bank balance should decrease
	bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, userAddr, pair.Denom)
	suite.Require().Equal(convertAmt, bal.Amount, "conversion should decrease source balance")

	// Module bal should also decrease
	userBal := suite.GetERC20BalanceOf(
		types.ERC20MintableBurnableContract.ABI,
		pair.GetAddress(),
		userEvmAddr,
	)
	suite.Require().Equal(
		// String() due to non-equal struct values for 0
		big.NewInt(50).String(),
		userBal.String(),
		"balance should decrease module account by unlock amount",
	)

	suite.EventsContains(suite.GetEvents(),
		sdk.NewEvent(
			types.EventTypeConvertERC20ToCoin,
			sdk.NewAttribute(types.AttributeKeyERC20Address, pair.GetAddress().String()),
			sdk.NewAttribute(types.AttributeKeyInitiator, userEvmAddr.String()),
			sdk.NewAttribute(types.AttributeKeyReceiver, userAddr.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, sdk.NewCoin(pair.Denom, convertAmt).String()),
		),
	)
}

func (suite *ConversionTestSuite) TestConvertERC20ToCoin_EmptyContract() {
	contractAddr := testutil.MustNewInternalEVMAddressFromString("0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3")
	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	userAddr := sdk.AccAddress(suite.Key1.PubKey().Address().Bytes())
	userEvmAddr := types.NewInternalEVMAddress(common.BytesToAddress(suite.Key1.PubKey().Address()))
	convertAmt := sdkmath.NewInt(100)

	// Trying to convert erc20 from an empty contract should fail
	err := suite.Keeper.ConvertERC20ToCoin(
		suite.Ctx,
		userEvmAddr,
		userAddr,
		pair.GetAddress(),
		convertAmt,
	)
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "failed to retrieve balance: failed to unpack method balanceOf")

	// bank balance should not change
	bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, userAddr, pair.Denom)
	suite.Require().Equal(sdk.ZeroInt(), bal.Amount)
}
