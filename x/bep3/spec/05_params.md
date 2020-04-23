# Parameters

The bep3 module contains the following parameters:

| Key               | Type                    | Example                                       | Description                   |
|-------------------|-------------------------|-----------------------------------------------|-------------------------------|
| BnbDeputyAddress  | string (sdk.AccAddress) | "kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj" | deputy's Kava address         |
| MinBlockLock      | int64                   | 80                                            | minimum swap expire height    |
| MaxBlockLock      | int64                   | 600                                           | maximum swap expire height    |
| SupportedAssets   | AssetParams             | []AssetParam                                  | array of supported assets     |
|-------------------|-------------------------|-----------------------------------------------|-------------------------------|
| AssetParam        | AssetParam              | AssetParam{"bnb", 714, sdk.NewInt(100), true} | a supported asset             |
| AssetParam.Denom  | string                  | "bnb"                                         | asset's name                  |
| AssetParam.CoinID | int64                   | 714                                           | asset's international coin ID |
| AssetParam.Limit  | sdk.Int                 | sdk.NewInt(100)                               | asset's supply limit          |
| AssetParam.Active | boolean                 | true                                          | asset's state: live or paused |
