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

Note that the above command will not stop `kvd` from running, it merely stops proposal / validation for blocks after that time.Validators may safely exit by issuing `CTRL+C` if running as a process.

Kava developers will update this PR with the final block number when it is reached. __Make sure the kvd process is stopped before proceeding and that you have backed up your validator__. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

The following up steps assume the directory structure below: change filepaths, directory names as needed

```bash
# Kvcli Folder
~/.kvcli
# Kvd Folder
~/.kvd
# Go Path
~/go
```

### Pre-Migration

1. Backup existing kava-2 .kvd and .kvcli
```sh
cp -R ~/.kvcli ~/.kvcli.bak
cp -R ~/.kvd ~/.kvd.bak
```

2. Backup existing kava-2 kvd and kvcli binaries (in case of rollback)
```sh
cp ~/go/bin/kvcli ~/go/bin/kvcli.bak
cp ~/go/bin/kvd ~/go/bin/kvd.bak
```

### Migration

We denote `(kava-2)kvd` as the previous client (0.3.5) to be used for commands e.g `(kava-2)kvd export` and `(kava-3)kvd` as the new client (0.8.1) to be used for commands.

1. Export state

- Ensure that all `kvd` processes have stopped running.

```sh
  (kava-2) kvd export --height 2598890 --for-zero-height > kava_2_exported.json
  # Check ShaSum for later reference
  $ jq -S -c -M '' kava_2_exported.json | shasum -a 256
  # Should return
  > 7b5ec6f003b3aaf0544e7490f2e383a3b0339ec4db372ce84e68992f018e20e6  -
```

2. Update to kava-3

This will replace the `kvd` and `kvcli` binaries in your GOPATH.

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
```

3. Migrate the kava-2 keys from previous key store to new key store

This will scan for any keys in `.kvcli` and produce new files ending in `kavaxxx.address` and `key_name.info` for the new keystore to access.

```sh
  # Migrate keys
  (kava-3) kvcli keys migrate
```

4. Migrate the exported genesis state

```sh
  # Migrate genesis state
  (kava-3) kvd migrate kava_2_exported.json > kava_3_migrated.json
  # Check ShaSum for later reference
  $ jq -S -c -M '' kava_3_migrated.json | shasum -a 256
  # Should return
  > a7dcd440604a150a55a33e8cd22d7d1884d06ed4668e8090f6824755f4099325  -
```

5. Write Params to genesis state and validate

```sh
  # Migrate parameters
  (kava-3) kvd write-params kava_3_migrated.json --chain-id kava-3 --genesis-time 2020-06-10T14:00:00Z > genesis.json
  # Check ShaSum for later reference
  # Note: jq must be installed
  # DO NOT WRITE THE JQ OUTPUT TO FILE. Use only for calculating the hash.
  $ jq -S -c -M '' genesis.json | shasum -a 256
  # Should return
  > f73628abfab82601c9af97a023d357a95507b9c630c5331564f48c4acab97b85  -
  # Verify output of genesis migration
  (kava-3) kvd validate-genesis genesis.json # should say it's valid
```

6. Restart node with new kava-3 genesis state

```sh
  cp genesis.json ~/.kvd/config/genesis.json
  # Unsafe Reset All is a irreversible action that wipes on-chain data and prepares the chain for a start from genesis
  # If you have not backed up your previous kava-2 state, do not proceed.
  (kava-3) kvd unsafe-reset-all
  (kava-3) kvd start
```

### Coordination

If the `kava-3` chain does not launch by June 10, 2020 at 16:00 UTC, the launch should be considered a failure. Validators should restore the state from `kava-2` and coordinate a relaunch. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
