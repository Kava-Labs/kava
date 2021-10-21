<!--
order: 5
-->

# Parameters

The auction module contains the following parameters:

| Key                 | Type                   | Example                | Description                                                                           |
|---------------------|------------------------|------------------------|---------------------------------------------------------------------------------------|
| MaxAuctionDuration  | string (time.Duration) | "48h0m0s"              |                                                                                       |
| BidDuration         | string (time.Duration) | "3h0m0s"               |                                                                                       |
| IncrementSurplus    | string (dec)           | "0.050000000000000000" | percentage change in bid required for a new bid on a surplus auction                  |
| IncrementDebt       | string (dec)           | "0.050000000000000000" | percentage change in lot required for a new bid on a debt auction                     |
| IncrementCollateral | string (dec)           | "0.050000000000000000" | percentage change in either bid or lot required for a new bid on a collateral auction |
