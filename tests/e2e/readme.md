# end-2-end tests for kava

These tests use [`kvtool`](https://github.com/kava-labs/kvtool) to spin up a kava node configuration
and then runs tests against the running network. It is a git sub-repository in this directory. If not
present, you must initialize the subrepo: `git submodule update --init`.

Steps to run
1. Ensure latest `kvtool` is installed: `make update-kvtool`
2. Run the test suite: `make test-e2e`
   This will build a docker image tagged `kava/kava:local` that will be run by kvtool.

**Note:** The suite will use your locally installed `kvtool` if present. If not present, it will be
installed. If the `kvtool` repo is updated, you must manually update your existing local binary: `make update-kvtool`

## Configuration

The test suite uses env variables that can be set in [`.env`](.env). See that file for a complete list
of options. The variables are parsed and imported into a `SuiteConfig` in [`testutil/config.go`](testutil/config.go).

The variables in `.env` will not override variables that are already present in the environment.
ie. Running `E2E_INCLUDE_IBC_TESTS=false make test-e2e` will disable the ibc tests regardless of how
the variable is set in `.env`.

### Running on Live Network

The end-to-end tests support being run on a live network. The primary toggle for setting up the suite to use a live network is the `E2E_RUN_KVTOOL_NETWORKS` flag. When set exactly to `false`, the configuration requires the following three environment variables:
* `E2E_KAVA_RPC_URL`
* `E2E_KAVA_GRPC_URL`
* `E2E_KAVA_EVM_RPC_URL`

See an example environment configuration with full description of all supported configurations in [`.env.live-network-example`](./.env.live-network-example). This example expects a local kvtool network to be running: `kvtool testnet bootstrap`.

When run against a live network, the suite will automatically return all the sdk funds sent to `SigningAccount`s on the chain, and will return any ERC20 balance from those accounts if the ERC20 is registered via `Chain.RegisterERC20`. The pre-deployed ERC20 that is required for the tests is registered on setup.

At this time, live-network tests do not support `E2E_INCLUDE_IBC_TESTS=true` and they do not support automated upgrades.

## `Chain`s

A `testutil.Chain` is the abstraction around details, query clients, & signing accounts for interacting with a
network. After networks are running, a `Chain` is initialized & attached to the main test suite `testutil.E2eTestSuite`.

The primary Kava network is accessible via `suite.Kava`.

Details about the chains can be found [here](runner/chain.go#L62-84).

## `SigningAccount`s

Each `Chain` wraps a map of signing clients for that network. The `SigningAccount` contains clients
for both the Kava EVM and Cosmos-Sdk co-chains.

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

## Chain Upgrades

When a named upgrade handler is included in the current working repo of Kava, the e2e test suite can
be configured to run all the tests on the upgraded chain. This includes the ability to add additional
tests to verify and do acceptance on the post-upgrade chain.

This configuration is controlled by the following env variables:
* `E2E_INCLUDE_AUTOMATED_UPGRADE` - toggles on the upgrade functionality. Must be set to `true`.
* `E2E_KAVA_UPGRADE_NAME` - the named upgrade, likely defined in [`app/upgrades.go`](../../app/upgrades.go)
* `E2E_KAVA_UPGRADE_HEIGHT` - the height at which to run the upgrade
* `E2E_KAVA_UPGRADE_BASE_IMAGE_TAG` - the [kava docker image tag](https://hub.docker.com/r/kava/kava/tags) to base the upgrade on

When all these are set, the chain is started with the binary contained in the docker image tagged
`E2E_KAVA_UPGRADE_BASE_IMAGE_TAG`. Then an upgrade proposal is submitted with the desired name and
height. The chain runs until that height and then is shutdown due to needing the upgrade. The chain
is restarted with the local repo's Kava code and the upgrade is run. Once completed, the whole test
suite is run.

For a full example of how this looks, see [this commit](https://github.com/Kava-Labs/kava/commit/5da48c892f0a5837141fc7de88632c7c68fff4ae)
on the [example/e2e-test-upgrade-handler](https://github.com/Kava-Labs/kava/tree/example/e2e-test-upgrade-handler) branch.
