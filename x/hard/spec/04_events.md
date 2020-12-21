<!--
order: 4
-->

# Events

The hard module emits the following events:

## Handlers

### MsgDeposit

| Type         | Attribute Key | Attribute Value       |
| ------------ | ------------- | --------------------- |
| message      | module        | hard                  |
| message      | sender        | `{sender address}`    |
| hard_deposit | amount        | `{amount}`            |
| hard_deposit | depositor     | `{depositor address}` |
| hard_deposit | deposit_denom | `{deposit denom}`     |

### MsgWithdraw

| Type                | Attribute Key | Attribute Value       |
| ------------------- | ------------- | --------------------- |
| message             | module        | hard                  |
| message             | sender        | `{sender address}`    |
| hard_deposit        | amount        | `{amount}`            |
| hard_deposit        | depositor     | `{depositor address}` |
| hard_deposit        | deposit_denom | `{deposit denom}`     |
| hard_deposit        | deposit_type  | `{deposit type}`      |
| delete_hard_deposit | depositor     | `{depositor address}` |
| delete_hard_deposit | deposit_denom | `{deposit denom}`     |

### MsgClaimReward

| Type              | Attribute Key    | Attribute Value          |
| ----------------- | ---------------- | ------------------------ |
| message           | module           | hard                     |
| message           | sender           | `{sender address}`       |
| claim_hard_reward | amount           | `{amount}`               |
| claim_hard_reward | claim_holder     | `{claim holder address}` |
| claim_hard_reward | deposit_denom    | `{deposit denom}`        |
| claim_hard_reward | claim_type       | `{claim type}`           |
| claim_hard_reward | claim_multiplier | `{claim multiplier}`     |

## BeginBlock

| Type                        | Attribute Key       | Attribute Value         |
| --------------------------- | ------------------- | ----------------------- |
| hard_lp_distribution        | block_height        | `{block height}`        |
| hard_lp_distribution        | rewards_distributed | `{rewards distributed}` |
| hard_lp_distribution        | deposit_denom       | `{deposit denom}`       |
| hard_delegator_distribution | block_height        | `{block height}`        |
| hard_delegator_distribution | rewards_distributed | `{rewards distributed}` |
| hard_delegator_distribution | deposit_denom       | `{deposit denom}`       |
