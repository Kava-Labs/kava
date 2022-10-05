# kava-11 Upgrade Instructions

## Software Version and Key Dates

- The version of Kava for kava-11 is v0.19.0
- The kava-10 chain will be shutdown with a `SoftwareUpgradeProposal` that activates at block height **1907500**, which is approximately 15:00 UTC on October, 12 2022.


## Upgrade Procedure

### Before the upgrade

Kava Labs has submitted a `SoftwareUpgradeProposal` that specifies block height **1907500** as the upgrade block height. This height corresponds to approximately 15:00 UTC on October 12th, 2022. Once the proposal passes, the chain will shutdown automatically at the specified height and does not require manual intervention by validators.

### On the day of the upgrade

**The kava chain is expected to halt at block height **1907500**. Do not stop your node and begin the upgrade before the upgrade height, or you may go offline and be unable to recover until after the upgrade!**

**Make sure the kava process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

1. Update to v0.19.0

```sh
  # in the `kava` folder
  git pull
  git checkout v0.19.0
  make install

  # verify versions
  kava version --long
  # name: kava
  # server_name: kava
  # version: v0.19.0
  # commit: [TBD]
  # build_tags: netgo,ledger
  # go: go version go1.17.1 linux/amd64



  # Restart node -
  # ! Be sure to remove --halt-time flag if it is set in systemd/docker
  kava start
```

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to applying the kava-11 upgrade, validators are encouraged to take a full data snapshot at the upgrade height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kava directory.

It is critically important to back-up the .kava/data/priv_validator_state.json file after stopping your kava process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.18.x of the Kava software and restore to their latest snapshot before restarting their nodes.

### Coordination

If the kava-11 chain does not launch by October 13, 2022 at 00:00 UTC, the launch should be considered a failure and validators should refer to the [rollback](./rollback.md) instructions to restart the previous kava-9 chain. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).