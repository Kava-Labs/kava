# Kava Local Development Tutorial

Kava provides tools for easily setting up a local Kava development environment for Solidity smart contract deployment and interaction.

## kvtool

kvtool enables developers to spin up a local Kava development environment. A dockerized local testnet with ethermint EVM can be created with:

```bash
kvtool testnet bootstrap --kava.configTemplate evm
```

The `faucet` Ethereum account is funded with sufficient gas tokens to be used in EVM transactions such as deploying smart contracts. Configuration file `hardhat.config.js` contains field `networks.local.accounts` which holds an array of private keys, including the private key for `faucet`. The private key was generated with:

```bash
# Add an alias to the dockerized kava blockchain
alias dkava='docker exec -it generated_kavanode_1 kava'

# Generate private key for named account 'faucet'
dkava keys unsafe-export-eth-key faucet --keyring-backend test
```

## Contract deployment

kvtool's `evm` directory contains an integrated Hardhat project for smart contract deployment and interaction.

```bash
# Navigate to the evm directory
cd evm

# Install dependencies
yarn
```

A sample Solidity smart contract located at `evm/contracts/token/Token.sol` can be easily deployed with:

```bash
yarn token-deploy
# TEST token deployed to: 0xff994EcA9074305D353E78a7159fE407537Cf87C
```

Under the hood `yarn token-deploy` is running cmd `npx hardhat run --network local scripts/deploy/token.js`. You can also use the `npx hardhat ...` syntax to interact with the contracts from the CLI.

## Contract interaction

Basic contract interaction is possible through additional scripts with the full list available in `evm/package.json`. Each yarn cmd is aliased to one of the Hardhat tasks found in `hardhat.config.js`.

```bash
# Get the deployed TEST token's total supply
yarn token-totalSupply --token [TOKEN_CONTRACT_ADDRESS]
# Total Supply is 100000000
```

You can get information about any cmd including param options by using the `--help` flag, for example:

```bash
yarn token-totalSupply --help
```

Using the available scripts, we can transfer tokens:

```bash
yarn token-transfer --token [TOKEN_CONTRACT_ADDRESS] --recipient [RECIPIENT_ADDRESS] --amount [AMOUNT]

# We can build a valid cmd by plugging in the relevant params:
yarn token-transfer --token 0xff994EcA9074305D353E78a7159fE407537Cf87C --recipient 0xdd2fd4581271e230360230f9337d5c0430bf44c0 --amount 50
# 0x5452643Ffa7Ab2BE7EF7C376f62FDfA7804Ed1d2 has transferred 50 TEST to 0xdd2fd4581271e230360230f9337d5c0430bf44c0
```

## Deploying your own contracts

You can easily deploy your own contracts:

1. Add your Solidity smart contract to `evm/contracts`.
2. Add a deployment script to `evm/scripts/deploy/[YOUR_SCRIPT].js` modeled on `evm/scripts/deploy/token.js`.
3. Add Hardhat tasks corresponding to your contract's public methods to the existing tasks in `hardhat.config.js`.

Done! Now you can deploy and interact with your contract;:

```bash
# Deployment
npx hardhat run --network local scripts/deploy/[YOUR_SCRIPT].js

# Interaction
npx hardhat [YOUR_TASK_NAME] --network local
```

If you want a cleaner CLI experience, you can add some script aliases to `evm/package.json` as seen in the `scripts` section.
