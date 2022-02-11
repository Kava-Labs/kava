# Connect MetaMask

Kava EVM is compatible with MetaMask, the most popular browser wallet.

## Metamask config

To access Kava EVM, you'll first need to add Kava's network configuration in MetaMask:

### Local

- Network Name: Localhost 8545
- New RPC URL: http://localhost:8545
- Chain ID: 8888
- Currency Symbol: Gas

<img src="https://user-images.githubusercontent.com/15370712/153627170-5715ece7-b3ea-4d26-b9b3-954bcc5660a4.png" width="485" height="391">

### Testnet

- Network Name:
- New RPC URL:
- Chain ID:
- Currency Symbol:

### Mainnet

- Network Name:
- New RPC URL:
- Chain ID:
- Currency Symbol:

## MetaMask account

To send transactions from your account using MetaMask, import your Kava account to MetaMask via private key. You can do so by clicking on the profile picture image and then selecting 'Import Account'. From the 'Select Type' dropdown, select 'Private Key'.

<img src="https://user-images.githubusercontent.com/15370712/153627252-15ef6da4-383a-4198-8f32-348b69ff21b3.png" width="239.3" height="400">

It's possible to generate your Kava EVM compatible private key using kvtool. From the `kvtool` repository:

```bash
# Start a local dockerized kava blockchain
kvtool testnet bootstrap --kava.configTemplate evm

# Add an alias to the dockerized kava blockchain
alias dkava='docker exec -it generated_kavanode_1 kava'

# Generate private key for named account 'validator'
dkava keys unsafe-export-eth-key validator --keyring-backend test
# 8A753652CB29472A01FAECF50EC1244C307B585466B2C52373539DA0100F3CED
```
