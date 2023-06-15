package testutil

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/kava-labs/kava/tests/e2e/contracts/greeter"
	"github.com/kava-labs/kava/tests/util"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

// InitKavaEvmData is run after the chain is running, but before the tests are run.
// It is used to initialize some EVM state, such as deploying contracts.
func (suite *E2eTestSuite) InitKavaEvmData() {
	whale := suite.Kava.GetAccount(FundedAccountName)

	// ensure funded account has nonzero erc20 balance
	balance := suite.Kava.GetErc20Balance(suite.DeployedErc20.Address, whale.EvmAddress)
	if balance.Cmp(big.NewInt(0)) != 1 {
		panic(fmt.Sprintf("expected funded account (%s) to have erc20 balance", whale.EvmAddress.Hex()))
	}

	// expect the erc20 to be enabled for conversion to sdk.Coin
	params, err := suite.Kava.Evmutil.Params(context.Background(), &evmutiltypes.QueryParamsRequest{})
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

	// deploy an example contract
	greeterAddr, _, _, err := greeter.DeployGreeter(
		whale.evmSigner.Auth,
		whale.evmSigner.EvmClient,
		"what's up!",
	)
	suite.NoError(err, "failed to deploy a contract to the EVM")
	suite.Kava.ContractAddrs["greeter"] = greeterAddr
}

func (suite *E2eTestSuite) FundKavaErc20Balance(toAddress common.Address, amount *big.Int) EvmTxResponse {
	// funded account should have erc20 balance
	whale := suite.Kava.GetAccount(FundedAccountName)

	data := util.BuildErc20TransferCallData(toAddress, amount)
	nonce, err := suite.Kava.EvmClient.PendingNonceAt(context.Background(), whale.EvmAddress)
	suite.NoError(err)

	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(nonce, suite.DeployedErc20.Address, big.NewInt(0), 1e5, big.NewInt(1e10), data),
		Data: fmt.Sprintf("fund %s with ERC20 balance (%s)", toAddress.Hex(), amount.String()),
	}

	return whale.SignAndBroadcastEvmTx(req)
}
