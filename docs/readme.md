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

The scheduled mainnet upgrade to `kava-9` took place on January 19th, 2022 at 15:00 UTC. The current version of Kava for `kava-9` is [__v0.16.1__](https://github.com/Kava-Labs/kava/releases/tag/v0.16.1).

The canonical genesis file can be found [here](https://github.com/Kava-Labs/launch/tree/master/kava-9)

The canonical genesis file hash is

```
jq -S -c -M '' genesis.json | shasum -a 256
5c688df5ae6cba9c9e5a9bab045eb367dd54ce9b7f5fab78cf3e636cf2e2b793  -
```

For additional information about upgrades:

- Upgrading a [Validator](https://github.com/Kava-Labs/kava/blob/master/docs/kava-8/upgrade_validator.md)
- Upgrading a [Wallet](https://github.com/Kava-Labs/kava/blob/master/docs//kava-8/upgrade_wallet.md)
- Upgrading an [Oracle](https://github.com/Kava-Labs/kava/blob/master/docs//kava-8/upgrade_oracle.md)

### Validator Guide

Use this [guide](https://github.com/Kava-Labs/kava/blob/master/docs/validator_guide.md) to set up a mainnet validator. Note that this is a minimal guide and does not cover more advanced topics like [sentry node architecture](https://github.com/stakefish/cosmos-validator-design) and [double signing protection](https://github.com/tendermint/tmkms). It is strongly recommended that any parties considering validating do additional research.  If you have questions, please join the active conversation in the #validators thread of the [__Kava Discord Channel__](https://discord.com/invite/kQzh3Uv).

## Testnet

For further information on joining the testnet, head over to the [testnet repo](https://github.com/Kava-Labs/kava-testnets).

## License

Copyright © Kava Labs, Inc. All rights reserved.

Licensed under the [Apache v2 License](LICENSE.md).
