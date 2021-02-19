<!--
order: 4
-->

# Events

The `x/incentive` module emits the following events:

## ClaimUSDXMintingReward

| Type                 | Attribute Key       | Attribute Value           |
|----------------------|---------------------|---------------------------|
| claim_reward         | claimed_by          | `{claiming address}'      |
| claim_reward         | claim_amount        | `{amount claimed}'        |
| claim_reward         | claim_type          | `{amount claimed}'        |
| message              | module              | incentive                 |
| message              | sender              | hard_liquidity_provider   |

## MsgClaimHardLiquidityProviderReward

| Type                 | Attribute Key       | Attribute Value           |
|----------------------|---------------------|---------------------------|
| claim_reward         | claimed_by          | `{claiming address}'      |
| claim_reward         | claim_amount        | `{amount claimed}'        |
| claim_reward         | claim_type          | `{amount claimed}'        |
| message              | module              | incentive                 |
| message              | sender              | usdx_minting              |
