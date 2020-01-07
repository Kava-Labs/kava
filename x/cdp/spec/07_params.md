# Parameters

The cdp module contains the following parameters:

| Key               | Type                    | Example                             | Description |
|-------------------|-------------------------|-------------------------------------|-------------|
| CollateralParams  | array (CollateralParam) | [{see below}]                       | array of params for each enabled collateral type |
| DebtParams        | array (DebtParam)       | [{see below}]                       | array of params for each enabled pegged asset |
| GlobalDebtLimit   | array (coin)            | [{"denom":"usdx","amount":"1000"}]  | maximum pegged asset that can be minted across the whole system |
| CircuitBreaker    | bool                    | false                               | flag to disable user interactions with the system |
<!-- TODO what is the denom for GlobalDebtLimit?-->

Each CollateralParam has the following parameters:

| Key              | Type          | Example                                     | Description            |
|------------------|---------------|---------------------------------------------|------------------------|
| Denom            | string        | "pbnb"                                      | collateral coin denom  |
| LiquidationRatio | string (dec)  | "1.500000000000000000"                      | the ratio under which a cdp with this collateral type will be liquidated |
| DebtLimit        | array (coin)  | [{"denom":"pbnb","amount":"1000000000000"}] | maximum pegged asset that can be minted backed by this collateral type |
| StabilityFee     | string (dec)  | "0.000000003170000000"                      | per second fee
| Prefix           | number (byte) | 34                                          | identifier used in store keys - **must** be unique across collateral types |
| MarketID         | string        | "BNB/USD"                                   | price feed identifier for this collateral type |
| ConversionFactor | string (int)  | "1000000"                                   | multiplier to go from collateral amount (say $1.50) to internal representation of that amount (1500000)|

Each DebtParam has the following parameters:

| Key              | Type         | Example                                   | Description             |
|------------------|--------------|-------------------------------------------|-------------------------|
| Denom            | string       | "usdx"                                    | pegged asset coin denom |
| ReferenceAsset   | string       | "USD"                                     |  |
| DebtLimit        | array (coin) | [{"denom":"usdx","amount":"10000000000"}] | maximum pegged asset that can be inted of this type |
| ConversionFactor | string (int) | "1000000"                                 |  |
| DebtFloor        | string (int) | "10000000"                                | minimum amount of debt that a CDP can contain |

<!-- TODO add descriptions above -->
