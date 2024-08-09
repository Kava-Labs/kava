# Precompile

The `precompile` package defines stateful precompile contracts used for creating precompile contracts that enhance the EVM, enable interactions between cosmos state and EVM state, or are used soley for testing purposes.

This package is made of two subpackages:

- `contracts` - Defines stateful precompiles and their constructors.
- `registry` - Defines stateful precompile addresses and registers them with the global registry
 defined at `github.com/kava-labs/go-ethereum/precompile/modules`.

This is architected to isolate the dependency on the global registry (`github.com/kava-labs/go-ethereum/precompile/modules` package) to as few places as possible, have one source of truth for address registration (see `./registry/registry.go`), and isolate registration testing to one package.

In order to use the precompile registry, it must be imported for it's init function to run and register the precompiles.  For the kava app, this is done in the `app.go` file with the import `_ "github.com/kava-labs/kava/precompile/registry"`.  This is done in the `app.go` since this is the main file used for app dependencies and so all modules using the app cause the registry to be loaded.  This is important for any consumers of the app outside of the kava cmd, as well as test code using the app for integration and unit testing.

## Defining a new precompile

1) Add the expected 0x address to the expected list in `./registry/registry_test.go`.
2) Create a new sub-directory under `./contracts` with a `contract_test.go` and `contract.go` file.
3) Implement `NewContract` function with associated tests in contract and contract test files.
4) Add the contract registration to `./registry/registry.go`.

## Contracts

### Noop

This contract is used for testing purposes only and should not be used on public chains.  The functions of this contract (once implemented), will be used to exercise and test the various aspects of the EVM such as gas usage, argument parsing, events, etc. The specific operations tested under this contract are still to be determined.

