# Parameters

The cdp module contains the following parameters:

| Key              | Type                    | Example                            | Description                                                      |
|------------------|-------------------------|------------------------------------|------------------------------------------------------------------|
| CollateralParams | array (CollateralParam) | [{see below}]                      | array of params for each enabled collateral type                 |
| DebtParams       | array (DebtParam)       | [{see below}]                      | array of params for each enabled pegged asset                    |
| GlobalDebtLimit  | array (coin)            | [{"denom":"usdx","amount":"1000"}] | maximum pegged assets that can be minted across the whole system |
| CircuitBreaker   | bool                    | false                              | flag to disable user interactions with the system                |

Each CollateralParam has the following parameters:

| Key              | Type          | Example                                     | Description                                                                                               |
|------------------|---------------|---------------------------------------------|-----------------------------------------------------------------------------------------------------------|
| Denom            | string        | "pbnb"                                      | collateral coin denom                                                                                     |
| LiquidationRatio | string (dec)  | "1.500000000000000000"                      | the ratio under which a cdp with this collateral type will be liquidated                                  |
| DebtLimit        | array (coin)  | [{"denom":"pbnb","amount":"1000000000000"}] | maximum pegged asset that can be minted backed by this collateral type                                    |
| StabilityFee     | string (dec)  | "0.000000003170000000"                      | per second fee                                                                                            |
| Prefix           | number (byte) | 34                                          | identifier used in store keys - **must** be unique across collateral types                                |
| MarketID         | string        | "BNB/USD"                                   | price feed identifier for this collateral type                                                            |
| ConversionFactor | string (int)  | "1000000"                                   | multiplier to go from external amount (say BTC1.50) to internal representation of that amount (150000000) |

Each DebtParam has the following parameters:

| Key              | Type         | Example                                   | Description                                                                                           |
|------------------|--------------|-------------------------------------------|-------------------------------------------------------------------------------------------------------|
| Denom            | string       | "usdx"                                    | pegged asset coin denom                                                                               |
| ReferenceAsset   | string       | "USD"                                     | asset this asset is pegged to, informational purposes only                                            |
| DebtLimit        | array (coin) | [{"denom":"usdx","amount":"10000000000"}] | maximum pegged asset that can be minted of this type                                                  |
| ConversionFactor | string (int) | "1000000"                                 | multiplier to go from external amount (say $1.50) to internal representation of that amount (1500000) |
| DebtFloor        | string (int) | "10000000"                                | minimum amount of debt that a CDP can contain                                                         |
