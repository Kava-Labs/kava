# Parameters

The earn module contains the following parameters:

| Key           | Type             | Example            | Description            |
| ------------- | ---------------- | ------------------ | ---------------------- |
| AllowedVaults | `[]AllowedVault` | `[]AllowedVault{}` | List of allowed vaults |

## AllowedVault

| Key               | Type               | Example              | Description                                                                   |
| ----------------- | ------------------ | -------------------- | ----------------------------------------------------------------------------- |
| Denom             | `string`           | `ukava`              | The denom of the allowed vault                                                |
| Strategies        | `[]StrategyType`   | `STRATEGY_TYPE_LEND` | The strategy types for the vault                                              |
| IsPrivateVault    | `bool`             | `false`              | If the vault is private and should only allow deposits from AllowedDepositors |
| AllowedDepositors | `[]sdk.AccAddress` | `[]sdk.AccAddress{}` | Slice of addresses that are allowed to deposit                                |
