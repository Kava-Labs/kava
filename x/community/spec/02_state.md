<!--
order: 2
-->

# State

## Parameters and Genesis State

`Params` define the module parameters, containing the information required to
set the current staking rewards per second at a future date. When the 
`upgrade_time_disable_inflation` time is reached, `staking_rewards_per_second`
will be set to `upgrade_time_set_staking_rewards_per_second`.

```protobuf
// Params defines the parameters of the community module.
message Params {
  option (gogoproto.equal) = true;

  // upgrade_time_disable_inflation is the time at which to disable mint and kavadist module inflation.
  // If set to 0, inflation will be disabled from block 1.
  google.protobuf.Timestamp upgrade_time_disable_inflation = 1 [
    (gogoproto.stdtime) = true,
    (gogoproto.nullable) = false
  ];

  // staking_rewards_per_second is the amount paid out to delegators each block from the community account
  string staking_rewards_per_second = 2 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // upgrade_time_set_staking_rewards_per_second is the initial staking_rewards_per_second to set
  // and use when the disable inflation time is reached
  string upgrade_time_set_staking_rewards_per_second = 3 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];
}
```

`GenesisState` defines the state that must be persisted when the blockchain
stops/restarts in order for normal function of the module to resume. It contains
the parameters and staking rewards state to keep track of payout between blocks.

```protobuf
// GenesisState defines the community module's genesis state.
message GenesisState {
  // params defines all the parameters related to commmunity
  Params params = 1 [(gogoproto.nullable) = false];

  // StakingRewardsState stores the internal staking reward data required to
  // track staking rewards across blocks
  StakingRewardsState staking_rewards_state = 2 [(gogoproto.nullable) = false];
}

// StakingRewardsState represents the state of staking reward accumulation between blocks.
message StakingRewardsState {
  // last_accumulation_time represents the last block time which rewards where calculated and distributed.
  // This may be zero to signal accumulation should start on the next interval.
  google.protobuf.Timestamp last_accumulation_time = 1 [
    (gogoproto.stdtime) = true,
    (gogoproto.nullable) = false
  ];

  // accumulated_truncation_error represents the sum of previous errors due to truncation on payout
  // This value will always be on the interval [0, 1).
  string last_truncation_error = 2 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];
}
```
