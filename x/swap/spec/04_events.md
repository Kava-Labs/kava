<!--
order: 4
-->

# Events

The swap module emits the following events:

## Handlers

### MsgDeposit

| Type         | Attribute Key | Attribute Value       |
| ------------ | ------------- | --------------------- |
| message      | module        | swap                  |
| message      | sender        | `{sender address}`    |
| swap_deposit | pool_id       | `{poolID}`            |
| swap_deposit | depositor     | `{depositor address}` |
| swap_deposit | amount        | `{amount}`            |
| swap_deposit | shares        | `{shares}`            |

### MsgWithdraw

| Type          | Attribute Key | Attribute Value       |
| ------------- | ------------- | --------------------- |
| message       | module        | swap                  |
| message       | sender        | `{sender address}`    |
| swap_withdraw | pool_id       | `{poolID}`            |
| swap_withdraw | owner         | `{owner address}`     |
| swap_withdraw | amount        | `{amount}`            |
| swap_withdraw | shares        | `{shares}`            |


### MsgSwapExactForTokens

| Type          | Attribute Key | Attribute Value          |
| ------------- | ------------- | ------------------------ |
| message       | module        | swap                     |
| message       | sender        | `{sender address}`       |
| swap_trade    | pool_id       | `{poolID}`               |
| swap_trade    | requester     | `{requester address}`    |
| swap_trade    | swap_input    | `{input amount}`         |
| swap_trade    | swap_output   | `{output amount}`        |
| swap_trade    | fee_paid      | `{fee amount}`           |
| swap_trade    | exact         | `{exact trade direction}`|


### MsgSwapForExactTokens

| Type          | Attribute Key | Attribute Value          |
| ------------- | ------------- | ------------------------ |
| message       | module        | swap                     |
| message       | sender        | `{sender address}`       |
| swap_trade    | pool_id       | `{poolID}`               |
| swap_trade    | requester     | `{requester address}`    |
| swap_trade    | swap_input    | `{input amount}`         |
| swap_trade    | swap_output   | `{output amount}`        |
| swap_trade    | fee_paid      | `{fee amount}`           |
| swap_trade    | exact         | `{exact trade direction}`|
