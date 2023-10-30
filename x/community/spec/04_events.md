<!--
order: 4
-->

# Events

The community module emits the following events:

## Handlers

### MsgFundCommunityPool

| Type    | Attribute Key | Attribute Value     |
| ------- | ------------- | ------------------- |
| message | module        | community           |
| message | action        | fund_community_pool |
| message | sender        | {senderAddress}     |
| message | amount        | {amountCoins}       |

## Keeper events

In addition to handlers events, the bank keeper will produce events when the
following methods are called (or any method which ends up calling them)

### CheckAndDisableMintAndKavaDistInflation

```json
{
  "type": "inflation_stop",
  "attributes": [
    {
      "key": "inflation_disable_time",
      "value": "{{RFC3339 formatted time inflation was disabled}}",
      "index": true
    }
  ]
}
```

### PayoutAccumulatedStakingRewards

```json
{
  "type": "staking_rewards_paid",
  "attributes": [
    {
      "key": "staking_reward_amount",
      "value": "{{sdk.Coins being paid to validators}}",
      "index": true
    }
  ]
}
```
