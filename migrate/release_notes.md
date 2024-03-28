# Kava 16 Release Candidate 1 (v0.26.0-alpha.0)

For deployment to Kava Mainnet at height `xxxx` around 2024-04-17 15:00:00 UTC.

[Click to view upgrade instructions](https://github.com/Kava-Labs/kava/blob/v0.26.0/migrate/v0_26/migrate.md)

| Software                       | Version  |
| ------------------------------ | -------- |
| Golang                         | v1.21    |
| Cosmos SDK                     | v0.47.10 |
| CometBFT (formerly Tendermint) | v0.37.4  |
| Rocksdb                        | v8.10.0+ |

Please note that if you run your node with `rocksdb` as the database, this update will require an update to `v8.10.0` of `rocksdb`.

## Upgrade Changes

For a complete list of changes, see [CHANGELOG.md](https://github.com/Kava-Labs/kava/blob/v0.26.0/CHANGELOG.md#v0260).

### Cosmos SDK Updated to v0.47.10

- Updated to cosmos-sdk v0.47.10
- Removed support for legacy REST API on all kava modules
- Updated `x/incentive` cli to use grpc query client instead of legacy REST API
- Added grpc query service to `x/validator-vesting` to replace legacy REST API

### BEP3 EVM Native Conversion

- Update EVM native conversion logic to handle bep3 assets

### Packet Forwarding Middleware

- Added ibc packet forwarding middleware for ibc transfers

### CDP Performance Improvements

- Add module param and logic for running `x/cdp` begin blocker every n blocks
