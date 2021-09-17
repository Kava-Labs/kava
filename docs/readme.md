---
title: Kava Documentation
description: The documentation of Kava.
footer:
  newsletter: false
aside: true
---

# Kava

## DeFi for Crypto.

Reference implementation of Kava, a blockchain for cross-chain DeFi. Built using the [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

## Mainnet

The current recommended version of the software for mainnet is [v0.15.1](https://github.com/Kava-Labs/kava/releases/tag/v0.15.1). The master branch of this repository often contains considerable development work since the last mainnet release and is __not__ runnable on mainnet.

### Installation

```bash
git checkout v0.15.1
make install
```

### Upgrade

The scheduled mainnet upgrade to `kava-8` took place on August 30th, 2021 at 15:00 UTC. The current version of Kava for `kava-8` is [__v0.15.1__](https://github.com/Kava-Labs/kava/releases/tag/v0.15.1).

The canonical genesis file can be found [here](https://github.com/Kava-Labs/launch/tree/master/kava-8)

The canonical genesis file hash is

```
jq -S -c -M '' genesis.json | shasum -a 256
f0c90f0cbf96d230a83cd2309b8fd032e52d7fb933881541472df1bf2703a939  -
```

## Testnet

For further information on joining the testnet, head over to the [testnet repo](https://github.com/Kava-Labs/kava-testnets).

## License

Copyright © Kava Labs, Inc. All rights reserved.

Licensed under the [Apache v2 License](LICENSE.md).
