<!--
order: 1
-->

# Concepts

## Singular Source of Inflation

The kavamint module mints new tokens each block. How many tokens are minted is determined by rates set in kavamint's parameters. Where the newly minted tokens are transferred to is dictated by which parameter. The kavamint module supports minting for the following purposes:

| Purpose         | Parameter                | Rate Based On           | Transferred To  |
| --------------- | ------------------------ | ----------------------- | --------------- |
| Community Pool  | `CommunityPoolInflation` | TotalSupply             | x/community     |
| Staking Rewards | `StakingRewardsApy`      | TotalSupply*BondedRatio | x/auth fee pool |

`TotalSupply` is the total number of tokens. `BondedRatio` is the percentage of those tokens that are delegated to a bonded validator.

### How Amount is Calculated

The number of seconds since the previous block is determined: `secondsSinceLastBlock`.

For each yearly rate set in the parameters, an equivalent rate for the number of seconds passed is determined assuming the rate compounds every second and, for simplicity, there are a fixed number of seconds per year (`365 x 24 x 60 x 60 = 31,536,000`). The rate compounded over `secondsSinceLastBlock` seconds is the `31,536,000`th root of `rate + 1` to the `secondsSinceLastBlock` power (minus 100% to determine number of new tokens). This number is truncated to an `sdk.Int`.

**example**:
Assume `CommunityPoolInflation` is 20%, the `TotalSupply` is 1e10 tokens and `secondsSinceLastBlock` is 5 seconds. The number of tokens minted is determined to be 289 tokens.

```math
({\sqrt[secondsPerYear]{1 + CommunityPoolInflation}}^{secondsSinceLastBlock} - 1)*TotalSupply
= ({\sqrt[31,536,000]{1.2}}^{5} - 1)*1e10
\approx 289
```

## Staking Reward Distribution

It helps to understand how newly minted tokens are created and distributed as validator rewards by default in a Cosmos SDK blockchain. In the standard Cosmos SDK mint module, tokens are minted each block based on a variable inflation rate determined by the parameters and the desired bonded token ratio. The tokens are then transferred to the auth module's fee pool to be distributed as staking rewards after a portion of the total fee pool is transferred to the community pool (determined by the CommunityTax parameter of the distribution module).

The kavamint module separates the default community pool inflation from staking rewards. For staking rewards, it uses an APY determined by StakingRewardsApy. The basis is the total bonded tokens. The tokens are distributed to the same fee pool as above so they can be distributed as rewards to delegators with the default distribution module.

## Community Pool Inflation

Every block, tokens are distributed to KAVA's community pool, the module account of x/community. The number of tokens distributed is determined by the total supply.

## Disabling Default Inflation

The intention of this module is to replace the default minting of the x/mint module. To make that the case, the minimum and maximum inflation parameters of the mint module should be set to 0. Additionally, it's recommended that the community tax parameter of the distribution module should be set to 0 as the logic for funding the community pool is controlled by kavamint.
