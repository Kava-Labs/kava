<!--
order: 5
-->

# Parameters

The kavadist module has the following parameters:

| Key                  | Type                 | Example       | Description                                              |
| -------------------- | -------------------- | ------------- | -------------------------------------------------------- |
| Active               | bool                 | true          | an all-or-nothing toggle of token minting in this module |
| Periods              | array (Period)       | [{see below}] | array of params for each inflationary period             |
| InfrastructureParams | InfrastructureParams | [{see below}] | object containing infrastructure partner payout params   |

`InfrastructureParams` has the following parameters

| Key                   | Type                  | Example       | Description                                                 |
| --------------------- | --------------------- | ------------- | ----------------------------------------------------------- |
| InfrastructurePeriods | array (Period)        | [{see below}] | array of params for each inflationary period                |
| CoreRewards           | array (CoreReward)    | [{see below}] | array of params for reward weights for core infra providers |
| PartnerRewards        | array (PartnerReward) | [{see below}] | array of params for infrastructure partner reward schedules |

Each `Period` has the following parameters

| Key       | Type      | Example                | Description                             |
| --------- | --------- | ---------------------- | --------------------------------------- |
| Start     | time.Time | "2020-03-01T15:20:00Z" | the time when the period will start     |
| End       | time.Time | "2020-06-01T15:20:00Z" | the time when the period will end       |
| Inflation | sdk.Dec   | "1.000000003022265980" | the per-second inflation for the period |

Each `CoreReward` has the following properties

| Key     | Type           | Example                                       | Description                                              |
| ------- | -------------- | --------------------------------------------- | -------------------------------------------------------- |
| Address | sdk.AccAddress | "kava1x07eng0q9027j7wayap8nvqegpf625uu0w90tq" | address of core infrastructure provider                  |
| Weight  | sdk.Dec        | "0.912345678907654321"                        | % of remaining minted rewards allocated to this provider |

Each `PartnerReward` has the following properties

| Key              | Type           | Example                                       | Description                        |
| ---------------- | -------------- | --------------------------------------------- | ---------------------------------- |
| Address          | sdk.AccAddress | "kava1x0cztstumgcfrw69s5nd5qtu9vdcg7alqtyhgr" | address of infrastructure partner  |
| RewardsPerSecond | object (coin)  | {"denom": "ukava", "amount": "1285" }         | per second reward for this partner |
