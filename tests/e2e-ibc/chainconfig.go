package main

import (
	"os"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

func DefaultKavaChainConfig(chainId string) ibc.ChainConfig {
	kavaImageTag := os.Getenv("KAVA_TAG")
	if kavaImageTag == "" {
		kavaImageTag = "v0.26.0-rocksdb"
	}

	// app.toml overrides
	jsonRpcOverrides := make(testutil.Toml)
	jsonRpcOverrides["address"] = "0.0.0.0:8545"
	appTomlOverrides := make(testutil.Toml)
	appTomlOverrides["json-rpc"] = jsonRpcOverrides

	// genesis param overrides
	genesis := []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.evm.params.evm_denom", "akava"),
	}

	return ibc.ChainConfig{
		Type:                  "cosmos",
		ChainID:               chainId,
		Images:                []ibc.DockerImage{{Repository: "kava/kava", Version: kavaImageTag, UidGid: "0:0"}},
		Bin:                   "kava",
		Bech32Prefix:          "kava",
		Denom:                 "ukava",
		GasPrices:             "0ukava", // 0 gas price makes calculating expected balances simpler
		GasAdjustment:         1.5,
		TrustingPeriod:        "168h0m0s",
		ModifyGenesis:         cosmos.ModifyGenesis(genesis),
		ExposeAdditionalPorts: []string{"8545/tcp"},
		ConfigFileOverrides:   map[string]any{"config/app.toml": appTomlOverrides},
	}
}
