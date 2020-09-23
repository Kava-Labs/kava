<!--
order: 3
-->

# Messages

Users claim rewards using a `MsgClaimReward`.

```go
// MsgClaimReward message type used to claim rewards
type MsgClaimReward struct {
  Sender sdk.AccAddress `json:"sender" yaml:"sender"`
  CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
  MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}
```

## State Modifications

* Accumulated rewards for active claims are transferred from the `kavadist` module account to the users account as vesting coins
* The number of coins transferred is determined by the multiplier in the message. For example, the multiplier equals 1.0, 100% of the claim's reward value is transferred. If the multiplier equals 0.5, 50% of the claim's reward value is transferred.
* The corresponding claim object(s) are deleted from the store
