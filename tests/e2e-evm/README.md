# E2E EVM

This is a E2E test suite focused on testing EVM and EVM <> Cosmos integrations for the Kava protocol and blockchain.

This test suite uses viem as the main API used for interacting with the EVM and Ethereum JSON RPC endpoints.

## Networks

The test suite runs on multiple networks to test compatibility with other EVM's and offer extended testing capabilities when required.

### Hardhat

```
npx hardhat test --network hardhat
```

### Kvtool

```
npx hardhat test --network kvtool
```
