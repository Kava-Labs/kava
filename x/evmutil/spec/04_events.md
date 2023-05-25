<!--
order: 4
-->

# Events

The evmutil module emits the following events:

## Handlers

### MsgConvertERC20ToCoin

| Type                      | Attribute Key | Attribute Value    |
| ------------------------- | ------------- | ------------------ |
| convert_evm_erc20_to_coin | initiator     | `{initiator}`      |
| convert_evm_erc20_to_coin | receiver      | `{receiver}`       |
| convert_evm_erc20_to_coin | erc20_address | `{erc20_address}`  |
| convert_evm_erc20_to_coin | amount        | `{amount}`         |
| message                   | module        | evmutil            |
| message                   | sender        | {'sender address'} |

### MsgConvertCoinToERC20

| Type                        | Attribute Key | Attribute Value    |
| --------------------------- | ------------- | ------------------ |
| convert_evm_erc20_from_coin | initiator     | `{initiator}`      |
| convert_evm_erc20_from_coin | receiver      | `{receiver}`       |
| convert_evm_erc20_from_coin | erc20_address | `{erc20_address}`  |
| convert_evm_erc20_from_coin | amount        | `{amount}`         |
| message                     | module        | evmutil            |
| message                     | sender        | {'sender address'} |

### MsgConvertCosmosCoinToERC20

| Type                         | Attribute Key | Attribute Value    |
| ---------------------------- | ------------- | ------------------ |
| convert_cosmos_coin_to_erc20 | initiator     | `{initiator}`      |
| convert_cosmos_coin_to_erc20 | receiver      | `{receiver}`       |
| convert_cosmos_coin_to_erc20 | erc20_address | `{erc20_address}`  |
| convert_cosmos_coin_to_erc20 | amount        | `{amount}`         |
| message                      | module        | evmutil            |
| message                      | sender        | {'sender address'} |
