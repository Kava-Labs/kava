<!--
order: 4
-->

# Events

The `x/incentive` module emits the following events:

## ClaimReward

| Type         | Attribute Key | Attribute Value      |
| ------------ | ------------- | -------------------- |
| claim_reward | claimed_by    | `{claiming address}' |
| claim_reward | claim_amount  | `{amount claimed}'   |
| claim_reward | claim_type    | `{amount claimed}'   |
| message      | module        | incentive            |
| message      | sender        | claim_reward         |
