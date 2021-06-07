<!--
order: 5
-->

# Parameters

Example parameters for the swap module:

| Key          | Type                | Example       | Description                             |
| ------------ | ------------------- | ------------- | --------------------------------------- |
| AllowedPools | array (AllowedPool) | [{see below}] | Array of tradable pools supported       |
| SwapFee      | sdk.Dec             | 0.03          | Global trading fee in percentage format |

[Example](Example) parameters for `AllowedPool`:

| Key    | Type   | Example | Description         |
| ------ | ------ | ------- | ------------------- |
| TokenA | string | "ukava" | First coin's denom  |
| TokenB | string | "usdx"  | Second coin's denom |
