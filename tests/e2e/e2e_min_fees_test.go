package e2e_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
)

func (suite *IntegrationTestSuite) TestEthGasPriceReturnsMinFee() {
	// read expected min fee from app.toml
	minGasPrices, err := getMinFeeFromAppToml(suite.KavaHomePath())
	suite.NoError(err)

	// evm uses akava, get akava min fee
	evmMinGas := minGasPrices.AmountOf("akava").TruncateInt().BigInt()

	// returns eth_gasPrice, units in kava
	gasPrice, err := suite.Kava.EvmClient.SuggestGasPrice(context.Background())
	suite.NoError(err)

	suite.Equal(evmMinGas, gasPrice)
}

func (suite *IntegrationTestSuite) TestEvmRespectsMinFee() {
	// setup sender & receiver
	sender := suite.Kava.NewFundedAccount("evm-min-fee-test-sender", sdk.NewCoins(ukava(1e3)))
	randoReceiver := util.SdkToEvmAddress(app.RandomAddress())

	// get min gas price for evm (from app.toml)
	minFees, err := getMinFeeFromAppToml(suite.KavaHomePath())
	suite.NoError(err)
	minGasPrice := minFees.AmountOf("akava").TruncateInt()

	// attempt tx with less than min gas price (min fee - 1)
	tooLowGasPrice := minGasPrice.Sub(sdk.OneInt()).BigInt()
	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(0, randoReceiver, big.NewInt(5e2), 1e5, tooLowGasPrice, nil),
		Data: "this tx should fail because it's gas price is too low",
	}
	res := sender.SignAndBroadcastEvmTx(req)

	// expect the tx to fail!
	suite.ErrorAs(res.Err, &util.ErrEvmFailedToBroadcast{})
	suite.ErrorContains(res.Err, "insufficient fee")
}

func getMinFeeFromAppToml(kavaHome string) (sdk.DecCoins, error) {
	// read the expected min gas price from app.toml
	parsed := struct {
		MinGasPrices string `toml:"minimum-gas-prices"`
	}{}
	appToml, err := os.ReadFile(filepath.Join(kavaHome, "config", "app.toml"))
	if err != nil {
		return nil, err
	}
	err = toml.Unmarshal(appToml, &parsed)
	if err != nil {
		return nil, err
	}

	// convert to dec coins
	return sdk.ParseDecCoins(strings.ReplaceAll(parsed.MinGasPrices, ";", ","))
}
