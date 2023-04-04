<!--
order: 5
-->

# Parameters

The bep3 module contains the following parameters:

| Key               | Type           | Example                                       | Description                |
| ----------------- | -------------- | --------------------------------------------- | -------------------------- |
| BnbDeputyAddress  | sdk.AccAddress | "kava1r4v2zdhdalfj2ydazallqvrus9fkphmglhn6u6" | deputy's Kava address      |
| BnbDeputyFixedFee | sdkmath.Int    | sdkmath.NewInt(1000)                          | deputy's fixed bnb fee     |
| MinAmount         | sdkmath.Int    | sdkmath.NewInt(0)                             | minimum swap amount        |
| MaxAmount         | sdkmath.Int    | sdkmath.NewInt(1000000000000)                 | maximum swap amount        |
| MinBlockLock      | uint64         | 220                                           | minimum swap expire height |
| MaxBlockLock      | uint64         | 270                                           | maximum swap expire height |
| SupportedAssets   | AssetParams    | []AssetParam                                  | array of supported assets  |

Each AssetParam has the following parameters:

| Key               | Type        | Example             | Description                   |
| ----------------- | ----------- | ------------------- | ----------------------------- |
| AssetParam.Denom  | string      | "bnb"               | asset's name                  |
| AssetParam.CoinID | int64       | 714                 | asset's international coin ID |
| AssetParam.Limit  | sdkmath.Int | sdkmath.NewInt(100) | asset's supply limit          |
| AssetParam.Active | boolean     | true                | asset's state: live or paused |
