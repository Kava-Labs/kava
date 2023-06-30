# Kava 14 Upgrade Instructions

## Software Version and Key Dates

- The version of `kava` for Kava 14 is v0.24.0
- The Kava 13 chain will be shutdown with a `SoftwareUpgradeProposal` that activates at approximately 15:00 UTC on July, 12th 2023.

## Configuration Changes


**For validators with existing configurations, it is recommended to set `evm.max-tx-gas-wanted = 0` in app.toml to avoid proposing blocks that exceed the block gas limit.**

In previous versions, the default was non-zero and was used to mitigate DDoS style gas attacks.  However, this setting is not required anymore and can safely be set to zero.


### On the day of the upgrade

**The kava chain is expected to halt at block height **5597000**. Do not stop your node and begin the upgrade before the upgrade height, or you may go offline and be unable to recover until after the upgrade!**

**Make sure the kava process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

**Ensure you are using golang 1.20.x and not aa different version.** Golang 1.19 and below may cause app hash mismatches!

1. Update to v0.24.0

```sh
  # check go version - look for 1.20!
  go version
  # go version go1.20.5 linux/arm64

  # in the `kava` folder
  git fetch
  git checkout v0.24.0

  # Note: Golang 1.20 must be installed before this step
  make install

  # verify versions
  kava version --long
  # name: kava
  # server_name: kava
  # version: 0.24.0
  # commit: <commit placeholder>
  # build_tags: netgo ledger,
  # go: go version go1.20.5 linux/arm64
  # build_deps:
  #  ...
  # cosmos_sdk_version: v0.46.11

  # Restart node -
  kava start
```

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to applying the Kava 14 upgrade, validators are encouraged to take a full data snapshot at the upgrade height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kava directory.

It is critically important to back-up the .kava/data/priv_validator_state.json file after stopping your kava process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.23.x of the Kava software and restore to their latest snapshot before restarting their nodes.

### Coordination

If the Kava 14 chain does not launch by July 13th, 2023 at 00:00 UTC, the launch should be considered a failure. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
