# end-2-end tests for kava

These tests use [`kvtool`](https://github.com/kava-labs/kvtool) to spin up a kava node configuration
and then runs tests against the running network.

Steps to run
1. Build a Kava docker image tagged `kava/kava:local`: `make build-docker`
2. Run the test suite: `make test-e2e`

**Note:** The suite will use your locally installed `kvtool` if present. If not present, it will be
installed. If `kvtool` is updated, you must reinstall it, or delete it from your PATH so the suite
can install it anew.

## Configuration

The test suite uses env variables that can be set in [`.env`](.env). See that file for a complete list
of options. The variables are parsed and imported into a `SuiteConfig` in [`testutil/config.go`](testutil/config.go).

The variables in `.env` will not override variables that are already present in the environment.
ie. Running `E2E_INCLUDE_IBC_TESTS=false make test-e2e` will disable the ibc tests regardless of how
the variable is set in `.env`.

## `Chain`s

A `testutil.Chain` is the abstraction around details, query clients, & signing accounts for interacting with a
network. After networks are running, a `Chain` is initialized & attached to the main test suite `testutil.E2eTestSuite`.

The primary Kava network is accessible via `suite.Kava`.

Details about the chains can be found [here](runner/chain.go#L62-84).

## `SigningAccount`s

Each `Chain` wraps a map of signing clients for that network. The `SigningAccount` contains clients
for both the Kava EVM and SDK co-chains.

The methods `SignAndBroadcastKavaTx` and `SignAndBroadcastEvmTx` are used to submit transactions to
the sdk and evm chains, respectively.

### Creating a new account
```go
// create an account on the Kava network, initially funded with 10 KAVA
acc := suite.Kava.NewFundedAccount("account-name", sdk.NewCoins(sdk.NewCoin("ukava", 10e6)))

// you can also access accounts by the name with which they were registered to the suite
acc := suite.Kava.GetAccount("account-name")
```

Funds for new accounts are distributed from the account with the mnemonic from the `E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC`
env variable. The account will be generated with HD coin type 60 & the `ethsecp256k1` private key signing algorithm.
The initial funding account is registered with the name `"whale"`.

## IBC tests

When IBC tests are enabled, an additional network is spun up with a different chain id & an IBC channel is
opened between it and the primary Kava network.

The IBC network runs kava with a different chain id and staking denom (see [runner/chain.go](runner/chain.go)).

The IBC chain queriers & accounts are accessible via `suite.Ibc`.

IBC tests can be disabled by setting `E2E_INCLUDE_IBC_TESTS` to `false`.
