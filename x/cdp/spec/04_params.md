<!--
order: 4
-->

# Parameters

The cdp module contains the following parameters:

| Key                          | Type                    | Example                            | Description                                                      |
|------------------------------|-------------------------|------------------------------------|------------------------------------------------------------------|
| CollateralParams             | array (CollateralParam) | [{see below}]                      | array of params for each enabled collateral type                 |
| DebtParams                   | DebtParam               | `{see below}`                      | array of params for each enabled pegged asset                    |
| GlobalDebtLimit              | coin                    | `{"denom":"usdx","amount":"1000"}` | maximum pegged assets that can be minted across the whole system |
| SavingsDistributionFrequency | string (int)            | "84600"                            | number of seconds between distribution of the savings rate       |
| GlobalDebtLimit              | coin                    | `{"denom":"usdx","amount":"1000"}` | maximum pegged assets that can be minted across the whole system |
| DebtAuctionThreshold         | string (int)            | "100000000000"                     | amount of system debt before a debt auction is triggered         |
| SurplusAuctionThreshold      | string (int)            | "100000000000"                     | amount of system surplus before a surplus auction is triggered   |
| DebtAuctionLot               | string (int)            | "10000000000"                      | amount of debt that each debt auction will attempt to recoup     |
| SurplusAuctionLot            | string (int)            | "10000000000"                      | amount of surplus that will be sold at each surplus auction      |

Each CollateralParam has the following parameters:

| Key                 | Type          | Example                                    | Description                                                                     |
|---------------------|---------------|--------------------------------------------|-------------------------------------------------------------------------------|
| Denom               | string        | "bnb"                                      | collateral coin denom                                                         |
| LiquidationRatio    | string (dec)  | "1.500000000000000000"                     | the ratio under which a cdp with this collateral type will be liquidated      |
| DebtLimit           | coin          | `{"denom":"bnb","amount":"1000000000000"}` | maximum pegged asset that can be minted backed by this collateral type        |
| StabilityFee        | string (dec)  | "1.000000001547126"                        | per second fee                                                                |
| Prefix              | number (byte) | "34"                                       | identifier used in store keys - **must** be unique across collateral types    |
| SpotMarketID        | string        | "bnb:usd"                                  | price feed identifier for the spot price of this collateral type              |
| LiquidationMarketID | string        | "bnb:usd:30"                               | price feed identifier for the liquidation price of this collateral type       |
| ConversionFactor    | string (int)  | "6"                                        | 10^_ multiplier for external (BTC1.50) to internal (150000000) representation |

DebtParam has the following parameters:

| Key              | Type         | Example    | Description                                                                                                |
|------------------|--------------|------------|------------------------------------------------------------------------------------------------------------|
| Denom            | string       | "usdx"     | pegged asset coin denom                                                                                    |
| ReferenceAsset   | string       | "USD"      | asset this asset is pegged to, informational purposes only                                                 |
| ConversionFactor | string (int) | "6"        | 10^_ multiplier to go from external amount (say $1.50) to internal representation of that amount (1500000) |
| DebtFloor        | string (int) | "10000000" | minimum amount of debt that a CDP can contain                                                              |
| SavingsRate      | string (dec) | "0.95"     | the percentage of accumulated fees that go towards the savings rate                                        |
