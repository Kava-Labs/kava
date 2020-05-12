# Parameters

The pricefeed module has the following parameters:

| Key        | Type           | Example       | Description                                      |
|------------|----------------|---------------|--------------------------------------------------|
| Markets    | array (Market) | [{see below}] | array of params for each market in the pricefeed |

Each `Market` has the following parameters

| Key        | Type               | Example                  | Description                                                    |
|------------|--------------------|--------------------------|----------------------------------------------------------------|
| MarketID   | string             | "bnb:usd"                | identifier for the market -- **must** be unique across markets |
| BaseAsset  | string             | "bnb"                    | the base asset for the market pair                             |
| QuoteAsset | string             | "usd"                    | the quote asset for the market pair                            |
| Oracles    | array (AccAddress) | ["kava1...", "kava1..."] | addresses which can post prices for the market                 |
| Active     | bool               | true                     | flag to disable oracle interactions with the module            |
