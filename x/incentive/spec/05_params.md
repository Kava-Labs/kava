# Parameters

The incentive module contains the following parameters:

| Key        | Type           | Example       | Description                                      |
|------------|----------------|---------------|--------------------------------------------------|
| Active     | bool           | "true"        | boolean for if this module is active             |
| Rewards    | array (Reward) | [{see below}] | array of params for each inflationary period     |

Each `Reward` has the following parameters

| Key              | Type               | Example                            | Description                                                    |
|------------------|--------------------|------------------------------------|----------------------------------------------------------------|
| Active           | bool               | "true                              | boolean for if rewards for this collateral are active          | 
| Denom            | string             | "bnb"                              | the collateral for which rewards are eligible                  |
| AvailableRewards | object (coin)      | `{"denom":"kava","amount":"1000"}` | the rewards available per reward period                        |
| Duration         | string (time ns)   | "172800000000000"                  | the duration of each reward period                             |
| TimeLock         | string (time ns)   | "172800000000000"                  | the duration for which claimed rewards will be vesting         |
| ClaimDuration    | string (time ns)   | "172800000000000"                  | how long users have to claim rewards before they expire        |
