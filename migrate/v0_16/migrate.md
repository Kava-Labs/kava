# kava-9 Upgrade Instructions

## Software Version and Key Dates

- We will be upgrading from chain-id "kava-8" to chain-id "kava-9".
- The version of Kava for kava-9 is v0.16.0
- The kava-8 chain will be shutdown with a `SoftwareUpgradeProposal` that activates at block height __1803250__, which is approximately 14:00 UTC on January, 19 2022.  
- kava-9 genesis time is set to January 19, 2022 at 16:00 UTC
- The version of cosmos-sdk for kava-9 is v0.44.5
- The version of tendermint for kava-9 v0.34.14
- The minimum version of golang for kava-9 is __1.16+__.

__NOTE__: As part of the upgrade to kava-9, the `kvd` and `kvcli` binaries were combined into a single blockchain binary named `kava`. When restarting the chain, be sure to use `kava start` and not the deprecated `kvd start`. 

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to exporting kava-8 state, validators are encouraged to take a full data snapshot at the export height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kvd and .kvcli directories.

It is critically important to back-up the .kvd/data/priv_validator_state.json file after stopping your kvd process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.15.2 of the Kava software and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

### Before the upgrade

Kava Labs has submitted a `SoftwareUpgradeProposal` that specifies block height __1803250__ as the final block height for kava-8. This height corresponds to approximately 14:00 UTC on January 19th. Once the proposal passes, the chain will shutdown automatically at the specified height and does not require manual intervention by validators. 

### On the day of the upgrade

**The kava chain is expected to halt at block height __1803250__, at approximately 14:00 UTC, and restart with new software at 16:00 UTC January 19th. Do not stop your node and begin the upgrade before 14:00UTC on January 19th, or you may go offline and be unable to recover until after the upgrade!**

**Make sure the kvd process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

1. Export State (this **MUST** be done using **v0.15.x**)

```sh
# verify version before export: 
kvd version --long
# name: kava
# server_name: kvd
# client_name: kvcli
# version: 0.15.0 (any 0.15 version is fine)
# commit: 8691ac44ed0e65db7ebc4a2fe85c58c717f63c39
# build_tags: netgo,ledger
# go: go version go1.17.1 linux/amd64

# export genesis using v0.15.x
kvd export --for-zero-height --height 1803250 > export-genesis.json
```

**Note:** This can take a while!

2. Update to kava-9

```sh
  # in the `kava` folder
  git pull
  git checkout v0.16.0
  make install

  # verify versions
  kvd version --long
  # name: kava
  # server_name: kvd
  # client_name: kvcli
  # version: v0.16.0
  # commit: [PLACEHOLDER]
  # build_tags: netgo,ledger
  # go: go version go1.17.1 linux/amd64


  # Migrate genesis state
  kvd migrate export-genesis.json > genesis.json

  # Verify output of genesis migration
  kvd validate-genesis genesis.json # should say it's valid
  kvd assert-invariants genesis.json # should say invariants pass
  jq -S -c -M '' genesis.json | shasum -a 256
  # [PLACEHOLDER]

  # Restart node with migrated genesis state
  cp genesis.json ~/.kvd/config/genesis.json
  kvd unsafe-reset-all

  # Restart node -
  # ! Be sure to remove --halt-time flag if it is set in systemd/docker
  # NOTE: THE BINARY IS NOW NAMED KAVA
  kava start
```

### Coordination

If the kava-9 chain does not launch by January 19, 2022 at 20:00 UTC, the launch should be considered a failure and validators should refer to the [rollback](./rollback.md) instructions to restart the previous kava-8 chain. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
