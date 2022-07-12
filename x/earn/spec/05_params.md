# Parameters

The earn module contains the following parameters:

| Key    | Type    | Example | Description            |
| ------ | ------- | ------- | ---------------------- |
| Vaults | []Vault | true    | List of enabled vaults |

## Vault

| Key      | Type   | Example                                      | Description               |
| -------- | ------ | -------------------------------------------- | ------------------------- |
| Address  | bytes  | `0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2` | ERC20 address on Ethereum |
| Name     | string | `"Wrapped Ether"`                            | ERC20 token name          |
| Symbol   | string | `"WETH"`                                     | ERC20 token symbol        |
| Decimals | uint8  | `18`                                         | ERC20 token decimals      |

Governance param change proposals are used to add new Ethereum ERC20s to the
enabled list. Ethereum ERC20s that are not in the list are rejected from
being bridged to Kava.

## ConversionPair

| Key          | Type   | Example                                      | Description        |
| ------------ | ------ | -------------------------------------------- | ------------------ |
| ERC20Address | bytes  | `0xfcda0a4073b927e06432c999d6cc9975d3cd3403` | Kava ERC20 address |
| Denom        | string | `"weth"`                                     | Coin Denom         |
