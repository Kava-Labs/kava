<!--
order: 4
-->

# Events

The evmutil module emits the following events:

## Handlers

### MsgConvertERC20ToCoin

| Type                  | Attribute Key | Attribute Value    |
| --------------------- | ------------- | ------------------ |
| convert_erc20_to_coin | erc20_address | `{erc20 address}`  |
| convert_erc20_to_coin | initiator     | `{initiator}`      |
| convert_erc20_to_coin | receiver      | `{receiver}`       |
| convert_erc20_to_coin | amount        | `{amount}`         |
| message               | module        | evmutil            |
| message               | sender        | {'sender address}' |

### MsgConvertCoinToERC20

| Type                  | Attribute Key | Attribute Value    |
| --------------------- | ------------- | ------------------ |
| convert_coin_to_erc20 | initiator     | `{initiator}`      |
| convert_coin_to_erc20 | receiver      | `{receiver}`       |
| convert_coin_to_erc20 | erc20_address | `{erc20_address}`  |
| convert_coin_to_erc20 | amount        | `{amount}`         |
| message               | module        | evmutil            |
| message               | sender        | {'sender address}' |
