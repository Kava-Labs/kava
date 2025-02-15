# Kava 18 Upgrade Instructions

## Software Version and Key Dates

- The version of `kava` for Kava 18 is `v0.28.0`
- The Kava 18 chain will be shutdown with a `SoftwareUpgradeProposal` that
  activates at height 14136737 at approximately 15:00 UTC on February 25, 2025.

## Dependency Changes

packet-forwarding-middleware is updated to resolve unreliability of USDT IBC
transactions.

### For validators using RocksDB

> [!NOTE]
> If you use goleveldb or other database backends, this is not required.

There are no changes in rocksDB version requirements for this upgrade.

Same as the previous version, the recommended RocksDB version is `v9.3.1`.

Please reference the [RocksDB repository](https://github.com/facebook/rocksdb/tree/v9.3.1) to update your installation before building the RocksDB kava binary.

Node operators using rocksdb are encouraged to [use `tcmalloc` as their memory allocator](https://github.com/Kava-Labs/kava/blob/v0.26.2-iavl-v1/migrate/v0_26/iavl-v1.md#default-memory-allocator).

## On the day of the upgrade

The kava chain is expected to halt at block height **14136737**.
**Do not stop your node and begin the upgrade before the upgrade height**,
or you may go offline and be unable to recover until after the upgrade!

**Make sure the kava process is stopped before proceeding and that you have backed up your validator**.
Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

**Ensure you are using golang 1.22.7+ and not a different version.** Golang 1.22.6 and below may cause app hash mismatches!

To update to v0.28.0

```sh
# check go version - look for 1.22.7+!
go version
# go version go1.22.9 linux/amd64

# in the `kava` folder
git fetch
git checkout v0.28.0

# Note: Golang 1.22.7+ must be installed before this step
make install

# verify versions
kava version --long
# build_deps:
#  ...
# build_tags: netgo ledger,
# commit: xxx
# cosmos_sdk_version: v0.47.15
# go: go version go1.22.9 linux/amd64
# name: kava
# server_name: kava
# version: 0.28.0

# Restart node -
kava start
```

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries
a heightened risk of double-signing and being slashed. The most important piece
of this procedure is verifying your software version and genesis file hash
before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and
repeat the upgrade procedure again during the network startup. If you discover a
mistake in the process, the best thing to do is wait for the network to start
before correcting it. If the network is halted and you have started with a
different genesis file than the expected one, seek advice from a Kava developer
before resetting your validator.

### Recovery

Prior to applying the Kava 18 upgrade, validators are encouraged to take a full
data snapshot at the upgrade height before proceeding. Snap-shotting depends
heavily on infrastructure, but generally this can be done by backing up the
.kava directory.

It is critically important to back-up the .kava/data/priv_validator_state.json
file after stopping your kava process. This file is updated every block as your
validator participates in consensus rounds. It is a critical file needed to
prevent double-signing, in case the upgrade fails and the previous chain needs
to be restarted.

In the event that the upgrade does not succeed, validators and operators must
downgrade back to v0.27.x of the Kava software and restore to their latest
snapshot before restarting their nodes.

### Coordination

If the Kava 18 chain does not launch by February 26th at 00:00 UTC, the launch
should be considered a failure. In the event of launch failure, coordination
will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
