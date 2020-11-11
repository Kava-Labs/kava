<!--
order: 4
-->

# Events

The harvest module emits the following events:

## Handlers

### MsgDeposit

| Type                 | Attribute Key       | Attribute Value       |
|----------------------|---------------------|-----------------------|
| message              | module              | harvest               |
| message              | sender              | `{sender address}`    |
| harvest_deposit      | amount              | `{amount}`            |
| harvest_deposit      | depositor           | `{depositor address}` |
| harvest_deposit      | deposit_denom       | `{deposit denom}`     |

### MsgWithdraw

| Type                   | Attribute Key       | Attribute Value       |
|------------------------|---------------------|-----------------------|
| message                | module              | harvest               |
| message                | sender              | `{sender address}`    |
| harvest_deposit        | amount              | `{amount}`            |
| harvest_deposit        | depositor           | `{depositor address}` |
| harvest_deposit        | deposit_denom       | `{deposit denom}`     |
| harvest_deposit        | deposit_type        | `{deposit type}`      |
| delete_harvest_deposit | depositor           | `{depositor address}` |
| delete_harvest_deposit | deposit_denom       | `{deposit denom}`     |

### MsgClaimReward

| Type                   | Attribute Key       | Attribute Value          |
|------------------------|---------------------|--------------------------|
| message                | module              | harvest                  |
| message                | sender              | `{sender address}`       |
| claim_harvest_reward   | amount              | `{amount}`               |
| claim_harvest_reward   | claim_holder        | `{claim holder address}` |
| claim_harvest_reward   | deposit_denom       | `{deposit denom}`        |
| claim_harvest_reward   | claim_type          | `{claim type}`         |
| claim_harvest_reward   | claim_multiplier    | `{claim multiplier}`     |

## BeginBlock

| Type                           | Attribute Key       | Attribute Value          |
|--------------------------------|---------------------|--------------------------|
| harvest_lp_distribution        | block_height        | `{block height}`         |
| harvest_lp_distribution        | rewards_distributed | `{rewards distributed}`  |
| harvest_lp_distribution        | deposit_denom       | `{deposit denom}`        |
| harvest_delegator_distribution | block_height        | `{block height}`         |
| harvest_delegator_distribution | rewards_distributed | `{rewards distributed}`  |
| harvest_delegator_distribution | deposit_denom       | `{deposit denom}`        |
