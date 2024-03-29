# Kava 16 Upgrade Instructions

## Software Version and Key Dates

- The version of `kava` for Kava 16 is `v0.26.0`
- The Kava 16 chain will be shutdown with a `SoftwareUpgradeProposal` that
  activates at approximately 15:00 UTC on TBD, 2024.

## Dependency Changes

## API Changes

If you are a consumers of the legacy REST API known as the LCD, please note that this has been deprecated as a part of cosmos-sdk v47. [Additional details can be found here.](./legacy_rest.md)

### For validators using RocksDB

> [!NOTE]
> If you use goleveldb or other database backends, this is not required.

If you use RocksDB as your database backend, you will need to update RocksDB if you are using `<= v8.1.1`. The tested and recommended RocksDB version is `v8.10.0`.
Please reference the [RocksDB repository](https://github.com/facebook/rocksdb/tree/v8.10.0) to update your installation before building the RocksDB kava binary.

### On the day of the upgrade

The kava chain is expected to halt at block height **xxx**. **Do not stop your node and begin the upgrade before the upgrade height**, or you may go offline and be unable to recover until after the upgrade!

**Make sure the kava process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

**Ensure you are using golang 1.21.x and not a different version.** Golang 1.20 and below may cause app hash mismatches!

To update to v0.26.0

```sh
# check go version - look for 1.21!
go version
# go version go1.21.6 linux/amd64

# in the `kava` folder
git fetch
git checkout v0.26.0

# Note: Golang 1.21 must be installed before this step
make install

# verify versions
kava version --long
# name: kava
# server_name: kava
# version: 0.26.0
# commit: <commit placeholder>
# build_tags: netgo ledger,
# go: go version go1.21.6 linux/amd64
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

Prior to applying the Kava 16 upgrade, validators are encouraged to take a full data snapshot at the upgrade height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kava directory.

It is critically important to back-up the .kava/data/priv_validator_state.json file after stopping your kava process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.25.x of the Kava software and restore to their latest snapshot before restarting their nodes.

### Coordination

If the Kava 16 chain does not launch by TBD at 00:00 UTC, the launch should be considered a failure. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
