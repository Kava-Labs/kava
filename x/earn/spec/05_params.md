# Parameters

The earn module contains the following parameters:

| Key           | Type           | Example          | Description            |
| ------------- | -------------- | ---------------- | ---------------------- |
| AllowedVaults | []AllowedVault | []AllowedVault{} | List of allowed vaults |

## AllowedVault

| Key           | Type         | Example              | Description                     |
| ------------- | ------------ | -------------------- | ------------------------------- |
| Denom         | string       | `ukava`              | The denom of the allowed vault  |
| VaultStrategy | StrategyType | `STRATEGY_TYPE_LEND` | The strategy type for the vault |
