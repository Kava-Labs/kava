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

The current recommended version of the software for mainnet is [v0.14.1](https://github.com/Kava-Labs/kava/releases/tag/v0.14.1). The master branch of this repository often contains considerable development work since the last mainnet release and is __not__ runnable on mainnet.

### Installation

```bash
git checkout v0.14.1
make install
```

### Upgrade

The scheduled mainnet upgrade to `kava-7` took place on April 8th, 2021 at 15:00 UTC. The current version of Kava for `kava-7` is [__v0.14.1__](https://github.com/Kava-Labs/kava/releases/tag/v0.14.1).

The canonical genesis file can be found [here](https://github.com/Kava-Labs/launch/tree/master/kava-4)

The canonical genesis file hash is

```
jq -S -c -M '' genesis.json | shasum -a 256
9dbff5a0fb1a7aa20247f73e974bfd4a11090252768869ef8ccb23a515a01c51  -
```

## Testnet

For further information on joining the testnet, head over to the [testnet repo](https://github.com/Kava-Labs/kava-testnets).

## License

Copyright © Kava Labs, Inc. All rights reserved.

Licensed under the [Apache v2 License](LICENSE.md).
