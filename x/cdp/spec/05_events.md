# Events

The cdp module emits the following events:

## Handlers

### MsgCreateCDP

| Type        | Attribute Key | Attribute Value  |
|-------------|---------------|------------------|
| message     | module        | cdp              |
| message     | sender        | {sender address} |
| create_cdp  | cdp_id        | {cdp id}         |
| cdp_deposit | cdp_id        | {cdp id}         |
| cdp_deposit | amount        | {deposit amount} |
| cdp_draw    | cdp_id        | {cdp id}         |
| cdp_draw    | amount        | {draw amount}    |

### MsgWithdraw

| Type    | Attribute Key | Attribute Value  |
|---------|---------------|------------------|
| message | module        | cdp              |
| message | sender        | {sender address} |

### MsgDeposit

| Type        | Attribute Key | Attribute Value  |
|-------------|---------------|------------------|
| message     | module        | cdp              |
| message     | sender        | {sender address} |
| cdp_deposit | cdp_id        | {cdp id}         |
| cdp_deposit | amount        | {deposit amount} |

### MsgDrawDebt

| Type     | Attribute Key | Attribute Value  |
|----------|---------------|------------------|
| message  | module        | cdp              |
| message  | sender        | {sender address} |
| cdp_draw | cdp_id        | {cdp id}         |
| cdp_draw | amount        | {draw amount}    |

### MsgRepayDebt

| Type    | Attribute Key | Attribute Value  |
|---------|---------------|------------------|
| message | module        | cdp              |
| message | sender        | {sender address} |

## BeginBlock

| Type                    | Attribute Key | Attribute Value     |
|-------------------------|---------------|---------------------|
| cdp_liquidation         | module        | cdp                 |
| cdp_liquidation         | cdp_id        | {cdp id}            |
| cdp_liquidation         | depositor     | {depositor address} |
| cdp_begin_blocker_error | module        | cdp                 |
| cdp_begin_blocker_error | error_message | {error}             |
