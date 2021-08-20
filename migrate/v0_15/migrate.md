# kava-8 Upgrade Instructions

## Software Version and Key Dates

- We will be upgrading from chain-id "kava-7" to chain-id "kava-8".
- The version of Kava for kava-8 is v0.15.0
- The kava-7 chain will be shutdown with a `SoftwareUpgradeProposal` that activates at block height __1878509__, which is approximately 13:00 UTC on August, 30 2021.  
- kava-8 genesis time is set to August 30th, 2021 at 15:00 UTC
- The version of cosmos-sdk for kava-8 is v0.39.3
- The version of tendermint for kava-8 v0.33.9
- The minimum version of golang for kava-8 is 1.13+, 1.15+ has been tested and is recommended.

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to exporting kava-7 state, validators are encouraged to take a full data snapshot at the export height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kvd and .kvcli directories.

It is critically important to back-up the .kvd/data/priv_validator_state.json file after stopping your kvd process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.14.3 of the Kava software and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

### Before the upgrade

Kava Labs has submitted a `SoftwareUpgradeProposal` that specifies block height __1878509__ as the final block height for kava-7. This height corresponds to approximately 13:00 UTC on August 30th. Once the proposal passes, the chain will shutdown automatically at the specified height and does not require manual intervention by validators. 

### On the day of the upgrade

**The kava chain is expected to halt at block height 1878509, at approximately 13:00 UTC, and restart with new software at 15:00 UTC August 30th. Do not stop your node and begin the upgrade before 13:00UTC on August 30th, or you may go offline and be unable to recover until after the upgrade!**

**Make sure the kvd process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

1. Export State (this **MUST** be done using **v0.14.3**, previous v0.14.x versions will not produce the same genesis hash!)

```sh
# verify version before export: 
kvd version --long
# name: kava
# server_name: kvd
# client_name: kvcli
# version: 0.14.3
# commit: 19c4926078be89730b336edf5bb3873f3c396e43
# build_tags: netgo,ledger
# go: go version go1.15.8 linux/amd64

# export genesis using v0.14.3
kvd export --for-zero-height --height 1878509 > export-genesis.json
```

**Note:** This can take a while!

2. Update to kava-8

```sh
  # in the `kava` folder
  git pull
  git checkout v0.15.0
  make install

  # verify versions
  kvd version --long
  # name: kava
  # server_name: kvd
  # client_name: kvcli
  # version: v0.15.0
  # commit: PLACEHOLDER
  # build_tags: netgo,ledger
  # go: go version go1.15.8 linux/amd64


  # Migrate genesis state
  kvd migrate export-genesis.json > genesis.json

  # Verify output of genesis migration
  kvd validate-genesis genesis.json # should say it's valid
  jq -S -c -M '' genesis.json | shasum -a 256
  # PLACEHOLDER

  # Restart node with migrated genesis state
  cp genesis.json ~/.kvd/config/genesis.json
  kvd unsafe-reset-all

  # Restart node -
  # ! Be sure to remove --halt-time flag if it is set in systemd/docker
  kvd start
```

### Coordination

If the kava-8 chain does not launch by August 30, 2021 at 19:00 UTC, the launch should be considered a failure and validators should refer to the [rollback](./rollback.md) instructions to restart the previous kava-7 chain. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
