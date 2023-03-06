package testutil

import (
	"fmt"
	"os"
	"strconv"

	"github.com/subosito/gotenv"
)

func init() {
	// read the .env file, if present
	gotenv.Load()
}

type SuiteConfig struct {
	// A funded account used to fnd all other accounts.
	FundedAccountMnemonic string
	// Whether or not to start an IBC chain. Use `suite.SkipIfIbcDisabled()` in IBC tests in IBC tests.
	IncludeIbcTests bool
	// When true, the chains will remain running after tests complete (pass or fail)
	SkipShutdown bool
}

func ParseSuiteConfig() SuiteConfig {
	// this mnemonic is expected to be a funded account that can seed the funds for all
	// new accounts created during tests. it will be available under Accounts["whale"]
	fundedAccountMnemonic := os.Getenv("E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC")
	if fundedAccountMnemonic == "" {
		panic("no E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC provided")
	}
	var skipShutdown bool
	skipShutdownEnv := os.Getenv("E2E_SKIP_SHUTDOWN")
	if skipShutdownEnv != "" {
		skipShutdown = mustParseBool("E2E_SKIP_SHUTDOWN")
	}
	return SuiteConfig{
		FundedAccountMnemonic: fundedAccountMnemonic,
		IncludeIbcTests:       mustParseBool("E2E_INCLUDE_IBC_TESTS"),
		SkipShutdown:          skipShutdown,
	}
}

func mustParseBool(name string) bool {
	envValue := os.Getenv(name)
	if envValue == "" {
		panic(fmt.Sprintf("%s is unset but expected a bool", name))
	}
	value, err := strconv.ParseBool(envValue)
	if err != nil {
		panic(fmt.Sprintf("%s (%s) cannot be parsed to a bool: %s", name, envValue, err))
	}
	return value
}
