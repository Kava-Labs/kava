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

type ConversionCosmosNativeSuite struct {
	testutil.Suite
}

func TestConversionCosmosNativeSuite(t *testing.T) {
	suite.Run(t, new(ConversionCosmosNativeSuite))
}

// fail test if contract for denom not registered
func (suite *ConversionCosmosNativeSuite) denomContractRegistered(denom string) types.InternalEVMAddress {
	contractAddress, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
	suite.True(found)
	return contractAddress
}

// fail test if contract for denom IS registered
func (suite *ConversionCosmosNativeSuite) denomContractNotRegistered(denom string) {
	_, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
	suite.False(found)
}

// more tests of tests of this method are made to the msg handler, see ./msg_server_test.go
func (suite *ConversionCosmosNativeSuite) TestConvertCosmosCoinToERC20() {
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
