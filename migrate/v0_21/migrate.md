# Kava 12 Upgrade Instructions

## Software Version and Key Dates

- The version of `kava` for Kava 12 is v0.21.0
- The Kava 11 chain will be shutdown with a `SoftwareUpgradeProposal` that activates at block height **3607200**, which is approximately 15:00 UTC on Feburary, 15th 2023.


## Upgrade Procedure

### Before the upgrade

Kava Labs has submitted a `SoftwareUpgradeProposal` that specifies block height **3607200** as the upgrade block height. This height corresponds to approximately 15:00 UTC on Feburary 15th, 2023. Once the proposal passes, the chain will shutdown automatically at the specified height and does not require manual intervention by validators.

### On the day of the upgrade

**The kava chain is expected to halt at block height **3607200**. Do not stop your node and begin the upgrade before the upgrade height, or you may go offline and be unable to recover until after the upgrade!**

**Make sure the kava process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

**Ensure you are using golang 1.18 and not a newer version.** Golang 1.19+ may cause app hash mismatches!

1. Update to v0.21.0

```sh
  # check go version - look for 1.18!
  go version
  # go version go1.18.10 linux/arm64

  # in the `kava` folder
  git fetch
  git checkout v0.21.0

  # Note: Golang 1.18 must be installed before this step
  make install

  # verify versions
  kava version --long
  # name: kava
  # server_name: kava
  # version: 0.21.0
  # commit: <commit placeholder>
  # build_tags: netgo ledger,
  # go: go version go1.18.10 linux/arm64
  # build_deps:
  #  ...
  # cosmos_sdk_version: v0.45.10

  # Restart node -
  kava start
```

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to applying the Kava 12 upgrade, validators are encouraged to take a full data snapshot at the upgrade height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kava directory.

It is critically important to back-up the .kava/data/priv_validator_state.json file after stopping your kava process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.19.x of the Kava software and restore to their latest snapshot before restarting their nodes.

### Coordination

If the kava 12 chain does not launch by Feburary 16th, 2023 at 00:00 UTC, the launch should be considered a failure. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
