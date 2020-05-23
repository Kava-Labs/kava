# Upgrade to kava-3

1. `kvd export --for-zero-height > export-genesis.json`
1. get new kava version:
    1. in kava folder: `git pull`, `git checkout v0.8.0`, `make install`
    1. check versions with `kvd version --long` and `kvcli version -long`
1. `kvcli keys migrate`
1. `kvd migrate export-genesis.json > migrated-genesis.json`
1. `kvd write-params migrated-genesis.json --chain-id kava-3 --genesis-time 2020-06-01T14:00:00Z > genesis.json`
1. check genesis file:
    1. `kvd validate-genesis genesis.json` should say it's valid
    1. `shasum -a 256 genesis.json` should output _
1. `cp genesis.json ~/.kvd/config/genesis.json`
1. `kvd unsafe-reset-all`
1. `kvd start`
