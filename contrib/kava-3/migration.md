# Kava-3 Upgrade Instructions

## Background

The version of Kava for kava-3 is __v0.8.x__. Kava-3 is scheduled to launch __June 10, 2020 at 14:00 UTC__

🚨 Please note that an issue with tendermint v0.33 has been found that affects the stability of nodes running with the default pruning strategy. Please see the [Pruning](#Pruning) section for full details and mitigation 🚨

Many changes have occurred in both the Kava software and the cosmos-sdk software since the launch of kava-2. The primary changes in Kava are the addition of modules that comprise the [CDP system](https://docs.kava.io/). To review cosmos-sdk changes, see the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.38.4/CHANGELOG.md) and note that kava-3 is launching with __v0.38.4__ of the cosmos-sdk.

If you have technical questions or concerns, ask a developer or community member in the [Kava discord](https://discord.com/invite/kQzh3Uv).

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Pruning

kava-3 uses tendermint version 0.33. Recent testing in Game of Zones and Kava has shown that nodes which are running with the default or custom pruning strategy have a [memory leak](https://github.com/tendermint/iavl/issues/256) that can cause nodes to crash and lead to irrecoverable data loss. Until a patch is released, the __ONLY__ pruning strategies that are safe to run are `nothing` (an archival node, where nothing is deleted) or `everything` (only the most recent state is kept).

The pruning config is set in $HOME/.kvd/config/app.toml. Example safe configurations are:

```toml
pruning = "nothing"
```

and

```toml
pruning = "everything"
```

Exchange operators, data service providers, and other vendors who require access to historical state are recommended to run archival nodes (`pruning = "nothing"`). Other node operators can choose between a fully pruning node and archival node, with the main difference being increased storage required for archival nodes.

It is expected that a patch to tendermint will be released in a non-breaking manner and that nodes will be able to update seamlessly after the launch of kava-3.

### Recovery

Prior to exporting kava-2 state, validators are encouraged to take a full data snapshot at the export height before proceeding. Snapshotting depends heavily on infrastructure, but generally this can be done by backing up the .kvd and .kvcli directories.

It is critically important to back-up the .kvd/data/priv_validator_state.json file after stopping your kvd process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.3.5 of the Kava software and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

Set your node to produce the final block of kava-2 at __13:00__ UTC June 10th, 2020. To restart your node with that stop time,

```sh
kvd start --halt-time 1591794000
```

 Kava developers will update this PR with the final block number when it is reached. __Make sure the kvd process is stopped before proceeding and that you have backed up your validator__. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

1. Export state

```sh
  kvd export --for-zero-height > export-genesis.json
```

2. Update to kava-3

```sh
  # in the `kava` folder
  git pull
  git checkout v0.8.1
  make install

  # verify versions
  kvd version --long
  # name: kava
  # server_name: kvd
  # client_name: kvcli
  # version: 0.8.1
  # commit: 869189054d68d6ec3e6446156ea0a91eb45af09c
  # build_tags: netgo,ledger
  # go: go version go1.13.7 linux/amd64

  # Migrate keys
  kvcli keys migrate

  # Migrate genesis state
  kvd migrate export-genesis.json > migrated-genesis.json

  # Migrate parameters
  kvd write-params migrated-genesis.json --chain-id kava-3 --genesis-time 2020-06-10T14:00:00Z > genesis.json

  # Verify output of genesis migration
  kvd validate-genesis genesis.json # should say it's valid
  jq -S -c -M '' genesis.json | shasum -a 256 # note: jq must be installed
  # [PLACEHOLDER]

  # Restart node with migrated genesis state
  cp genesis.json ~/.kvd/config/genesis.json
  kvd unsafe-reset-all
  kvd start
```

### Coordination

If the `kava-3` chain does not launch by June 10, 2020 at 16:00 UTC, the launch should be considered a failure. Validators should restore the state from `kava-2` and coordinate a relaunch. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
