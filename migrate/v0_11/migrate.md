# Kava-4 Upgrade Instructions

## Software Version and Key Dates

* The version of Kava for kava-4 is v0.11
* Kava-3 validators should prepare to shutdown their nodes October 15th, 2020 at 12:00 UTC
* Kava-4 genesis time is set to October 15th, 2020 at 14:00 UTC
* The version of cosmos-sdk for kava-4 is v0.39.1
* The version of tendermint for kava-4 v0.33.7

### Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.

### Recovery

Prior to exporting kava-3 state, validators are encouraged to take a full data snapshot at the export height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kvd and .kvcli directories.

It is critically important to back-up the .kvd/data/priv_validator_state.json file after stopping your kvd process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.10.0 of the Kava software and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

### Before the upgrade

Set your node to produce the final block of kava-3 at __12:00__ UTC October 15th, 2020. To restart your node with that stop time,

```sh
kvd start --halt-time 1602763200
```

You can safely set the halt-time flag at any time.

### On the day of the upgrade

__The kava chain is expected to halt at 12:00 UTC, and restart with new software at 14:00 UTC October 15th. Do not stop your node and begin the upgrade before 12:00UTC on October 15th, or you may go offline and be unable to recover until after the upgrade!__

Kava developers will update this PR with the final block number when it is reached. __Make sure the kvd process is stopped before proceeding and that you have backed up your validator__. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.

1. Export State

```sh
kvd export --for-zero-height > export-genesis.json
```

2. Update to kava-4

```sh
  # in the `kava` folder
    git pull
    git checkout v0.11.0
    make install

  # verify versions
  kvd version --long
  # name: kava
  # server_name: kvd
  # client_name: kvcli
  # version: 0.11.0
  # commit: TODO Placeholder
  # build_tags: netgo,ledger
  # go: go version go1.15.2 linux/amd64


  # Migrate genesis state
  kvd migrate export-genesis.json > genesis.json

  # Verify output of genesis migration
  kvd validate-genesis genesis.json # should say it's valid
  jq -S -c -M '' genesis.json | shasum -a 256
  # [PLACEHOLDER]

  # Restart node with migrated genesis state
  cp genesis.json ~/.kvd/config/genesis.json
  kvd unsafe-reset-all
  kvd start
```

### Coordination

If the kava-4 chain does not launch by October 15, 2020 at 16:00 UTC, the launch should be considered a failure. Validators should restore the state from kava-3 and coordinate a relaunch. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).
