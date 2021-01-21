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

The current recommended version of the software for mainnet is [v0.12.1](https://github.com/Kava-Labs/kava/releases/tag/v0.12.1). Note, the master branch of this repository contains development work since the last mainnet release and it may **not** be runnable on mainnet.

### Installation

```bash
git checkout v0.12.1
make install
```

### Upgrade

The scheduled mainnet upgrade to `kava-4` took place on October 15th, 2020 at 14:00 UTC. The current version of Kava for `kava-4` is [__v0.12.1__](https://github.com/Kava-Labs/kava/releases/tag/v0.12.1).

The canonical genesis file can be found [here](https://github.com/Kava-Labs/launch/tree/master/kava-4)

The canonical genesis file hash is

```
jq -S -c -M '' genesis.json | shasum -a 256
# 760cd37ab07d136e5cbb8795244683f0725f63f5c69ccf61626fe735f1ed9793
```

## Testnet

For further information on joining the testnet, head over to the [testnet repo](https://github.com/Kava-Labs/kava-testnets).

## License

Copyright © Kava Labs, Inc. All rights reserved.

Licensed under the [Apache v2 License](LICENSE.md).
