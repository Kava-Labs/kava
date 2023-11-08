# Migrate Staking Reward Calculation to Endpoint

Kava 15 (v0.25.x) changes the mechanism for staking rewards, which will no longer be inflationary but will be paid out of the community module. In order to continue displaying expected yields or APYs to users, wallets and explorers will need to update.

The endpoint calculates staking rewards for the current kava version and is forward compatible with future changes.

All consumers who display yearly staking reward percentages are encouraged to migrate from the standard calculation to using the endpoint, as the standard calculation will no longer be accurate.

Endpoint: `/kava/community/v1beta1/annualized_rewards`
Example Response:
```json
{
  "staking_rewards": "0.203023625910000000"
}
```

## Before Kava 15

The staking APR is calculated the same way as other cosmos-sdk chains. Various parameters are fetched and then combined in this calculation:
```
staking_apr ≈ mint_inflation *
    (1 - distribution_params.community_tax) *
    (total_supply_ukava/pool.bonded_tokens)
```

_Note this doesn’t include transaction fees paid to stakers._

Endpoints used:

* https://api.data.kava.io/cosmos/mint/v1beta1/params
* https://api.data.kava.io/cosmos/distribution/v1beta1/params
* https://api.data.kava.io/cosmos/bank/v1beta1/supply/by_denom?denom=ukava
* https://api.data.kava.io/cosmos/staking/v1beta1/pool

Informational Endpoints

* https://api.data.kava.io/cosmos/mint/v1beta1/inflation
* https://api.data.kava.io/cosmos/mint/v1beta1/annual_provisions

## After Kava 15

Kava 15 implements new staking rewards as ratified in this proposal: https://www.mintscan.io/kava/proposals/141. They will come into effect at the “switchover time” on 2024-01-01 00:00 UTC.

* All delegating and claiming transactions remain unchanged. There is no change in how rewards are claimed or how claimable balances are queried.
* After the switchover time, inflation will be set to zero (and rewards will be paid from the community module account).
* After the switchover time, rewards are paid out according to:
```
staking apy ≈ community_params.staking_rewards_per_second *
    seconds_per_year / pool.bonded_tokens
```

_Note this doesn’t include transaction fees paid to stakers._

* There is a new endpoint `kava/community/v1beta1/annualized_rewards`
  * before the switchover time, it will return the current staking APY (calculated in the previous section)
  * after the switchover time, it will return the new staking APY above

* Existing endpoints above will remain active, but the params will change such that the old apr calculation will return 0.

  * https://api.data.kava.io/cosmos/mint/v1beta1/params
    *  no format changes
    *  `inflation_max` and `inflation_min` will be 0.0

  * https://api.data.kava.io/cosmos/distribution/v1beta1/params
    * no format changes
    * `community_tax` will be 0.0

  * https://api.data.kava.io/cosmos/bank/v1beta1/supply/by_denom?denom=ukava
    * no changes

  * https://api.data.kava.io/cosmos/staking/v1beta1/pool
    * no changes

  * https://api.data.kava.io/cosmos/mint/v1beta1/inflation
    * no format changes
    * `inflation` will be 0.0

  * https://api.data.kava.io/cosmos/mint/v1beta1/annual_provisions
    * no format changes
    * `annual_provisions` will be 0.0