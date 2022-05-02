# kava-10 Upgrade Instructions

## Software Version and Key Dates

- We will be upgrading from chain-id "kava-9" to chain-id "kava_2222-10".
- The version of Kava for kava-10 is v0.17.0
- The kava-9 chain will be shutdown with a `SoftwareUpgradeProposal` that activates at block height **1412300**, which is approximately 14:00 UTC on May, 10 2022.
- kava-10 genesis time is set to May 10, 2022 at 17:00 UTC
- The version of cosmos-sdk for kava-10 is v0.45.3
- The version of tendermint for kava-10 v0.34.19
- The minimum version of golang for kava-10 is **1.17+**.

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to exporting kava-9 state, validators are encouraged to take a full data snapshot at the export height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kava directory.

It is critically important to back-up the .kava/data/priv_validator_state.json file after stopping your kava process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.16.x of the Kava software and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

### Before the upgrade

Kava Labs has submitted a `SoftwareUpgradeProposal` that specifies block height **1412300** as the final block height for kava-9. This height corresponds to approximately 14:00 UTC on March 10th. Once the proposal passes, the chain will shutdown automatically at the specified height and does not require manual intervention by validators.

### On the day of the upgrade

**The kava chain is expected to halt at block height **1412300**, at approximately 14:00 UTC, and restart with new software at 17:00 UTC March 10th. Do not stop your node and begin the upgrade before 14:00 UTC on March 10th, or you may go offline and be unable to recover until after the upgrade!**

**Make sure the kava process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

1. Export State (this **MUST** be done using **v0.16.x**)

```sh
# verify version before export:
kava version --long
# name: kava
# server_name: kava
# version: 0.16.0 (any 0.16 version is fine)
# commit: 184ef2ad4127517828a4a04cc2c51594b66ac012
# build_tags: netgo,ledger
# go: go version go1.17.1 linux/amd64

# export genesis using v0.16.x
kvd export --for-zero-height --height 1412300 > export-genesis.json
```

**Note:** This can take a while!

2. Update to kava-10

```sh
  # in the `kava` folder
  git pull
  git checkout v0.17.0
  make install

  # verify versions
  kava version --long
  # name: kava
  # server_name: kava
  # version: v0.17.0
  # commit: [TBD]
  # build_tags: netgo,ledger
  # go: go version go1.17.1 linux/amd64


  # Migrate genesis state
  kava migrate export-genesis.json > genesis.json

  # Verify output of genesis migration
  kava validate-genesis genesis.json # should say it's valid
  kava assert-invariants genesis.json # should say invariants pass
  jq -S -c -M '' genesis.json | shasum -a 256
  # [PLACEHOLDER]

  # Restart node with migrated genesis state
  cp genesis.json ~/.kava/config/genesis.json
  kava tendermint unsafe-reset-all

  # Restart node -
  # ! Be sure to remove --halt-time flag if it is set in systemd/docker
  kava start
```

### Coordination

If the kava-10 chain does not launch by March 10, 2022 at 21:00 UTC, the launch should be considered a failure and validators should refer to the [rollback](./rollback.md) instructions to restart the previous kava-9 chain. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
