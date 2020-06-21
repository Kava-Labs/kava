<!--
order: 4
-->

# Events

The `x/issuance` module emits the following events:

## BeginBlock

| Type                 | Attribute Key       | Attribute Value |
|----------------------|---------------------|-----------------|
| issue_tokens         | amount_issued       | `{amount}`      |
| redeem_tokens        | amount_redeemed     | `{amount}`      |
| block_address        | address_blocked     | `{address}`     |
| block_address        | denom               | `{denom}`       |
| change_pause_status  | pause_status        | `{bool}`        |
| change_pause_status  | denom               | `{denom}`       |