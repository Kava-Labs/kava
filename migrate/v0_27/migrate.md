# Kava 17 Upgrade Instructions

## Software Version and Key Dates

- The version of `kava` for Kava 17 is `v0.27.0`
- The Kava 17 chain will be shutdown with a `SoftwareUpgradeProposal` that
  activates at height 12766500 at approximately 15:00 UTC on November 21, 2024.

## Data Storage Changes

Kava 17 uses IAVL V1, an improved storage format for the low-level data storage used by the Kava blockchain.

Using IAVL V1 brings performance speedups for syncing and massively reduces the data stored on disk (~2.4x less data storage required for a full-archive node).

For full-archive node operators, an IAVL V1 snapshot will be linked when available. We are working with partners to host IAVL V1 full historical data.
For validators & operators of pruning nodes, it is recommended that node data is recreated from scratch via statesync.

Node operators using rocksdb are encouraged to [use `tcmalloc` as their memory allocator](./iavl-v1.md#default-memory-allocator).

**See the [IAVL V1 migration guide](./iavl-v1.md).**

## Dependency Changes

### For validators using RocksDB

> [!NOTE]
> If you use goleveldb or other database backends, this is not required.

If you use RocksDB as your database backend, you will need to update RocksDB if you are using `< v9.3.1`. The tested and recommended RocksDB version is `v9.3.1`.
Please reference the [RocksDB repository](https://github.com/facebook/rocksdb/tree/v9.3.1) to update your installation before building the RocksDB kava binary.

Node operators using rocksdb are encouraged to [use `tcmalloc` as their memory allocator](https://github.com/Kava-Labs/kava/blob/v0.26.2-iavl-v1/migrate/v0_26/iavl-v1.md#default-memory-allocator).

## On the day of the upgrade

The kava chain is expected to halt at block height **12766500**. **Do not stop your node and begin the upgrade before the upgrade height**, or you may go offline and be unable to recover until after the upgrade!

**Make sure the kava process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

**Ensure you are using golang 1.22.7+ and not a different version.** Golang 1.22.6 and below may cause app hash mismatches!

To update to v0.27.0

```sh
# check go version - look for 1.22.7+!
go version
# go version go1.22.9 linux/amd64

# in the `kava` folder
git fetch
git checkout v0.27.0

# Note: Golang 1.22.7+ must be installed before this step
make install

# verify versions
kava version --long
# name: kava
# server_name: kava
# version: 0.27.0
# commit: e77e0fed1cd64fdbc9f40dac6737c9e7a33cd4ae
# build_tags: netgo ledger,
# go: go version go1.22.9 linux/amd64
# build_deps:
#  ...
# cosmos_sdk_version: v0.47.10

# Restart node -
kava start
```

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to applying the Kava 17 upgrade, validators are encouraged to take a full data snapshot at the upgrade height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kava directory.

It is critically important to back-up the .kava/data/priv_validator_state.json file after stopping your kava process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.26.x of the Kava software and restore to their latest snapshot before restarting their nodes.

### Coordination

If the Kava 17 chain does not launch by November 22nd at 00:00 UTC, the launch should be considered a failure. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
