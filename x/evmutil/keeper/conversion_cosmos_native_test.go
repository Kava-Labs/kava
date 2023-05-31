package keeper_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type convertCosmosCoinToERC20Suite struct {
	testutil.Suite
}

func TestConversionCosmosNativeToEvmSuite(t *testing.T) {
	suite.Run(t, new(convertCosmosCoinToERC20Suite))
}

// fail test if contract for denom not registered
func (suite *convertCosmosCoinToERC20Suite) denomContractRegistered(denom string) types.InternalEVMAddress {
	contractAddress, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
	suite.True(found)
	return contractAddress
}

// fail test if contract for denom IS registered
func (suite *convertCosmosCoinToERC20Suite) denomContractNotRegistered(denom string) {
	_, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
	suite.False(found)
}

// more tests of tests of this method are made to the msg handler, see ./msg_server_test.go
func (suite *convertCosmosCoinToERC20Suite) TestConvertCosmosCoinToERC20() {
	allowedDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	initialFunding := sdk.NewInt64Coin(allowedDenom, int64(1e10))
	initiator := app.RandomAddress()

	amount := sdk.NewInt64Coin(allowedDenom, 6e8)
	receiver1 := types.BytesToInternalEVMAddress(app.RandomAddress().Bytes())
	receiver2 := types.BytesToInternalEVMAddress(app.RandomAddress().Bytes())

	var contractAddress types.InternalEVMAddress

	caller, key := testutil.RandomEvmAccount()
	query := func(method string, args ...interface{}) ([]interface{}, error) {
		return suite.QueryContract(
			types.ERC20KavaWrappedCosmosCoinContract.ABI,
			caller,
			key,
			contractAddress,
			method,
			args...,
		)
	}
	checkTotalSupply := func(expectedSupply sdkmath.Int) {
		res, err := query("totalSupply")
		suite.NoError(err)
		suite.Len(res, 1)
		suite.BigIntsEqual(expectedSupply.BigInt(), res[0].(*big.Int), "unexpected total supply")
	}
	checkBalanceOf := func(address types.InternalEVMAddress, expectedBalance sdkmath.Int) {
		res, err := query("balanceOf", address.Address)
		suite.NoError(err)
		suite.Len(res, 1)
		suite.BigIntsEqual(expectedBalance.BigInt(), res[0].(*big.Int), fmt.Sprintf("unexpected balanceOf for %s", address))
	}

	suite.SetupTest()

	suite.Run("fails when denom not allowed", func() {
		suite.denomContractNotRegistered(allowedDenom)
		err := suite.Keeper.ConvertCosmosCoinToERC20(
			suite.Ctx,
			initiator,
			receiver1,
			sdk.NewCoin(allowedDenom, sdkmath.NewInt(6e8)),
		)
		suite.ErrorContains(err, "sdk.Coin not enabled to convert to ERC20 token")
		suite.denomContractNotRegistered(allowedDenom)
	})

	suite.Run("allowed denoms have contract deploys on first conversion", func() {
		// make the denom allowed for conversion
		params := suite.Keeper.GetParams(suite.Ctx)
		params.AllowedCosmosDenoms = types.NewAllowedCosmosCoinERC20Tokens(
			types.NewAllowedCosmosCoinERC20Token(allowedDenom, "Kava EVM Atom", "ATOM", 6),
		)
		suite.Keeper.SetParams(suite.Ctx, params)

		// fund account
		err := suite.App.FundAccount(suite.Ctx, initiator, sdk.NewCoins(initialFunding))
		suite.NoError(err, "failed to initially fund account")

		// first conversion
		err = suite.Keeper.ConvertCosmosCoinToERC20(
			suite.Ctx,
			initiator,
			receiver1,
			sdk.NewCoin(allowedDenom, sdkmath.NewInt(6e8)),
		)
		suite.NoError(err)

		// contract should be deployed & registered
		contractAddress = suite.denomContractRegistered(allowedDenom)

		// sdk coin deducted from initiator
		expectedBalance := initialFunding.Sub(amount)
		balance := suite.BankKeeper.GetBalance(suite.Ctx, initiator, allowedDenom)
		suite.Equal(expectedBalance, balance)

		// erc20 minted to receiver
		checkBalanceOf(receiver1, amount.Amount)
		// total supply of erc20 should have increased
		checkTotalSupply(amount.Amount)

		// event should be emitted
		suite.EventsContains(suite.GetEvents(),
			sdk.NewEvent(
				types.EventTypeConvertCosmosCoinToERC20,
				sdk.NewAttribute(types.AttributeKeyInitiator, initiator.String()),
				sdk.NewAttribute(types.AttributeKeyReceiver, receiver1.String()),
				sdk.NewAttribute(types.AttributeKeyERC20Address, contractAddress.Hex()),
				sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			),
		)
	})

	suite.Run("2nd deploy uses same contract", func() {
		// expect no initial balance
		checkBalanceOf(receiver2, sdkmath.NewInt(0))

		// 2nd conversion
		err := suite.Keeper.ConvertCosmosCoinToERC20(
			suite.Ctx,
			initiator,
			receiver2,
			sdk.NewCoin(allowedDenom, sdkmath.NewInt(6e8)),
		)
		suite.NoError(err)

		// contract address should not change
		convertTwiceContractAddress := suite.denomContractRegistered(allowedDenom)
		suite.Equal(contractAddress, convertTwiceContractAddress)

		// sdk coin deducted from initiator
		expectedBalance := initialFunding.Sub(amount).Sub(amount)
		balance := suite.BankKeeper.GetBalance(suite.Ctx, initiator, allowedDenom)
		suite.Equal(expectedBalance, balance)

		// erc20 minted to receiver
		checkBalanceOf(receiver2, amount.Amount)
		// total supply of erc20 should have increased
		checkTotalSupply(amount.Amount.MulRaw(2))

		// event should be emitted
		suite.EventsContains(suite.GetEvents(),
			sdk.NewEvent(
				types.EventTypeConvertCosmosCoinToERC20,
				sdk.NewAttribute(types.AttributeKeyInitiator, initiator.String()),
				sdk.NewAttribute(types.AttributeKeyReceiver, receiver2.String()),
				sdk.NewAttribute(types.AttributeKeyERC20Address, contractAddress.Hex()),
				sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			),
		)
	})
}

type convertCosmosCoinFromERC20Suite struct {
	testutil.Suite

	denom     string
	initiator types.InternalEVMAddress
	receiver  sdk.AccAddress

	contractAddress types.InternalEVMAddress
	initialPosition sdk.Coin

	query func(method string, args ...interface{}) ([]interface{}, error)
}

func (suite *convertCosmosCoinFromERC20Suite) SetupTest() {
	var err error
	suite.Suite.SetupTest()

	suite.denom = "magic"
	suite.initiator = testutil.RandomInternalEVMAddress()
	suite.receiver = app.RandomAddress()

	// manually create an initial position - sdk coin locked in module
	suite.initialPosition = sdk.NewInt64Coin(suite.denom, 1e12)
	err = suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(suite.initialPosition))
	suite.NoError(err)

	// deploy erc20 contract for the denom
	tokenInfo := types.AllowedCosmosCoinERC20Token{
		CosmosDenom: suite.denom,
		Name:        "Test Token",
		Symbol:      "MAGIC",
		Decimals:    6,
	}
	suite.contractAddress, err = suite.Keeper.GetOrDeployCosmosCoinERC20Contract(suite.Ctx, tokenInfo)
	suite.NoError(err)

	// manually create an initial position - minted tokens
	err = suite.Keeper.MintERC20(suite.Ctx, suite.contractAddress, suite.initiator, suite.initialPosition.Amount.BigInt())
	suite.NoError(err)

	caller, key := testutil.RandomEvmAccount()
	suite.query = func(method string, args ...interface{}) ([]interface{}, error) {
		return suite.QueryContract(
			types.ERC20KavaWrappedCosmosCoinContract.ABI,
			caller,
			key,
			suite.contractAddress,
			method,
			args...,
		)
	}
}

func (suite *convertCosmosCoinFromERC20Suite) checkTotalSupply(expectedSupply sdkmath.Int) {
	res, err := suite.query("totalSupply")
	suite.NoError(err)
	suite.Len(res, 1)
	suite.BigIntsEqual(expectedSupply.BigInt(), res[0].(*big.Int), "unexpected total supply")
}

func (suite *convertCosmosCoinFromERC20Suite) checkBalanceOf(address types.InternalEVMAddress, expectedBalance sdkmath.Int) {
	res, err := suite.query("balanceOf", address.Address)
	suite.NoError(err)
	suite.Len(res, 1)
	suite.BigIntsEqual(expectedBalance.BigInt(), res[0].(*big.Int), fmt.Sprintf("unexpected balanceOf for %s", address))
}

func TestConversionCosmosNativeFromEVMSuite(t *testing.T) {
	suite.Run(t, new(convertCosmosCoinFromERC20Suite))
}

func (suite *convertCosmosCoinFromERC20Suite) TestConvertCosmosCoinFromERC20_NoContractDeployed() {
	err := suite.Keeper.ConvertCosmosCoinFromERC20(
		suite.Ctx,
		suite.initiator,
		suite.receiver,
		sdk.NewInt64Coin("unsupported-denom", 1e6),
	)
	suite.ErrorContains(err, "no erc20 contract found for unsupported-denom")
}

func (suite *convertCosmosCoinFromERC20Suite) TestConvertCosmosCoinFromERC20() {
	// half the initial position
	amount := suite.initialPosition.SubAmount(suite.initialPosition.Amount.QuoRaw(2))

	suite.Run("partial withdraw", func() {
		err := suite.Keeper.ConvertCosmosCoinFromERC20(
			suite.Ctx,
			suite.initiator,
			suite.receiver,
			amount,
		)
		suite.NoError(err)

		suite.checkTotalSupply(amount.Amount)
		suite.checkBalanceOf(suite.initiator, amount.Amount)
		suite.App.CheckBalance(suite.T(), suite.Ctx, suite.receiver, sdk.NewCoins(amount))
	})

	suite.Run("full withdraw", func() {
		err := suite.Keeper.ConvertCosmosCoinFromERC20(
			suite.Ctx,
			suite.initiator,
			suite.receiver,
			amount,
		)
		suite.NoError(err)

		// expect no remaining erc20 balance
		suite.checkTotalSupply(sdkmath.ZeroInt())
		suite.checkBalanceOf(suite.initiator, sdkmath.ZeroInt())
		// expect full amount withdrawn to receiver
		suite.App.CheckBalance(suite.T(), suite.Ctx, suite.receiver, sdk.NewCoins(suite.initialPosition))
	})

	suite.Run("insufficient balance", func() {
		err := suite.Keeper.ConvertCosmosCoinFromERC20(
			suite.Ctx,
			suite.initiator,
			suite.receiver,
			amount,
		)
		suite.ErrorContains(err, "failed to convert to cosmos coins: insufficient funds")
	})
}
