# Kava EVM contracts

Contracts for the Kava EVM used by the Kava blockchain.
Includes an ERC20 contract for wrapping native cosmos sdk.Coins.

## Setup

```
npm install
```

## Test

```
npm test
```

## Development

```
# Watch contract + tests
npm run dev

# Watch tests only
npm run test-watch
```

## Production compiling & Ethermint JSON

Ethermint uses its own json format that includes the ABI and bytecode in a single file. The bytecode should have no `0x` prefix and should be under the property name `bin`. This structure is built from the compiled code with `npm ethermint-json`.

The following compiles the contract, builds the ethermint json and copies the file to the evmutil:
```
npm build
```
