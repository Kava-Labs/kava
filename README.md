<p align="center">
  <img src="./kava-logo.svg" width="300">
</p>
<h3 align="center">DeFi for Crypto.</h3>

<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/kava-labs/kava)](https://goreportcard.com/report/github.com/kava-labs/kava)
[![API Reference](https://godoc.org/github.com/Kava-Labs/kava?status.svg)](https://godoc.org/github.com/Kava-Labs/kava)
[![GitHub](https://img.shields.io/github/license/kava-labs/kava.svg)](https://github.com/Kava-Labs/kava/blob/master/LICENSE.md)
[![Twitter Follow](https://img.shields.io/twitter/follow/kava_labs.svg?label=Follow&style=social)](https://twitter.com/kava_labs)
[![riot.im](https://img.shields.io/badge/riot.im-JOIN%20CHAT-green.svg)](https://riot.im/app/#/room/#kava-validators:matrix.org)

</div>

<div align="center">

### [Telegram](https://t.me/kavalabs) | [Medium](https://medium.com/kava-labs) | [Validator Chat](https://riot.im/app/#/room/#kava-validators:matrix.org)

### Participate in the Kava testnet and [snag a founder badge](./docs/REWARDS.md)!

</div>

## Installing

This guide assumes you have worked with `cosmos-sdk` blockchains previously. If you are just getting started, great! See the complete guide [here](https://medium.com/kava-labs).

#### Installing KVD

```
git clone https://github.com/Kava-Labs/kava.git
cd kava
go install ./cmd/kvd ./cmd/kvcli
```

#### Create a Wallet

```
kvd init --chain-id=kava-testnet-1 <your-moniker>
kvcli keys add <your_wallet_name>
```

**Be sure to back up your mnemonic!**

#### Create a Genesis Transaction

```
kvd add-genesis-account $(kvcli keys show <your_wallet_name> -a) 1000000kva
kvd gentx --name <your_wallet_name> --amount 1000000kva --ip <your-public-ip>
```

A genesis transaction should be written to `$HOME/.kvd/config/gentx/gentx-<gen_tx_hash>.json`

#### Submit Genesis Transaction

To be included in the genesis file for testnet one, fork this repo and copy your genesis transaction to the `testnet-1/gentx` directory. Submit your fork including your genesis transaction as a PR on this repo [here](https://github.com/Kava-Labs/kava/pulls)

#### Seed Nodes

We request known community members who wish to run public p2p seed nodes make pull requests to add community run seed nodes below.

```
Cosmostation - 0a47e347aacee74d4818090a0a94acf30cd8044e@13.124.101.116:26656
Forbole - c72c25d0b5e321b3f225b9be9f8aed0a7ca463db@34.66.3.247:26656
Ping - 3964d2f8c6c9a0ab6441134d2d423e8fc8af6899@kava-test.ping.pub:26656
Figment Network - 3c30ea1e2cdc422594e3b3d7ea73439730db8657@54.39.186.65:26656
Dokia Capital - 323e556dfb83147939d412527fc6286660438532@kava01.dokia.cloud:26656
01node - b1bcd6969f03940032f7f9c315ff3bbc1ee8cd20@185.181.103.135:26656
```

## License

Copyright © Kava Labs, Inc. All rights reserved.

Licensed under the [Apache v2 License](LICENSE).
