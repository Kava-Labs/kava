package testutil

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/kava-labs/kava/tests/e2e/contracts/greeter"
	"github.com/kava-labs/kava/x/cdp/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

// InitKavaEvmData is run after the chain is running, but before the tests are run.
// It is used to initialize some EVM state, such as deploying contracts.
func (suite *E2eTestSuite) InitKavaEvmData() {
	whale := suite.Kava.GetAccount(FundedAccountName)

	// ensure funded account has nonzero erc20 balance
	balance := suite.Kava.GetErc20Balance(suite.DeployedErc20.Address, whale.EvmAddress)
	if balance.Cmp(big.NewInt(0)) != 1 {
		panic(fmt.Sprintf(
			"expected funded account (%s) to have erc20 balance of token %s",
			whale.EvmAddress.Hex(),
			suite.DeployedErc20.Address.Hex(),
		))
	}

	// expect the erc20 to be enabled for conversion to sdk.Coin
	params, err := suite.Kava.Grpc.Query.Evmutil.Params(context.Background(), &evmutiltypes.QueryParamsRequest{})
	if err != nil {
		panic(fmt.Sprintf("failed to fetch evmutil params during init: %s", err))
	}
	found := false
	erc20Addr := suite.DeployedErc20.Address.Hex()
	for _, p := range params.Params.EnabledConversionPairs {
		if common.BytesToAddress(p.KavaERC20Address).Hex() == erc20Addr {
			found = true
			suite.DeployedErc20.CosmosDenom = p.Denom
		}
	}
	if !found {
		panic(fmt.Sprintf("erc20 %s must be enabled for conversion to cosmos coin", erc20Addr))
	}
	suite.Kava.RegisterErc20(suite.DeployedErc20.Address)

	// expect the erc20's cosmos denom to be a supported cdp collateral type
	cdpParams, err := suite.Kava.Grpc.Query.Cdp.Params(context.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	found = false
	for _, cp := range cdpParams.Params.CollateralParams {
		if cp.Denom == suite.DeployedErc20.CosmosDenom {
			found = true
			suite.DeployedErc20.CdpCollateralType = cp.Type
		}
	}
	if !found {
		panic(fmt.Sprintf(
			"erc20's cosmos denom %s must be valid cdp collateral type",
			suite.DeployedErc20.CosmosDenom),
		)
	}

	// deploy an example contract
	greeterAddr, _, _, err := greeter.DeployGreeter(
		whale.evmSigner.Auth,
		whale.evmSigner.EvmClient,
		"what's up!",
	)
	suite.NoError(err, "failed to deploy a contract to the EVM")
	suite.Kava.ContractAddrs["greeter"] = greeterAddr
}

// FundKavaErc20Balance sends the pre-deployed ERC20 token to the `toAddress`.
func (suite *E2eTestSuite) FundKavaErc20Balance(toAddress common.Address, amount *big.Int) EvmTxResponse {
	// funded account should have erc20 balance
	whale := suite.Kava.GetAccount(FundedAccountName)
	res, err := whale.TransferErc20(suite.DeployedErc20.Address, toAddress, amount)
	suite.NoError(err)
	return res
}
