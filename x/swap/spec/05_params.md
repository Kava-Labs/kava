<!--
order: 5
-->

# Parameters

Example parameters for the swap module:

| Key     | Type          | Example       | Description                             |
| ------- | ------------- | ------------- | --------------------------------------- |
| Pairs   | array (Pairs) | [{see below}] | Array of tradable pairs supported       |
| SwapFee | sdk.Dec       | 0.03          | Global trading fee in percentage format |

Example parameters for `Pair`:

| Key    | Type   | Example | Description         |
| ------ | ------ | ------- | ------------------- |
| TokenA | string | "ukava" | First coin's denom  |
| TokenB | string | "usdx"  | Second coin's denom |
