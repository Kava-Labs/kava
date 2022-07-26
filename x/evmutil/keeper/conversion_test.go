package keeper_test

import (
	"math/big"
	"testing"

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
	suite.Require().Equal(sdk.NewCoin(pair.Denom, sdk.NewIntFromBigInt(amount)), coin)

	bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, recipient, pair.Denom)
	suite.Require().Equal(amount, bal.Amount.BigInt(), "minted amount should increase balance")
}

func (suite *ConversionTestSuite) TestBurn_InsufficientBalance() {
	pair := types.NewConversionPair(
		testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
		"erc20/usdc",
	)

	amount := sdk.NewInt(100)
	recipient := suite.Key1.PubKey().Address().Bytes()

	err := suite.Keeper.BurnConversionPairCoin(suite.Ctx, pair, sdk.NewCoin(pair.Denom, amount), recipient)
	suite.Require().Error(err)
	suite.Require().Equal("0erc20/usdc is smaller than 100erc20/usdc: insufficient funds", err.Error())
}

func (suite *ConversionTestSuite) TestBurn() {
	pair := types.NewConversionPair(
		testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
		"erc20/usdc",
	)

	amount := sdk.NewInt(100)
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
	suite.Require().Equal(sdk.NewCoin(pair.Denom, sdk.NewIntFromBigInt(amount)), coin)

	// Mint same initial balance for module account as backing erc20 supply
	err = suite.Keeper.MintERC20(
		suite.Ctx,
		pair.GetAddress(), // contractAddr
		moduleAddr,        //receiver
		amount,
	)
	suite.Require().NoError(err)

	err = suite.Keeper.ConvertCoinToERC20(
		suite.Ctx,
		originAcc,
		recipientAcc,
		sdk.NewCoin(pair.Denom, sdk.NewIntFromBigInt(amount)),
	)
	suite.Require().NoError(err)

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
		sdk.NewCoin(pair.Denom, sdk.NewIntFromBigInt(amount)),
	)

	suite.Require().Error(err)
	suite.Require().Equal("0erc20/usdc is smaller than 100erc20/usdc: insufficient funds", err.Error())
}

func (suite *ConversionTestSuite) TestConvertCoinToERC20_NotEnabled() {
	_ = suite.DeployERC20()
	// First deploy is already in params, second is new address
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
		sdk.NewCoin(pair.Denom, sdk.NewIntFromBigInt(amount)),
	)

	suite.Require().Error(err)
	suite.Require().Equal("erc20/notenabled: ERC20 token not enabled to convert to sdk.Coin", err.Error())
}
