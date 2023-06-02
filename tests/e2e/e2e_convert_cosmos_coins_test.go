package e2e_test

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/kava-labs/kava/tests/util"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

func (suite *IntegrationTestSuite) TestConvertCosmosCoinsToFromERC20() {
	// we expect a denom to be registered to the allowed denoms param
	// and for the funded account to have a balance for that denom

	params, err := suite.Kava.Evmutil.Params(context.Background(), &evmutiltypes.QueryParamsRequest{})
	suite.NoError(err)
	suite.GreaterOrEqual(
		len(params.Params.AllowedCosmosDenoms), 1,
		"kava chain expected to have at least one AllowedCosmosDenom for ERC20 conversion",
	)

	initialAmount := int64(1e10)
	convertAmount := int64(5e9)
	tokenInfo := params.Params.AllowedCosmosDenoms[0]
	denom := tokenInfo.CosmosDenom
	initialFunds := sdk.NewCoins(
		sdk.NewInt64Coin(suite.Kava.StakingDenom, 1e6), // gas money
		sdk.NewInt64Coin(denom, initialAmount),         // conversion-enabled cosmos coin
	)
	initialModuleBalance := suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)

	user := suite.Kava.NewFundedAccount("cosmo-coin-converter", initialFunds)
	fee := sdk.NewCoins(ukava(7500))

	///////////////////////////////
	// CONVERT COSMOS COIN -> ERC20
	///////////////////////////////
	convertToErc20Msg := evmutiltypes.NewMsgConvertCosmosCoinToERC20(
		user.SdkAddress.String(),
		user.EvmAddress.Hex(),
		sdk.NewInt64Coin(denom, convertAmount),
	)
	tx := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&convertToErc20Msg},
		GasLimit:  2e6,
		FeeAmount: fee,
		Data:      "converting sdk coin to erc20",
	}
	res := user.SignAndBroadcastKavaTx(tx)
	suite.NoError(res.Err)

	// query for the deployed contract
	deployedContracts, err := suite.Kava.Evmutil.DeployedCosmosCoinContracts(
		context.Background(),
		&evmutiltypes.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: []string{denom}},
	)
	suite.NoError(err)
	suite.Len(deployedContracts.DeployedCosmosCoinContracts, 1)

	contractAddress := deployedContracts.DeployedCosmosCoinContracts[0].Address

	// check erc20 balance
	bz, err := suite.Kava.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress.Address,
		Data: util.BuildErc20BalanceOfCallData(user.EvmAddress),
	}, nil)
	suite.NoError(err)
	erc20Balance := new(big.Int).SetBytes(bz)
	suite.BigIntsEqual(big.NewInt(convertAmount), erc20Balance, "unexpected erc20 balance post-convert")

	// check cosmos coin is deducted from account
	expectedFunds := initialFunds.AmountOf(denom).SubRaw(convertAmount)
	balance := suite.Kava.QuerySdkForBalances(user.SdkAddress).AmountOf(denom)
	suite.Equal(expectedFunds, balance)

	// check that module account has sdk coins
	expectedModuleBalance := initialModuleBalance.AddRaw(convertAmount)
	actualModuleBalance := suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)
	suite.Equal(expectedModuleBalance, actualModuleBalance)

	///////////////////////////////
	// CONVERT ERC20 -> COSMOS COIN
	///////////////////////////////
	convertFromErc20Msg := evmutiltypes.NewMsgConvertCosmosCoinFromERC20(
		user.EvmAddress.Hex(),
		user.SdkAddress.String(),
		sdk.NewInt64Coin(denom, convertAmount),
	)

	tx = util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&convertFromErc20Msg},
		GasLimit:  2e5,
		FeeAmount: fee,
		Data:      "converting erc20 to cosmos coin",
	}
	res = user.SignAndBroadcastKavaTx(tx)
	suite.NoError(res.Err)

	// check erc20 balance
	bz, err = suite.Kava.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress.Address,
		Data: util.BuildErc20BalanceOfCallData(user.EvmAddress),
	}, nil)
	suite.NoError(err)
	erc20Balance = new(big.Int).SetBytes(bz)
	suite.BigIntsEqual(big.NewInt(0), erc20Balance, "expected all erc20 to be converted back")

	// check cosmos coin is added back to account
	expectedFunds = initialFunds.AmountOf(denom)
	balance = suite.Kava.QuerySdkForBalances(user.SdkAddress).AmountOf(denom)
	suite.Equal(expectedFunds, balance)

	// check that module account has sdk coins deducted
	actualModuleBalance = suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)
	suite.Equal(initialModuleBalance, actualModuleBalance)
}
