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

### MsgWithdraw

| Type                | Attribute Key | Attribute Value       |
| ------------------- | ------------- | --------------------- |
| message             | module        | hard                  |
| message             | sender        | `{sender address}`    |
| hard_withdrawal     | amount        | `{amount}`            |
| hard_withdrawal     | depositor     | `{depositor address}` |

### MsgBorrow

| Type           | Attribute Key | Attribute Value       |
| -------------- | ------------- | --------------------- |
| message        | module        | hard                  |
| message        | sender        | `{sender address}`    |
| hard_borrow    | borrow_coins  | `{amount}`            |
| hard_withdrawal| borrower      | `{borrower address}` |

### MsgRepay

| Type           | Attribute Key | Attribute Value       |
| -------------- | ------------- | --------------------- |
| message        | module        | hard                  |
| message        | sender        | `{sender address}`    |
| hard_repay     | repay_coins   | `{amount}`            |
| hard_repay     | sender        | `{borrower address}`  |

## BeginBlock

| Type                        | Attribute Key       | Attribute Value         |
| --------------------------- | ------------------- | ----------------------- |
| hard_lp_distribution        | block_height        | `{block height}`        |
| hard_lp_distribution        | rewards_distributed | `{rewards distributed}` |
| hard_lp_distribution        | deposit_denom       | `{deposit denom}`       |
| hard_delegator_distribution | block_height        | `{block height}`        |
| hard_delegator_distribution | rewards_distributed | `{rewards distributed}` |
| hard_delegator_distribution | deposit_denom       | `{deposit denom}`       |
