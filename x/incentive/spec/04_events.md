<!--
order: 4
-->

# Events

The `x/incentive` module emits the following events:

## MsgClaimReward

| Type                 | Attribute Key       | Attribute Value    |
|----------------------|---------------------|--------------------|
| claim_reward         | claimed_by          | {claiming address} |
| claim_reward         | claim_amount        | {amount claimed}   |
| message              | module              | incentive          |
| message              | sender              | {sender address}   |

## BeginBlock

| Type                 | Attribute Key       | Attribute Value    |
|----------------------|---------------------|--------------------|
| new_claim_period     | claim_period        | {claim period}     |
| new_reward_period    | reward_period       | {reward period}    |
| claim_period_expiry  | claim_period        | {claim period}     |
