package e2e_test

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *IntegrationTestSuite) TestEthGasPriceReturnsMinFee() {
	// read expected min fee from app.toml
	minGasPrices, err := getMinFeeFromAppToml(suite.KavaHomePath())
	suite.NoError(err)

	// evm uses akava, get akava min fee
	evmMinGas := minGasPrices.AmountOf("akava").BigInt()
	// convert akava to kava (primary denom of EVM)
	evmMinGas = evmMinGas.Quo(evmMinGas, big.NewInt(1e18))

	fmt.Println("coins: ", minGasPrices)
	fmt.Println("amount akava: ", evmMinGas)

	// returns eth_gasPrice, units in kava
	gasPrice, err := suite.Kava.EvmClient.SuggestGasPrice(context.Background())
	suite.NoError(err)

	suite.Equal(evmMinGas, gasPrice)
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
