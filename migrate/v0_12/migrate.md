# Kava-5 Launch Recovery Plan

## Software Version and Key Information

* The version of Kava for kava-6 is v0.12.3
* Use the same exported genesis from kava-4 (height 1267330), we will migrate directly to kava-6
* kava-6 launch will be at 2021-03-05 at 6:00 UTC


## Procedure

1. Export State (this __MUST__ be done using __v0.12.2__, previous v0.12.x versions will not produce the same genesis hash!)

### Note this is the same as original migration plan

```sh
kvd export --for-zero-height --height 1267330 > export-genesis.json


jq -S -c -M '' export-genesis.json | shasum -a 256
6908d68987561b8e7ce646350302f64ae418014779acd84a5f3ea9a4db55bec9  -
```

__Note:__ This can take a while!

2. Update to kava-6

```sh
  # in the `kava` folder
    git pull
    git checkout v0.12.3
    make install

  # verify versions
  kvd version --long
  # name: kava
  # server_name: kvd
  # client_name: kvcli
  # version: 0.12.3
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