<!--
order: 2
-->

# State

## Genesis state

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the committee module to resume.

```protobuf
message GenesisState {
  repeated Account accounts = 1;
}

```

## Account

An `Account` is a struct representing the excess `akava` balance of an address.

Since an address's total `akava` balance is derived from its `ukava` balance and the excess `akava` balance stored by the `Account` struct, the `akava` balance here should never exceed 1 `ukava` (10^12 `akava`).

```protobuf
message Account {
  bytes address = 1;
  string balance = 2;
}
```

## Store

For complete implementation details for how items are stored, see [keys.go](../types/keys.go). The `evmutil` module store state consists of accounts.
